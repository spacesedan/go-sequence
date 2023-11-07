package internal

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
)

func NewRedis(logger *slog.Logger) (*redis.Client, error) {
	var err error
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
        IdleTimeout: 0,
        WriteTimeout: 0,
	})

	res := rdb.Ping(context.Background())
	if res.Err() != nil {
		err = res.Err()
		logger.Error("internal.NewRedis",
			slog.Group("failed to ping redis",
				slog.String("reason", err.Error())))

		return nil, err
	}
	logger.Info("internal.NewRedis", slog.String("status", "connection successful"))

	return rdb, nil
}
