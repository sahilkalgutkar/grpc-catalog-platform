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

	catalogv1 "github.com/skalgutkar/grpc-catalog-platform/gen/catalog/v1"
	pricingv1 "github.com/skalgutkar/grpc-catalog-platform/gen/pricing/v1"
	"github.com/skalgutkar/grpc-catalog-platform/services/catalog-service/internal/domain"
	"github.com/skalgutkar/grpc-catalog-platform/services/catalog-service/internal/server"
)

// fakePricingServer is a test double for pricing-service so catalog_server
// tests don't need a real pricing-service process — same bufconn approach
// pricing-service's own tests use, applied one layer up.
type fakePricingServer struct {
	pricingv1.UnimplementedPricingServiceServer
	resp *pricingv1.GetPriceResponse
	err  error
}

func (f *fakePricingServer) GetPrice(context.Context, *pricingv1.GetPriceRequest) (*pricingv1.GetPriceResponse, error) {
	return f.resp, f.err
}

func bufDial(t *testing.T, register func(*grpc.Server)) *grpc.ClientConn {
	t.Helper()

	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	grpcServer := grpc.NewServer()
	register(grpcServer)

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

	return conn
}

func newTestClient(t *testing.T, pricing pricingv1.PricingServiceServer) catalogv1.ProductServiceClient {
	t.Helper()

	pricingConn := bufDial(t, func(s *grpc.Server) {
		pricingv1.RegisterPricingServiceServer(s, pricing)
	})
	pricingClient := pricingv1.NewPricingServiceClient(pricingConn)

	catalog := domain.NewCatalog([]domain.Product{
		{ID: "p1", Name: "Widget", UnitPriceCents: 1000},
	})

	catalogConn := bufDial(t, func(s *grpc.Server) {
		catalogv1.RegisterProductServiceServer(s, server.NewCatalogServer(catalog, pricingClient))
	})

	return catalogv1.NewProductServiceClient(catalogConn)
}

func TestGetProduct_WithoutQuantity_SkipsPricingCall(t *testing.T) {
	t.Parallel()

	// A pricing server that errors on any call — if GetProduct reached it
	// despite quantity being unset, this test would fail.
	client := newTestClient(t, &fakePricingServer{err: status.Error(codes.Internal, "should not be called")})

	resp, err := client.GetProduct(context.Background(), &catalogv1.GetProductRequest{Id: "p1"})

	require.NoError(t, err)
	require.Equal(t, "p1", resp.GetId())
	require.Equal(t, "Widget", resp.GetName())
	require.Equal(t, int64(1000), resp.GetUnitPriceCents())
	require.Equal(t, int64(0), resp.GetTotalPriceCents())
}

func TestGetProduct_WithQuantity_CallsPricing(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, &fakePricingServer{resp: &pricingv1.GetPriceResponse{
		ProductId:       "p1",
		Quantity:        50,
		DiscountPercent: 10,
		TotalPriceCents: 45000,
	}})

	resp, err := client.GetProduct(context.Background(), &catalogv1.GetProductRequest{Id: "p1", Quantity: 50})

	require.NoError(t, err)
	require.Equal(t, int32(50), resp.GetQuantity())
	require.Equal(t, int32(10), resp.GetDiscountPercent())
	require.Equal(t, int64(45000), resp.GetTotalPriceCents())
}

func TestGetProduct_NotFound(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, &fakePricingServer{})

	_, err := client.GetProduct(context.Background(), &catalogv1.GetProductRequest{Id: "does-not-exist"})

	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetProduct_MissingID(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, &fakePricingServer{})

	_, err := client.GetProduct(context.Background(), &catalogv1.GetProductRequest{})

	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetProduct_PricingServiceError(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, &fakePricingServer{err: status.Error(codes.Unavailable, "pricing-service down")})

	_, err := client.GetProduct(context.Background(), &catalogv1.GetProductRequest{Id: "p1", Quantity: 5})

	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestListProducts(t *testing.T) {
	t.Parallel()

	pricingConn := bufDial(t, func(s *grpc.Server) {
		pricingv1.RegisterPricingServiceServer(s, &fakePricingServer{})
	})
	pricingClient := pricingv1.NewPricingServiceClient(pricingConn)

	catalog := domain.NewCatalog([]domain.Product{
		{ID: "p1", Name: "Widget", UnitPriceCents: 1000},
		{ID: "p2", Name: "Gadget", UnitPriceCents: 2500},
	})
	catalogConn := bufDial(t, func(s *grpc.Server) {
		catalogv1.RegisterProductServiceServer(s, server.NewCatalogServer(catalog, pricingClient))
	})
	client := catalogv1.NewProductServiceClient(catalogConn)

	resp, err := client.ListProducts(context.Background(), &catalogv1.ListProductsRequest{PageSize: 1})

	require.NoError(t, err)
	require.Len(t, resp.GetProducts(), 1)
	require.Equal(t, "p1", resp.GetProducts()[0].GetId())
}
