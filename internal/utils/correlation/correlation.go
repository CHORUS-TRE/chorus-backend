package correlation

// CorrelationIDContextKey is the key under which the correlation is stored in the context.
type CorrelationIDContextKey struct{}

const (

	// CorrelationIDHTTPHeaderKey is the HTTP-header field name.
	CorrelationIDHTTPHeaderKey = "X-correlation-id"

	// CorrelationIDGRPCMetadataKey is the GRPC metadata key.
	CorrelationIDGRPCMetadataKey = "correlation-id"
)
