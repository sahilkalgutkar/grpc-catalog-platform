package server

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogv1 "github.com/skalgutkar/grpc-catalog-platform/gen/catalog/v1"
	pricingv1 "github.com/skalgutkar/grpc-catalog-platform/gen/pricing/v1"
	"github.com/skalgutkar/grpc-catalog-platform/services/catalog-service/internal/domain"
)

// CatalogServer implements catalogv1.ProductServiceServer. GetProduct's
// quantity-based pricing is a separate RPC to pricing-service rather than
// logic duplicated here — the catalog and pricing concerns stay owned by
// their own services, same as the domain.Catalog/domain.Calculate split
// in the sibling packages.
type CatalogServer struct {
	catalogv1.UnimplementedProductServiceServer

	catalog       *domain.Catalog
	pricingClient pricingv1.PricingServiceClient
}

func NewCatalogServer(catalog *domain.Catalog, pricingClient pricingv1.PricingServiceClient) *CatalogServer {
	return &CatalogServer{catalog: catalog, pricingClient: pricingClient}
}

func (s *CatalogServer) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	product, err := s.catalog.Get(req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to look up product")
	}

	resp := &catalogv1.GetProductResponse{
		Id:             product.ID,
		Name:           product.Name,
		UnitPriceCents: product.UnitPriceCents,
	}

	// quantity is optional — GetProduct without it returns the plain unit
	// price and skips the pricing-service call entirely.
	if req.GetQuantity() > 0 {
		priceResp, err := s.pricingClient.GetPrice(ctx, &pricingv1.GetPriceRequest{
			ProductId:      product.ID,
			UnitPriceCents: product.UnitPriceCents,
			Quantity:       req.GetQuantity(),
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to price product: %v", err)
		}
		resp.Quantity = priceResp.GetQuantity()
		resp.TotalPriceCents = priceResp.GetTotalPriceCents()
		resp.DiscountPercent = priceResp.GetDiscountPercent()
	}

	return resp, nil
}

func (s *CatalogServer) ListProducts(_ context.Context, req *catalogv1.ListProductsRequest) (*catalogv1.ListProductsResponse, error) {
	products := s.catalog.List(req.GetPageSize())

	summaries := make([]*catalogv1.ProductSummary, 0, len(products))
	for _, p := range products {
		summaries = append(summaries, &catalogv1.ProductSummary{
			Id:             p.ID,
			Name:           p.Name,
			UnitPriceCents: p.UnitPriceCents,
		})
	}

	return &catalogv1.ListProductsResponse{Products: summaries}, nil
}
