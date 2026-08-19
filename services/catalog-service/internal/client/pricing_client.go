package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pricingv1 "github.com/skalgutkar/grpc-catalog-platform/gen/pricing/v1"
)

const requestIDMetadataKey = "x-request-id"

// NewPricingClient dials pricing-service and wraps every outgoing call with
// a unary interceptor that forwards the caller's x-request-id from the
// incoming gRPC/gateway request, so pricing-service's own Logging
// interceptor (see its internal/interceptor package) can correlate both
// services' logs for one end-to-end request.
func NewPricingClient(addr string) (pricingv1.PricingServiceClient, func() error, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(propagateRequestID),
	)
	if err != nil {
		return nil, nil, err
	}
	return pricingv1.NewPricingServiceClient(conn), conn.Close, nil
}

func propagateRequestID(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if requestID := requestIDFromIncomingContext(ctx); requestID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, requestIDMetadataKey, requestID)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func requestIDFromIncomingContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(requestIDMetadataKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
