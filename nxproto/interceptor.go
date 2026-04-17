package nxproto

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// LoggingInterceptor creates a Connect RPC unary interceptor that logs
// procedure, request, response (or error), and elapsed time. Sensitive
// proto fields are redacted. Works for both client-side and server-side.
func LoggingInterceptor(logger *zap.Logger) connect.UnaryInterceptorFunc {
	SensitiveIdx() // warm up the index

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := req.Spec().Procedure
			reqField := protoField("request", req.Any())

			start := time.Now()
			resp, err := next(ctx, req)
			elapsed := time.Since(start)

			if err != nil {
				logger.Warn("rpc",
					zap.String("procedure", procedure),
					reqField,
					zap.String("error", err.Error()),
					zap.Duration("elapsed", elapsed),
				)
			} else {
				var respField zap.Field
				if resp != nil {
					respField = protoField("response", resp.Any())
				} else {
					respField = zap.Skip()
				}
				logger.Info("rpc",
					zap.String("procedure", procedure),
					reqField,
					respField,
					zap.Duration("elapsed", elapsed),
				)
			}

			return resp, err
		}
	}
}

func protoField(key string, v any) zap.Field {
	msg, ok := v.(proto.Message)
	if !ok {
		return zap.Skip()
	}
	return ProtoJSON(key, msg)
}
