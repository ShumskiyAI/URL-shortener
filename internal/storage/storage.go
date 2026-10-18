package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound = errors.New("Not found")
)

type storage struct {
	db *redis.Client
}

func NewStorage() *storage {
	return &storage{
		db: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // нет пароля
			DB:       0,
		}),
	}
}

func (s *storage) Set(ctx context.Context, long, short string) error {
	_, err := s.db.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, "link:"+short, long, 24*time.Hour)
		pipe.Set(ctx, "rev:"+long, short, 24*time.Hour)
		return nil
	})
	if err != nil {
		log.Printf("error set URL in DB: %v", err)
		return fmt.Errorf("error set URL in DB: %w", err)
	}
	return nil
}

func (s *storage) GetLong(ctx context.Context, short string) (string, error) {
	long, err := s.db.Get(ctx, "link:"+short).Result()

	if errors.Is(err, redis.Nil) {
		log.Printf("%s is not found", short)
		return "", ErrNotFound
	}
	if err != nil {
		log.Printf("error get URL from DB: %v", err)
		return "", fmt.Errorf("error get URL from DB: %w", err)
	}
	return long, nil
}

func (s *storage) GetShortByLong(ctx context.Context, long string) (string, error) {
	short, err := s.db.Get(ctx, "rev:"+long).Result()
	if errors.Is(err, redis.Nil) {
		log.Printf("%s is not found", long)
		return "", ErrNotFound
	}
	if err != nil {
		log.Printf("error get URL from DB: %v", err)
		return "", fmt.Errorf("error get URL from DB: %w", err)
	}
	return short, nil
}
