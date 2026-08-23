package redis

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dtt4h/go-marketplace/internal/config"
	"github.com/redis/go-redis/v9"
)

func New(cfg config.RedisConfig, log *slog.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	log.Info("redis connected", slog.String("addr", cfg.Addr()))
	return client, nil
}
