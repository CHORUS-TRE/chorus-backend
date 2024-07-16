package middleware

import (
	"context"
	"errors"
	"strings"

	jwt_helper "github.com/CHORUS-TRE/chorus-backend/internal/jwt/helper"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	jwt_go "github.com/golang-jwt/jwt"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	pb "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
)

const (
	endpointContextKey = "endpoint"
	TenantIDContextKey = "ctx_tenant_id"
)

var unaryEndpointInterceptor grpc.UnaryServerInterceptor = func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	grpc_ctxtags.Extract(ctx).Set(endpointContextKey, info.FullMethod)

	return handler(ctx, req)
}

func extractEndpoint(ctx context.Context) (string, error) {
	if tag, found := grpc_ctxtags.Extract(ctx).Values()[endpointContextKey]; found {
		if endpoint, ok := tag.(string); ok {
			return endpoint, nil
		}
	}
	return "", errors.New("unable to extract endpoint from grpc-request")
}

func isPublicEndpoint(endpoint string) (bool, error) {
	parts := strings.Split(endpoint, "/")
	if len(parts) != 3 {
		return false, errors.New("invalid endpoint format")
	}

	serviceName := parts[1]
	methodName := parts[2]

	// Get the service descriptor
	sd, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(serviceName))
	if err != nil {
		return false, err
	}

	// Get the method descriptor
	methods := sd.(protoreflect.ServiceDescriptor).Methods()
	for i := 0; i < methods.Len(); i++ {
		m := methods.Get(i)
		if string(m.Name()) == methodName {
			// Check if the method has the public_endpoint option
			opts := m.Options().(*descriptorpb.MethodOptions)
			if proto.GetExtension(opts, pb.E_PublicEndpoint).(bool) {
				return true, nil
			}
			break
		}
	}

	return false, nil
}

// NewAuthUnaryServerInterceptors returns an second-round JWT-authentication
// intercepor for unary requests for a given key-function and claims-factory.
func NewAuthUnaryServerInterceptors(
	keyFunc jwt_go.Keyfunc,
	claimsFactory jwt_model.ClaimsFactory) []grpc.UnaryServerInterceptor {

	authFunc := func(ctx context.Context) (context.Context, error) {
		// Authenticate requests do not bear a JWT-token.
		endpoint, err := extractEndpoint(ctx)

		if err != nil {
			return nil, err
		}
		isPublic, err := isPublicEndpoint(endpoint)
		if err != nil {
			return nil, err
		}

		if isPublic {
			// return context.WithValue(ctx, "auth", "public"), nil
			return ctx, nil
		}

		tokenString, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		claims, err := jwt_helper.ParseToken(ctx, tokenString, keyFunc, claimsFactory)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authentication token: %v", err)
		}
		tenantID := jwt_helper.TenantIDFromClaims(claims)
		ctx = context.WithValue(ctx, jwt_model.JWTClaimsContextKey, claims)
		ctx = context.WithValue(ctx, logger.TenantIDContextKey{}, tenantID)
		grpc_ctxtags.Extract(ctx).Set("auth.sub", jwt_helper.UserFromClaims(claims))
		grpc_ctxtags.Extract(ctx).Set(TenantIDContextKey, tenantID)

		return ctx, nil
	}

	return []grpc.UnaryServerInterceptor{
		unaryEndpointInterceptor,
		grpc_auth.UnaryServerInterceptor(authFunc),
	}
}

func NewAuthStreamServerInterceptors(
	keyFunc jwt_go.Keyfunc,
	claimsFactory jwt_model.ClaimsFactory) []grpc.StreamServerInterceptor {

	authFunc := func(ctx context.Context) (context.Context, error) {
		tokenString, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		claims, err := jwt_helper.ParseToken(ctx, tokenString, keyFunc, claimsFactory)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}

		grpc_ctxtags.Extract(ctx).Set("auth.sub", jwt_helper.UserFromClaims(claims))
		newCtx := context.WithValue(ctx, jwt_model.JWTClaimsContextKey, claims)
		return newCtx, nil
	}

	return []grpc.StreamServerInterceptor{
		// XXX: streamEndpointInterceptor,
		grpc_auth.StreamServerInterceptor(authFunc),
	}
}
