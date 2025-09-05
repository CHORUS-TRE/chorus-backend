package middleware

import (
	"context"
	"errors"

	choruserrors "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	// Check if the error is a ChorusError
	var chorusErr *choruserrors.ChorusError
	if errors.As(err, &chorusErr) {
		// TODO: log causedBy error
		return nil, chorusErr.ToGRPCStatus().Err()
	}

	if _, ok := status.FromError(err); ok {
		return nil, err // Already a gRPC error
	}

	return nil, choruserrors.NewInternalError("An unexpected error occurred", err).ToGRPCStatus().Err()
}
