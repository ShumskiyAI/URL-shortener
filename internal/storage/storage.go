package storage

import (
	"context"
	"errors"
	"fmt"
	"url-shortner/internal/apperr"

	"github.com/redis/go-redis/v9"
)

type storage struct {
	rdb *redis.Client
}

func NewStorage(ctx context.Context) (*storage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // нет пароля
		DB:       0,
	})
	if err := rdb.Ping(ctx); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &storage{
		rdb: rdb,
	}, nil
}

func (s *storage) Set(ctx context.Context, long, short string) error {
	_, err := s.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, "link:"+short, long, 0)
		pipe.Set(ctx, "rev:"+long, short, 0)
		return nil
	})
	if err != nil {
		return fmt.Errorf("set URL from Redis: %w", err)
	}

	return nil
}

func (s *storage) GetLong(ctx context.Context, short string) (string, error) {
	long, err := s.rdb.Get(ctx, "link:"+short).Result()
	if errors.Is(err, redis.Nil) {
		return "", apperr.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get URL from Redis: %w", err)
	}

	return long, nil
}

func (s *storage) GetShortByLong(ctx context.Context, long string) (string, error) {
	short, err := s.rdb.Get(ctx, "rev:"+long).Result()
	if errors.Is(err, redis.Nil) {
		return "", apperr.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get URL from Redis: %w", err)
	}

	return short, nil
}

func (s *storage) NextID(ctx context.Context) (int64, error) {
	id, err := s.rdb.Incr(ctx, "url:sequence").Result()
	if err != nil {
		return 0, fmt.Errorf("increment URL sequence: %w", err)
	}

	return id, nil
}

func (s *storage) Close() error {
	return s.rdb.Close()
}
