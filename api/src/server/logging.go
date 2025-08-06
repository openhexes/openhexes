package server

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/openhexes/openhexes/api/src/auth"
	"github.com/openhexes/openhexes/api/src/config"
	"go.uber.org/zap"
)

type LoggingInterceptor struct {
	cfg *config.Config
}

func NewLoggingInterceptor(cfg *config.Config) *LoggingInterceptor {
	return &LoggingInterceptor{
		cfg: cfg,
	}
}

func (i *LoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		ctx = i.cfg.Logging.InjectLogger(ctx)
		log := config.GetLogger(ctx)

		start := time.Now()

		spec := request.Spec()
		prefix := "grpc:server"
		if spec.IsClient {
			prefix = "grpc:client"
		}

		account := auth.AccountFromContext(ctx)
		log.Info(
			fmt.Sprintf("%s:request", prefix),
			zap.String("account.id", account.ID.String()),
			zap.String("method", spec.Procedure),
			zap.Uint8("streamType", uint8(spec.StreamType)),
			zap.Int("size", len(request.Header().Get("Content-Length"))),
		)

		response, err := next(ctx, request)
		fields := []zap.Field{
			zap.Duration("duration", time.Since(start)),
		}
		if err != nil {
			fields = append(
				fields,
				zap.Uint32("code", uint32(connect.CodeOf(err))),
				zap.Error(err),
			)
			log.Warn(fmt.Sprintf("%s:error", prefix), fields...)
		} else {
			log.Info(fmt.Sprintf("%s:response", prefix), fields...)
		}
		return response, err
	})
}

func (i *LoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

func (i *LoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx = i.cfg.Logging.InjectLogger(ctx)
		log := config.GetLogger(ctx)

		start := time.Now()
		spec := conn.Spec()
		account := auth.AccountFromContext(ctx)

		log.Info(
			"grpc:server:request",
			zap.String("account.id", account.ID.String()),
			zap.String("method", spec.Procedure),
			zap.Uint8("streamType", uint8(spec.StreamType)),
		)

		err := next(ctx, conn)
		fields := []zap.Field{
			zap.Duration("duration", time.Since(start)),
		}
		if err != nil {
			fields = append(
				fields,
				zap.Uint32("code", uint32(connect.CodeOf(err))),
				zap.Error(err),
			)
			log.Warn("grpc:server:error", fields...)
		} else {
			log.Info("grpc:server:response", fields...)
		}
		return err
	})
}
