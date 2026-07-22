package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, dsn string, log *slog.Logger) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	log.Info("database connected", slog.String("dsn", maskDSN(dsn)))
	return pool, nil
}

func maskDSN(dsn string) string {
	idx := strings.Index(dsn, "://")
	if idx < 0 {
		return "***"
	}
	afterScheme := dsn[idx+3:]
	at := strings.Index(afterScheme, "@")
	if at < 0 {
		return "***"
	}
	return dsn[:idx+3] + "***" + dsn[idx+3+at:]
}
