package grpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/mechanus/constants"
)

var _ connect.Interceptor = &MetadataInterceptor{}

// MetadataInterceptor sets metadata information in request, response and the logger
type MetadataInterceptor struct{}

// WrapStreamingClient implements connect.Interceptor.
func (m *MetadataInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		ctx = logging.Context(ctx, logging.From(ctx).With(
			"procedure", s.Procedure,
		))

		return next(ctx, s)
	}
}

// WrapStreamingHandler implements connect.Interceptor.
func (m *MetadataInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		ctx = logging.Context(ctx, logging.From(ctx).With(
			"procedure", shc.Spec().Procedure,
		))

		return next(ctx, shc)
	}
}

// WrapUnary implements connect.Interceptor.
func (m *MetadataInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		ctx = logging.Context(ctx, logging.From(ctx).With(
			"procedure", req.Spec().Procedure,
		))

		resp, err := next(ctx, req)
		if err != nil {
			return resp, err
		}
		if resp != nil && resp.Header() != nil {
			resp.Header().Set("Server", constants.SERVICE_NAME)
		}

		return resp, err
	}
}
