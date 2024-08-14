package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/protocol/grpc"
	"github.com/CHORUS-TRE/chorus-backend/internal/protocol/rest"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	google_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	health_grpc "google.golang.org/grpc/health/grpc_health_v1"
)

// startCmd represents the command that boots the servers.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start chorus server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

// runServer initializes and runs both the HTTP- and gRPC server. This is a blocking call.
func runServer() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		logger.TechLog.Info(ctx, "Goodbye")
	}()

	cfg := provider.ProvideConfig()
	err := runExportConfig()
	if err != nil {
		return errors.Wrap(err, "unable to export config")
	}

	info := provider.ProvideComponentInfo()
	c := make(chan os.Signal, 1)

	stopLoggers, err := logger.InitLoggers(cfg, logger.WithSignal(c), logger.WithComponentName(info.Name), logger.WithComponentID(info.ComponentID), logger.WithComponentVersion(info.Version))
	if err != nil {
		return fmt.Errorf("failed to initialize loggers: %w", err)
	}

	defer stopLoggers()

	logger.TechLog.Info(ctx, "component starting",
		zap.String("name", info.Name),
		zap.String("version", info.Version),
		zap.String("id", info.ComponentID),
		zap.String("git_commit", info.Commit),
		zap.String("go_version", info.GoVersion))

	httpHostPort := fmt.Sprintf("%s:%s", cfg.Daemon.HTTP.Host, cfg.Daemon.HTTP.Port)
	grpcHostPort := fmt.Sprintf("%s:%s", cfg.Daemon.GRPC.Host, cfg.Daemon.GRPC.Port)

	started := make(chan struct{})

	// 1. Init and serve the HTTP server, but it will return 503 errors until
	// the gRPC server has started.
	handler, mux, opts := rest.InitServer(ctx, cfg, getVersion(), started, provider.ProvideWorkbench().ProxyWorkbench, provider.ProvideKeyFunc(cfg.Daemon.JWT.Secret.PlainText()), provider.ProvideClaimsFactory())

	httpSrv := &http.Server{
		Addr:    httpHostPort,
		Handler: handler,
	}

	defer httpSrv.Close()

	go func() {
		err := httpSrv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.TechLog.Fatal(ctx, "HTTP server stopped unexpectedly", zap.Error(err))
		}
	}()

	serverMetrics := provider.ProvideServerMetrics()

	// 2. Initialize and run gRPC server.
	server, err := grpc.InitServer(provider.ProvideClientWhitelister(), provider.ProvideKeyFunc(cfg.Daemon.JWT.Secret.PlainText()), provider.ProvideClaimsFactory(), serverMetrics, cfg)
	if err != nil {
		return err
	}

	defer server.Stop()

	registerGRPCServices(server)

	// initialize Prometheus metrics
	serverMetrics.InitializeMetrics(server)

	listen, err := net.Listen("tcp", grpcHostPort)
	if err != nil {
		return err
	}

	defer listen.Close()

	go func() {
		err := server.Serve(listen)
		if err != nil {
			logger.TechLog.Error(ctx, "grpc server stopped unexpectedly", zap.Error(err))
		}
	}()

	// 3. Open the HTTP endpoints after the gRPC server has been listening.

	// This works because the HTTP server does not actually serve any requests
	// until the channel is closed.
	registerHTTPEndpoints(ctx, mux, grpcHostPort, opts)

	// All endpoints are ready so the HTTP server can finally start offering the
	// services.
	close(started)

	// provider.InitDaemonJobs()

	// 4. Wait for the signal to stop.
	signal.Notify(c, os.Interrupt)
	<-c
	signal.Stop(c)

	logger.TechLog.Info(ctx, "shutting down gRPC and HTTP-gateway servers")

	// Give a chance to opened connection to end gracefully.
	server.GracefulStop()

	return nil
}

func registerGRPCServices(server *google_grpc.Server) {
	cfg := provider.ProvideConfig()
	if cfg.Services.AuthenticationService.Enabled {
		chorus.RegisterAuthenticationServiceServer(server, provider.ProvideAuthenticationController())
	}
	chorus.RegisterUserServiceServer(server, provider.ProvideUserController())
	chorus.RegisterStewardServiceServer(server, provider.ProvideStewardController())
	chorus.RegisterNotificationServiceServer(server, provider.ProvideNotificationController())
	chorus.RegisterHealthServiceServer(server, provider.ProvideHealthController())
	chorus.RegisterAppServiceServer(server, provider.ProvideAppController())
	chorus.RegisterAppInstanceServiceServer(server, provider.ProvideAppInstanceController())
	chorus.RegisterWorkspaceServiceServer(server, provider.ProvideWorkspaceController())
	chorus.RegisterWorkbenchServiceServer(server, provider.ProvideWorkbenchController())

	// Setup a standard health check service to allow a client to poll the
	// status.
	health_grpc.RegisterHealthServer(server, health.NewServer())
}

func registerHTTPEndpoints(ctx context.Context, mux *runtime.ServeMux, grpcHostPort string, opts []google_grpc.DialOption) {
	cfg := provider.ProvideConfig()
	if cfg.Services.AuthenticationService.Enabled {
		if err := chorus.RegisterAuthenticationServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
			logger.TechLog.Fatal(ctx, "failed to register http authentication handler", logger.WithErrorField(err))
		}
	}
	if err := chorus.RegisterAttachmentServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
		logger.TechLog.Fatal(ctx, "failed to register http attachment handler", logger.WithErrorField(err))
	}
	if err := chorus.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
		logger.TechLog.Fatal(ctx, "failed to register http user handler", logger.WithErrorField(err))
	}
	if err := chorus.RegisterStewardServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
		logger.TechLog.Fatal(ctx, "failed to register http steward service handler", logger.WithErrorField(err))
	}
	if err := chorus.RegisterNotificationServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
		logger.TechLog.Fatal(ctx, "failed to register http notification service handler", logger.WithErrorField(err))
	}
	if err := chorus.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
		logger.TechLog.Fatal(ctx, "failed to register http health service handler", logger.WithErrorField(err))
	}
	if err := chorus.RegisterAppServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
		logger.TechLog.Fatal(ctx, "failed to register http app service handler", logger.WithErrorField(err))
	}
	if err := chorus.RegisterAppInstanceServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
		logger.TechLog.Fatal(ctx, "failed to register http app instance service handler", logger.WithErrorField(err))
	}
	if err := chorus.RegisterWorkspaceServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
		logger.TechLog.Fatal(ctx, "failed to register http workspace service handler", logger.WithErrorField(err))
	}
	if err := chorus.RegisterWorkbenchServiceHandlerFromEndpoint(ctx, mux, grpcHostPort, opts); err != nil {
		logger.TechLog.Fatal(ctx, "failed to register http workspace service handler", logger.WithErrorField(err))
	}
}
