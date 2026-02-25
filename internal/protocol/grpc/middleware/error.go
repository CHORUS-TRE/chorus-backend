package middleware

import (
	"context"
	"errors"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	// Check if the error is a ChorusError
	var cErr *cerr.ChorusError
	if errors.As(err, &cErr) {
		logger.TechLog.Error(ctx, "request failed",
			zap.String("method", info.FullMethod),
			zap.String("code", cErr.ChorusCode.String()),
			zap.String("message", cErr.Message),
			zap.Error(cErr.CausedBy),
		)
		return nil, cErr.ToGRPCStatus().Err()
	}

	// If it's already a gRPC status error, return it as is
	if _, ok := status.FromError(err); ok {
		return nil, err // Already a gRPC error
	}

	return nil, cerr.ErrInternal.Wrap(err, "An unexpected error occurred.").ToGRPCStatus().Err()
}
