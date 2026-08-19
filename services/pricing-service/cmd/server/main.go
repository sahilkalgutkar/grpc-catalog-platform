package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pricingv1 "github.com/skalgutkar/grpc-catalog-platform/gen/pricing/v1"
	"github.com/skalgutkar/grpc-catalog-platform/services/pricing-service/internal/config"
	"github.com/skalgutkar/grpc-catalog-platform/services/pricing-service/internal/interceptor"
	"github.com/skalgutkar/grpc-catalog-platform/services/pricing-service/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery(logger),
			interceptor.Logging(logger),
		),
	)

	pricingv1.RegisterPricingServiceServer(grpcServer, server.NewPricingServer())

	healthServer := health.NewServer()
	healthServer.SetServingStatus("pricing.v1.PricingService", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Reflection makes grpcurl/grpcui usable against this service without
	// shipping the .proto separately — handy for local debugging and for
	// demonstrating the API in an interview without a client on hand.
	reflection.Register(grpcServer)

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stopCh
		logger.Info("shutting down")
		grpcServer.GracefulStop()
	}()

	logger.Info("pricing-service grpc listening", "port", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("grpc server error", "error", err)
		os.Exit(1)
	}
}
