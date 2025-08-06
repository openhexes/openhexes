package config

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	ContextKey contextKey = "logger"
)

type Logging struct {
	Level int8 `env:"LEVEL" envDefault:"0"`

	logger *zap.Logger
}

func (cfg *Config) setUpLogging(_ context.Context) error {
	var err error
	if cfg.Test.ID == "" {
		cfg.Logging.logger, err = zap.NewProduction()
	} else {
		lcfg := zap.NewDevelopmentConfig()
		lcfg.Level = zap.NewAtomicLevelAt(zapcore.Level(cfg.Logging.Level))
		cfg.Logging.logger, err = lcfg.Build()
	}
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	zap.ReplaceGlobals(cfg.Logging.logger)
	return nil
}

func (cfg *Logging) InjectLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKey, cfg.logger.With(
		zap.String("trace.id", GetTraceID(ctx)),
	))
}

func GetLogger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(ContextKey).(*zap.Logger)
	if !ok {
		return zap.L()
	}
	return logger
}

func GetTraceID(ctx context.Context) string {
	sctx := trace.SpanContextFromContext(ctx)
	if sctx.HasTraceID() {
		return sctx.TraceID().String()
	}
	return ""
}
