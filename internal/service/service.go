package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"url-shortner/internal/apperr"

	"github.com/yihleego/base62"
)

type store interface {
	Set(ctx context.Context, long, short string) error
	GetLong(ctx context.Context, short string) (string, error)
	GetShortByLong(ctx context.Context, long string) (string, error)
	NextID(ctx context.Context) (int64, error)
}

type service struct {
	store store
}

func NewService(store store) *service {
	return &service{
		store: store,
	}
}

func (s *service) Shorten(ctx context.Context, long string) (string, error) {
	err := validateURL(long)
	if err != nil {
		return "", apperr.ErrInvalidURL
	}

	existing, err := s.store.GetShortByLong(ctx, long)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, apperr.ErrNotFound) {
		return "", fmt.Errorf("get existing URL: %w", err)
	}

	id, err := s.store.NextID(ctx)
	if err != nil {
		return "", err
	}

	short := encodeBase62(id)
	err = s.store.Set(ctx, long, short)
	if err != nil {
		return "", fmt.Errorf("save URL: %w", err)
	}

	return short, nil
}

func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https, got: %q", u.Scheme)
	}

	if u.Host == "" {
		return errors.New("URL host is empty")
	}

	if !strings.Contains(u.Host, ".") {
		return errors.New(`URL doesn't contain "."`)
	}
	// if strings.ContainsAny(url.Host, "\t \n") {
	// 	return errors.New("url host contains whitespace")
	// }

	return nil
}

func encodeBase62(id int64) string {
	if id == 0 {
		return "0"
	}

	idStr := strconv.Itoa(int(id))

	short := base62.StdEncoding.EncodeToString([]byte(idStr))

	return short
}

func (s *service) GetLongURL(ctx context.Context, short string) (string, error) {
	long, err := s.store.GetLong(ctx, short)
	if err != nil {
		return "", fmt.Errorf("get URL from DB: %w", err)
	}

	return long, nil
}

// чтобы не пересоздавать один и тот же url несколько раз, а возвращать тот же
func (s *service) GetShortURL(ctx context.Context, long string) (string, error) {
	short, err := s.store.GetShortByLong(ctx, long)
	if err != nil {
		return "", fmt.Errorf("get URL from DB: %w", err)
	}

	return short, nil
}
