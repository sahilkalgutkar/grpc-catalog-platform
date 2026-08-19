package server_test

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pricingv1 "github.com/skalgutkar/grpc-catalog-platform/gen/pricing/v1"
	"github.com/skalgutkar/grpc-catalog-platform/services/pricing-service/internal/server"
)

// newTestClient spins up PricingServer on an in-memory bufconn listener
// and returns a real gRPC client connected to it — no network socket, no
// Docker, no localstack. This is the standard pattern for testing a Go
// gRPC service end-to-end (transport + interceptors + handler) without
// paying for a real network stack.
func newTestClient(t *testing.T) pricingv1.PricingServiceClient {
	t.Helper()

	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	grpcServer := grpc.NewServer()
	pricingv1.RegisterPricingServiceServer(grpcServer, server.NewPricingServer())

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return pricingv1.NewPricingServiceClient(conn)
}

func TestGetPrice_Success(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	resp, err := client.GetPrice(context.Background(), &pricingv1.GetPriceRequest{
		ProductId:      "prod-1",
		UnitPriceCents: 1000,
		Quantity:       50,
	})

	require.NoError(t, err)
	require.Equal(t, "prod-1", resp.GetProductId())
	require.Equal(t, int32(10), resp.GetDiscountPercent())
	require.Equal(t, int64(45000), resp.GetTotalPriceCents())
}

func TestGetPrice_MissingProductID(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	_, err := client.GetPrice(context.Background(), &pricingv1.GetPriceRequest{
		UnitPriceCents: 1000,
		Quantity:       1,
	})

	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetPrice_InvalidQuantity(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	_, err := client.GetPrice(context.Background(), &pricingv1.GetPriceRequest{
		ProductId:      "prod-1",
		UnitPriceCents: 1000,
		Quantity:       0,
	})

	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
