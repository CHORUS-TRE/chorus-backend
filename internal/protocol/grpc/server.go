package grpc

import (
	"context"
	"fmt"

	jwt_go "github.com/golang-jwt/jwt"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/encoding/gzip" // Install the gzip compressor
	"google.golang.org/grpc/status"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/metrics"
	"github.com/CHORUS-TRE/chorus-backend/internal/protocol/grpc/middleware"
)

// InitServer initialize and return gRPC server with middleware interceptors specified in './middleware'.
func InitServer(whitelister middleware.ClientWhitelister, keyFunc jwt_go.Keyfunc, claimsFactory jwt_model.ClaimsFactory, serverMetrics *metrics.ServerMetrics, cfg config.Config) (*grpc.Server, error) {
	// Create unary and stream interceptors.
	var unary []grpc.UnaryServerInterceptor
	var stream []grpc.StreamServerInterceptor

	// Add logging middleware.
	unary = append(unary, middleware.NewLoggingUnaryServerInterceptors(logger.TechLog.Logger)...)
	stream = append(stream, middleware.NewLoggingStreamServerInterceptors(logger.TechLog.Logger)...)

	// Add correlation ID middleware.
	unary = append(unary, middleware.NewCorrelationIDUnaryServerInterceptors())
	stream = append(stream, middleware.NewCorrelationIDStreamServerInterceptors())

	// metrics (prometheus) middleware
	unary = append(unary, metrics.NewMetricsUnaryServerInterceptors(serverMetrics)...)
	stream = append(stream, metrics.NewMetricsStreamServerInterceptors(serverMetrics)...)

	// Add JWT-authentication middleware.
	unary = append(unary, middleware.NewAuthUnaryServerInterceptors(keyFunc, claimsFactory)...)
	stream = append(stream, middleware.NewAuthStreamServerInterceptors(keyFunc, claimsFactory)...)

	// Add validator middleware.
	unary = append(unary, grpc_validator.UnaryServerInterceptor())
	stream = append(stream, grpc_validator.StreamServerInterceptor())

	// Add IP whitelisting middleware.
	unary = append(unary, middleware.NewIPWhitelistUnaryServerInterceptor(whitelister))
	stream = append(stream, middleware.NewIPWhitelistStreamServerInterceptor(whitelister))

	// Add recovery middleware.
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			logger.TechLog.Fatal(context.Background(), "goodbye world, panic occurred", zap.String("panic_error", fmt.Sprintf("%v", p)), zap.Stack("panic_stack_trace"))
			return status.Errorf(codes.Internal, "panic occurred: %v", p)
		}),
	}
	unary = append(unary, grpc_recovery.UnaryServerInterceptor(recoveryOpts...))
	stream = append(stream, grpc_recovery.StreamServerInterceptor(recoveryOpts...))

	// gRPC grpcServer startup options.
	opts := []grpc.ServerOption{}
	opts = append(opts, grpc.ChainUnaryInterceptor(unary...))
	opts = append(opts, grpc.ChainStreamInterceptor(stream...))
	opts = append(opts, grpc.MaxRecvMsgSize(cfg.Daemon.GRPC.MaxRecvMsgSize))
	opts = append(opts, grpc.MaxSendMsgSize(cfg.Daemon.GRPC.MaxSendMsgSize))

	return grpc.NewServer(opts...), nil
}
