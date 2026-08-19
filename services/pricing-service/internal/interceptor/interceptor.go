package interceptor

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const requestIDMetadataKey = "x-request-id"

// Logging returns a unary server interceptor that logs method, duration,
// and outcome for every RPC, and correlates it with the caller's
// x-request-id if one was propagated (see the matching client
// interceptor in catalog-service's internal/client package).
func Logging(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		requestID := requestIDFromContext(ctx)

		resp, err := handler(ctx, req)

		logger.Info("rpc handled",
			"method", info.FullMethod,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestID,
			"code", status.Code(err).String(),
		)
		return resp, err
	}
}

// Recovery returns a unary server interceptor that converts a panic in
// the handler into codes.Internal instead of crashing the process or
// leaking a raw stack trace to the caller.
func Recovery(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in rpc handler", "method", info.FullMethod, "panic", r)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

func requestIDFromContext(ctx context.Context) string {
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
