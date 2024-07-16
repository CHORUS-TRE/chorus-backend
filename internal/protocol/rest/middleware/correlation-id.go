package middleware

import (
	"context"
	"net/http"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/correlation"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	"google.golang.org/grpc/metadata"
)

// AddCorrelationID creates a HTTP-handler that extracts the X-Correlation-ID
// field from the request header.
func AddCorrelationID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if correlationID := r.Header.Get(correlation.CorrelationIDHTTPHeaderKey); correlationID != "" {
			ctx = context.WithValue(ctx, correlation.CorrelationIDContextKey{}, correlationID)
		} else {
			ctx = context.WithValue(ctx, correlation.CorrelationIDContextKey{}, uuid.Next())
		}
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CorrelationIDMetadata is a ServeMuxOption that passes the correlation-id to
// the gRPC metadata.
func CorrelationIDMetadata(ctx context.Context, req *http.Request) metadata.MD {
	if correlationID, ok := ctx.Value(correlation.CorrelationIDContextKey{}).(string); ok {
		return metadata.Pairs(correlation.CorrelationIDGRPCMetadataKey, correlationID)
	}

	return metadata.MD{}
}
