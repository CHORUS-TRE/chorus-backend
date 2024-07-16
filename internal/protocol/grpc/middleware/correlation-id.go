package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/correlation"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewCorrelationIDUnaryServerInterceptors returns a gRPC-server interceptor that extracts the correlation-ID
// field from the GRPC context.
func NewCorrelationIDUnaryServerInterceptors() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		correlationID := getCorrelationIDFromContextMetadata(ctx)

		ctx = context.WithValue(ctx, correlation.CorrelationIDContextKey{}, correlationID)

		return handler(ctx, req)
	}
}

// NewCorrelationIDStreamServerInterceptors returns a gRPC-server interceptor that extracts the correlation-ID
// field from the GRPC context in stream requests.
func NewCorrelationIDStreamServerInterceptors() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		correlationID := getCorrelationIDFromContextMetadata(ctx)

		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = context.WithValue(ctx, correlation.CorrelationIDContextKey{}, correlationID)

		return handler(srv, wrapped)
	}
}

// getCorrelationIDFromContextMetadata get the correlation ID from the GRPC
// metadata that are in the context (under the key 'correlation-id'). If there
//
//	is no such correlation ID, a new one is generated.
func getCorrelationIDFromContextMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Next()
	}

	values := md.Get(correlation.CorrelationIDGRPCMetadataKey)
	if len(values) == 0 {
		return uuid.Next()
	}

	return values[0]
}
