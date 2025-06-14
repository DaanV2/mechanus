package grpc_interceptors

import (
	"context"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/mechanus/constants"
)

var _ connect.Interceptor = &LoggingInterceptor{}

// LoggingInterceptor sets metadata information in request, response and the logger
type LoggingInterceptor struct{}

// WrapStreamingClient implements connect.Interceptor.
func (m *LoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		logger := logging.From(ctx).With("procedure", s.Procedure)
		ctx = logging.Context(ctx, logger)
		logger.Debug("streaming client")

		return next(ctx, s)
	}
}

// WrapStreamingHandler implements connect.Interceptor.
func (m *LoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		logger := logging.From(ctx).With("procedure", shc.Spec().Procedure)
		ctx = logging.Context(ctx, logger)
		logger.Debug("streaming handler")

		return next(ctx, shc)
	}
}

// WrapUnary implements connect.Interceptor.
func (m *LoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		logger := logging.From(ctx).With("procedure", req.Spec().Procedure)
		ctx = logging.Context(ctx, logger)
		logger.Debug("unary handler")

		resp, err := next(ctx, req)
		if err != nil {
			logger.Error("unary error", "error", err)

			return resp, err
		}
		if resp != nil && resp.Header() != nil {
			resp.Header().Set("Server", constants.SERVICE_NAME)
		}

		return resp, err
	}
}
