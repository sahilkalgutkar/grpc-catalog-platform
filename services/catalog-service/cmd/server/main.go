package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	catalogv1 "github.com/skalgutkar/grpc-catalog-platform/gen/catalog/v1"
	"github.com/skalgutkar/grpc-catalog-platform/services/catalog-service/internal/client"
	"github.com/skalgutkar/grpc-catalog-platform/services/catalog-service/internal/config"
	"github.com/skalgutkar/grpc-catalog-platform/services/catalog-service/internal/domain"
	"github.com/skalgutkar/grpc-catalog-platform/services/catalog-service/internal/server"
)

// seedProducts stands in for a real product store (Postgres/DynamoDB —
// see the comment on domain.Catalog for why that's out of scope here).
var seedProducts = []domain.Product{
	{ID: "p1", Name: "Widget", UnitPriceCents: 1000},
	{ID: "p2", Name: "Gadget", UnitPriceCents: 2500},
	{ID: "p3", Name: "Gizmo", UnitPriceCents: 500},
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	pricingClient, closePricing, err := client.NewPricingClient(cfg.PricingServiceAddr)
	if err != nil {
		logger.Error("failed to dial pricing-service", "error", err)
		os.Exit(1)
	}
	defer closePricing()

	catalogServer := server.NewCatalogServer(domain.NewCatalog(seedProducts), pricingClient)

	grpcServer := grpc.NewServer()
	catalogv1.RegisterProductServiceServer(grpcServer, catalogServer)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("catalog.v1.ProductService", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Reflection makes grpcurl/grpcui usable against this service without
	// shipping the .proto separately.
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}

	// The REST gateway is mounted against catalogServer directly rather
	// than dialing back into the gRPC listener above — see the proto's
	// comment on ProductService: one set of business logic, two
	// transports, no self-loopback network hop.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	if err := catalogv1.RegisterProductServiceHandlerServer(ctx, mux, catalogServer); err != nil {
		logger.Error("failed to register gateway", "error", err)
		os.Exit(1)
	}
	httpServer := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: mux}

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stopCh
		logger.Info("shutting down")
		grpcServer.GracefulStop()
		_ = httpServer.Shutdown(context.Background())
	}()

	go func() {
		logger.Info("catalog-service grpc listening", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("grpc server error", "error", err)
		}
	}()

	logger.Info("catalog-service http gateway listening", "port", cfg.HTTPPort)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("http server error", "error", err)
		os.Exit(1)
	}
}
