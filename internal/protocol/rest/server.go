package rest

import (
	"context"
	"net/http"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/protocol/rest/middleware"
	jwt_go "github.com/golang-jwt/jwt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// InitServer initializes a HTTP-server and returns an empty request multiplexer
// for a GRPC gateway and a configuration object.
func InitServer(ctx context.Context, cfg config.Config, version string, started <-chan struct{}, pw middleware.ProxyWorkbenchHandler, keyFunc jwt_go.Keyfunc, claimsFactory jwt_model.ClaimsFactory) (http.Handler, *runtime.ServeMux, []grpc.DialOption) {

	mux := runtime.NewServeMux(
		runtime.WithMetadata(middleware.CorrelationIDMetadata),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
		runtime.WithIncomingHeaderMatcher(newHeaderMatcher(cfg)),
		runtime.WithOutgoingHeaderMatcher(newOutgoingHeaderMatcher()),
	)

	handler := middleware.AddLogger(logger.TechLog, mux)
	handler = middleware.AddInstrumenting(handler)
	handler = middleware.AddCorrelationID(handler)
	handler = middleware.AddRoot(handler, version, started)
	handler = middleware.AddMetrics(handler)
	handler = middleware.AddDoc(handler)
	handler = middleware.AddCORS(handler, cfg)
	if cfg.Services.WorkbenchService.StreamProxyEnabled {
		handler = middleware.AddProxyWorkbench(handler, pw, keyFunc, claimsFactory)
	}
	if cfg.Services.AuthenticationService.DevAuthEnabled {
		handler = middleware.AddDevAuth(handler)
	}
	handler = middleware.AddJWTFromCookie(handler)

	//nolint: staticcheck
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.UseCompressor("gzip"),
			grpc.MaxCallRecvMsgSize(cfg.Daemon.HTTP.MaxCallRecvMsgSize),
			grpc.MaxCallSendMsgSize(cfg.Daemon.HTTP.MaxCallSendMsgSize),
		),
	}

	return handler, mux, opts
}

func newHeaderMatcher(cfg config.Config) runtime.HeaderMatcherFunc {
	if cfg.Daemon.HTTP.HeaderClientIP == "" {
		return nil
	}

	return func(name string) (string, bool) {
		return name, strings.EqualFold(name, cfg.Daemon.HTTP.HeaderClientIP)
	}
}

func newOutgoingHeaderMatcher() runtime.HeaderMatcherFunc {
	return func(key string) (string, bool) {
		switch key {
		case "set-cookie":
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}
}
