package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestPropagateRequestID_ForwardsFromIncomingToOutgoing(t *testing.T) {
	incoming := metadata.NewIncomingContext(context.Background(), metadata.Pairs(requestIDMetadataKey, "req-1"))

	var capturedCtx context.Context
	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		capturedCtx = ctx
		return nil
	}

	err := propagateRequestID(incoming, "/x", nil, nil, nil, invoker)

	require.NoError(t, err)
	md, ok := metadata.FromOutgoingContext(capturedCtx)
	require.True(t, ok)
	require.Equal(t, []string{"req-1"}, md.Get(requestIDMetadataKey))
}

func TestPropagateRequestID_NoIncomingRequestID(t *testing.T) {
	var capturedCtx context.Context
	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		capturedCtx = ctx
		return nil
	}

	err := propagateRequestID(context.Background(), "/x", nil, nil, nil, invoker)

	require.NoError(t, err)
	_, ok := metadata.FromOutgoingContext(capturedCtx)
	require.False(t, ok)
}

// TestPropagateRequestID_IncomingMetadataWithoutRequestIDKey covers the
// branch requestIDFromIncomingContext takes when metadata.FromIncomingContext
// succeeds (ok == true) but the x-request-id key itself was never set — a
// gRPC/gateway request that carries other metadata but no request ID. This
// is distinct from TestPropagateRequestID_NoIncomingRequestID, which never
// reaches the incoming-metadata branch at all because there's no incoming
// context to begin with.
func TestPropagateRequestID_IncomingMetadataWithoutRequestIDKey(t *testing.T) {
	incoming := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other-key", "other-value"))

	var capturedCtx context.Context
	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		capturedCtx = ctx
		return nil
	}

	err := propagateRequestID(incoming, "/x", nil, nil, nil, invoker)

	require.NoError(t, err)
	_, ok := metadata.FromOutgoingContext(capturedCtx)
	require.False(t, ok, "no x-request-id key present means nothing should be forwarded")
}

func TestPropagateRequestID_PropagatesInvokerError(t *testing.T) {
	wantErr := context.DeadlineExceeded
	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return wantErr
	}

	err := propagateRequestID(context.Background(), "/x", nil, nil, nil, invoker)

	require.ErrorIs(t, err, wantErr)
}

// NewClient never blocks on connection establishment (that's the whole
// point of the NewClient/Dial split in modern grpc-go), so this doesn't
// need a real pricing-service listening — it's exercising the dial-option
// wiring (transport credentials, the interceptor), not connectivity.
func TestNewPricingClient_ReturnsUsableClient(t *testing.T) {
	pricingClient, closeFn, err := NewPricingClient("localhost:9091")

	require.NoError(t, err)
	require.NotNil(t, pricingClient)
	require.NotNil(t, closeFn)
	require.NoError(t, closeFn())
}

// TestNewPricingClient_InvalidTargetReturnsError exercises the error branch
// of NewPricingClient: a malformed target (a NUL byte, which trips grpc's
// underlying target-URL parsing) makes grpc.NewClient fail before any
// connection is attempted, and NewPricingClient must surface that error
// with a nil client and nil close func rather than a partially-built one.
func TestNewPricingClient_InvalidTargetReturnsError(t *testing.T) {
	pricingClient, closeFn, err := NewPricingClient("unix:///bad\x00target")

	require.Error(t, err)
	require.Nil(t, pricingClient)
	require.Nil(t, closeFn)
}
