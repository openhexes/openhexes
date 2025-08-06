package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Config struct {
	Test     Test     `envPrefix:"TEST__"`
	Auth     Auth     `envPrefix:"AUTH__"`
	Postgres Postgres `envPrefix:"POSTGRES__"`
	Server   Server   `envPrefix:"SERVER__"`
	Logging  Logging  `envPrefix:"LOGGING__"`
}

type Option func(*Config)

func WithTestMode() Option {
	return func(cfg *Config) {
		cfg.Test.ID = fmt.Sprintf(
			"%s-%s",
			strings.ReplaceAll(time.Now().Format(time.TimeOnly), ":", "-"),
			uuid.NewString()[24:],
		)
	}
}

func WithRandomServerAddress() Option {
	return func(cfg *Config) {
		cfg.Server.Address = "localhost:0"
	}
}

func New(ctx context.Context, opts ...Option) (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := cfg.setUpLogging(ctx); err != nil {
		return nil, fmt.Errorf("setting up logging: %w", err)
	}
	if err := cfg.setUpPostgres(ctx); err != nil {
		return nil, fmt.Errorf("setting up postgres: %w", err)
	}
	return &cfg, nil
}

func (cfg *Config) SetUp(ctx context.Context) error {
	log := GetLogger(ctx)
	log.Info("setting up", zap.String("test.id", cfg.Test.ID))

	if cfg.Test.ID != "" {
		if err := cfg.Postgres.CreateTemporaryDatabase(ctx); err != nil {
			return fmt.Errorf("creating temporary database: %w", err)
		}
		if err := cfg.Postgres.ApplyDatabaseMigrations(ctx); err != nil {
			return fmt.Errorf("applying database migrations: %w", err)
		}
	}

	log.Info("set up successfully")
	return nil
}

func (cfg *Config) TearDown(ctx context.Context) error {
	log := GetLogger(ctx)
	log.Info("tearing down", zap.String("test.id", cfg.Test.ID))

	cfg.Postgres.Pool.Close()

	if cfg.Test.ID != "" {
		if err := cfg.Postgres.DropTemporaryDatabase(ctx); err != nil {
			return fmt.Errorf("dropping temporary database: %w", err)
		}
	}

	cfg.Postgres.ServicePool.Close()

	log.Info("teared down successfully")
	return nil
}
