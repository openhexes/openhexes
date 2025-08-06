package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"connectrpc.com/connect"
	"github.com/exaring/otelpgx"
	pgxzap "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/multitracer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/openhexes/openhexes/api/src/db"
	"go.uber.org/zap"
)

type Postgres struct {
	Host           string `env:"HOST" envDefault:"localhost"`
	Port           int    `env:"PORT" envDefault:"5432"`
	User           string `env:"USER" envDefault:"postgres"`
	Password       string `env:"PASSWORD" envDefault:"postgres"`
	Database       string `env:"DB" envDefault:"postgres"`
	MaxConnections int32  `env:"MAX_CONNECTIONS" envDefault:"1000"`

	ServicePool *pgxpool.Pool
	Pool        *pgxpool.Pool
}

func (cfg *Config) setUpPostgres(ctx context.Context) error {
	var err error

	if cfg.Test.ID != "" {
		cfg.Postgres.Database = cfg.Test.ID
	}

	cfg.Postgres.ServicePool, err = cfg.initPostgresPool(ctx, true)
	if err != nil {
		return fmt.Errorf("initializing service database pool: %w", err)
	}

	cfg.Postgres.Pool, err = cfg.initPostgresPool(ctx, false)
	if err != nil {
		return fmt.Errorf("initializing database pool: %w", err)
	}

	return nil
}

func (cfg *Config) initPostgresPool(ctx context.Context, isService bool) (*pgxpool.Pool, error) {
	dbStr := ""
	if !isService {
		dbStr = cfg.Postgres.Database
	}
	addr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		dbStr,
	)

	dbcfg, err := pgxpool.ParseConfig(addr)
	if err != nil {
		return nil, fmt.Errorf("parsing database connection string: %w", err)
	}

	dbcfg.MaxConns = cfg.Postgres.MaxConnections

	dbcfg.ConnConfig.Tracer = multitracer.New(
		otelpgx.NewTracer(),
		&tracelog.TraceLog{
			Logger:   pgxzap.NewLogger(GetLogger(ctx).WithOptions(zap.AddCallerSkip(1))),
			LogLevel: tracelog.LogLevelInfo,
		},
	)

	pool, err := pgxpool.NewWithConfig(ctx, dbcfg)
	if err != nil {
		return nil, fmt.Errorf("initializing database connection pool: %w", err)
	}
	return pool, nil
}

func (cfg *Postgres) DropTemporaryDatabase(ctx context.Context) error {
	log := GetLogger(ctx)
	log.Warn("dropping temporary database", zap.String("db", cfg.Database))

	conn, err := cfg.ServicePool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE \"%s\"", cfg.Database))
	if err != nil {
		return fmt.Errorf("dropping database: %w", err)
	}
	return nil
}

func (cfg *Postgres) CreateTemporaryDatabase(ctx context.Context) error {
	log := GetLogger(ctx)
	log.Info("creating temporary database", zap.String("db", cfg.Database))

	conn, err := cfg.ServicePool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.Database))
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	return nil
}

func (cfg *Postgres) ApplyDatabaseMigrations(ctx context.Context) error {
	log := GetLogger(ctx)

	root, err := LocateAppRoot()
	if err != nil {
		return fmt.Errorf("locating app root: %w", err)
	}

	cmd := exec.CommandContext(
		ctx,
		"atlas", "migrate", "apply",
		"--dir", "file://sqlc/migrations",
		"-u", fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		),
	)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Info("applying migrations from %q", zap.String("path", root))
	start := time.Now()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running atlas: %w", err)
	}
	log.Info("migrations applied", zap.Duration("duration", time.Since(start)))
	return nil
}

type TxOption func(*pgx.TxOptions)

func WithIsolationLevel(level pgx.TxIsoLevel) TxOption {
	return func(options *pgx.TxOptions) {
		options.IsoLevel = level
	}
}

func (cfg *Postgres) Tx(ctx context.Context, fn func(tx pgx.Tx, q *db.Queries) error, opts ...TxOption) error {
	log := GetLogger(ctx)

	conn, err := cfg.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	options := &pgx.TxOptions{}
	for _, opt := range opts {
		opt(options)
	}

	tx, err := cfg.Pool.BeginTx(ctx, *options)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("starting transaction: %w", err))
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Warn("rolling back transaction", zap.Error(err))
		}
	}()

	err = fn(tx, db.New(conn))
	if errors.Is(err, pgx.ErrNoRows) {
		return connect.NewError(connect.CodeNotFound, err)
	}
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return connectErr // already wrapped
	} else if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("running query: %w", err))
	}

	if err := tx.Commit(ctx); err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("committing transaction: %w", err))
	}
	return nil
}
