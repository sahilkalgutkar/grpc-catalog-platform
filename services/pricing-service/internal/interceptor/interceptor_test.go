package interceptor_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/skalgutkar/grpc-catalog-platform/services/pricing-service/internal/interceptor"
)

func TestLogging_PassesThroughResponseAndLogsRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handlerCalled := false
	handler := func(ctx context.Context, req any) (any, error) {
		handlerCalled = true
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/pricing.v1.PricingService/GetPrice"}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", "req-123"))

	resp, err := interceptor.Logging(logger)(ctx, nil, info, handler)

	require.True(t, handlerCalled)
	require.NoError(t, err)
	require.Equal(t, "response", resp)
	require.Contains(t, buf.String(), "req-123")
	require.Contains(t, buf.String(), "/pricing.v1.PricingService/GetPrice")
}

func TestLogging_NoRequestIDStillHandles(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "not found")
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/x"}

	_, err := interceptor.Logging(logger)(context.Background(), nil, info, handler)

	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))
	require.Contains(t, buf.String(), "NotFound")
}

// TestLogging_IncomingMetadataWithoutRequestIDKey covers the branch
// requestIDFromContext takes when metadata.FromIncomingContext succeeds
// (ok == true) but the x-request-id key was never set — e.g. a caller that
// sends other metadata but no request ID. TestLogging_NoRequestIDStillHandles
// covers the ok == false case (no incoming metadata at all); this test
// covers the other empty-string path, where metadata is present but the key
// lookup comes back empty.
func TestLogging_IncomingMetadataWithoutRequestIDKey(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/x"}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other-key", "other-value"))

	resp, err := interceptor.Logging(logger)(ctx, nil, info, handler)

	require.NoError(t, err)
	require.Equal(t, "response", resp)
	require.Contains(t, buf.String(), `"request_id":""`)
}

func TestRecovery_PassesThroughNormalCall(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/x"}

	resp, err := interceptor.Recovery(logger)(context.Background(), nil, info, handler)

	require.NoError(t, err)
	require.Equal(t, "ok", resp)
}

func TestRecovery_ConvertsPanicToInternalError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := func(ctx context.Context, req any) (any, error) {
		panic("boom")
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/x"}

	resp, err := interceptor.Recovery(logger)(context.Background(), nil, info, handler)

	require.Nil(t, resp)
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
	require.Contains(t, buf.String(), "panic in rpc handler")
}
