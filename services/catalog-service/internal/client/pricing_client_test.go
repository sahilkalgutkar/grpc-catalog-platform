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
