package storage

import (
	"context"
	"errors"
	"testing"

	"url-shortner/internal/apperr"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTestStorage(t *testing.T) (*storage, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	s := &storage{rdb: rdb}

	t.Cleanup(func() {
		rdb.Close()
		mr.Close()
	})

	return s, mr
}

func TestSetAndGetLong(t *testing.T) {
	s, _ := setupTestStorage(t)
	ctx := context.Background()

	if err := s.Set(ctx, "https://example.com", "abc123"); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	long, err := s.GetLong(ctx, "abc123")
	if err != nil {
		t.Fatalf("GetLong error: %v", err)
	}
	if long != "https://example.com" {
		t.Errorf("expected %q, got %q", "https://example.com", long)
	}
}

func TestGetShortByLong(t *testing.T) {
	s, _ := setupTestStorage(t)
	ctx := context.Background()

	if err := s.Set(ctx, "https://example.com", "abc123"); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	short, err := s.GetShortByLong(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("GetShortByLong error: %v", err)
	}
	if short != "abc123" {
		t.Errorf("expected %q, got %q", "abc123", short)
	}
}

func TestGetLong_NotFound(t *testing.T) {
	s, _ := setupTestStorage(t)

	_, err := s.GetLong(context.Background(), "missing")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestNextID(t *testing.T) {
	s, _ := setupTestStorage(t)
	ctx := context.Background()

	id1, err := s.NextID(ctx)
	if err != nil {
		t.Fatalf("NextID error: %v", err)
	}
	id2, err := s.NextID(ctx)
	if err != nil {
		t.Fatalf("NextID error: %v", err)
	}

	if id1 != 1 {
		t.Errorf("expected first id 1, got %d", id1)
	}
	if id2 != 2 {
		t.Errorf("expected second id 2, got %d", id2)
	}
}
