package middleware

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	headerXForwardedFor = "x-forwarded-for"
)

// ClientWhitelister is an interface to deny the access to clients not in the
// whitelist.
type ClientWhitelister interface {
	Verify(ctx context.Context) error
}

// NewIPWhitelistUnaryServerInterceptor creates a new unary interceptor to ensure IP
// whitelisting for HTTP requests.
func NewIPWhitelistUnaryServerInterceptor(cw ClientWhitelister) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		if err := cw.Verify(ctx); err != nil {
			return nil, status.Errorf(codes.PermissionDenied, "ip whitelist: %v", err)
		}

		return handler(ctx, req)
	}
}

// NewIPWhitelistStreamServerInterceptor creates a new stream interceptor to
// ensure IP whitelisting for HTTP requests.
func NewIPWhitelistStreamServerInterceptor(cw ClientWhitelister) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		ctx := ss.Context()

		if err := cw.Verify(ctx); err != nil {
			return status.Errorf(codes.PermissionDenied, "ip whitelist: %v", err)
		}

		return handler(srv, ss)
	}
}

// IPWhitelister is a client whitelister which is using the client IP from the
// HTTP request headers.
//
// A request coming from the gateway is detected through the presence of the
// X-Forwarded-For header which is forwarded. The daemon offers a configuration
// to use a different HTTP header to get the client IP (e.g. CloudFlare
// True-Client-IP).
//
// While the client IP header configuration is in the scope of the daemon, the
// whitelist is configured by tenant, which is the reason why it is enforced as
// a gRPC interceptor so that authentication is performed before.
type IPWhitelister struct {
	headerKey string
	tables    map[uint64][]*net.IPNet
}

// NewIPWhitelister creates a new IP whitelister.
func NewIPWhitelister(cfg config.Config) (*IPWhitelister, error) {

	tables := make(map[uint64][]*net.IPNet)

	for tenantID, tenant := range cfg.Tenants {
		if !tenant.Enabled || !tenant.IPWhitelist.Enabled {
			continue
		}

		subnets := make([]*net.IPNet, len(tenant.IPWhitelist.Subnetworks))

		for i, str := range tenant.IPWhitelist.Subnetworks {
			_, subnet, err := net.ParseCIDR(str)
			if err != nil {
				return nil, fmt.Errorf("unable to parse cidr '%s': %v", str, err)
			}

			subnets[i] = subnet
		}

		tables[tenantID] = subnets
	}

	headerKey := strings.ToLower(cfg.Daemon.HTTP.HeaderClientIP)
	if headerKey == "" {
		headerKey = headerXForwardedFor
	}

	w := &IPWhitelister{
		tables:    tables,
		headerKey: headerKey,
	}

	return w, nil
}

// Verify looks up the details of the client from the context and returns an
// error when it does not comply with the whitelist.
func (w *IPWhitelister) Verify(ctx context.Context) error {

	tenantID, ok := ctx.Value(logger.TenantIDContextKey{}).(uint64)
	if !ok {
		// Either it's a public endpoint and thus we cannot check the IP as we
		// need the tenant configuration, or an attacker is trying to reach an
		// authenticated endpoint so it will be refused in the auth middleware.
		return nil
	}

	subnets, found := w.tables[tenantID]
	if !found {
		// IP whitelisting is not enabled for this tenant.
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("missing metadata in context")
	}

	values := md.Get(w.headerKey)
	if len(values) == 0 {

		// When the header is absent, it checks first that it does not deal with
		// a gateway request.
		if w.headerKey != headerXForwardedFor && len(md.Get(headerXForwardedFor)) > 0 {
			return errors.New("missing client ip header in metadata")
		}

		// Not a request from the gateway.
		return nil
	}

	ip := net.ParseIP(strings.Split(values[0], ",")[0])

	for _, subnet := range subnets {
		if subnet.Contains(ip) {
			return nil
		}
	}

	logger.TechLog.Warn(ctx, "client has been refused access by the IP whitelist", zap.Stringer("client", ip), zap.String("header", w.headerKey))

	return errors.New("client is not allowed")
}
