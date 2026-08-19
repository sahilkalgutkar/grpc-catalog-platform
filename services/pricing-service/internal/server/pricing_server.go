package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pricingv1 "github.com/skalgutkar/grpc-catalog-platform/gen/pricing/v1"
	"github.com/skalgutkar/grpc-catalog-platform/services/pricing-service/internal/domain"
)

// PricingServer implements pricingv1.PricingServiceServer. It's a thin
// adapter: validate the wire request, call the pure domain function,
// translate the result (or error) back to protobuf/gRPC status codes.
type PricingServer struct {
	pricingv1.UnimplementedPricingServiceServer
}

func NewPricingServer() *PricingServer {
	return &PricingServer{}
}

func (s *PricingServer) GetPrice(_ context.Context, req *pricingv1.GetPriceRequest) (*pricingv1.GetPriceResponse, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	total, discount, err := domain.Calculate(req.GetUnitPriceCents(), req.GetQuantity())
	switch {
	case err == nil:
		// fall through
	case err == domain.ErrInvalidUnitPrice, err == domain.ErrInvalidQuantity:
		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		return nil, status.Error(codes.Internal, "failed to calculate price")
	}

	return &pricingv1.GetPriceResponse{
		ProductId:       req.GetProductId(),
		Quantity:        req.GetQuantity(),
		DiscountPercent: discount,
		TotalPriceCents: total,
	}, nil
}
