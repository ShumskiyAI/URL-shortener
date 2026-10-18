package service

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"net/url"
	"strings"
	"url-shortner/internal/storage"

	"github.com/yihleego/base62"
)

type store interface {
	Set(ctx context.Context, long, short string) error
	GetLong(ctx context.Context, short string) (string, error)
	GetShortByLong(ctx context.Context, long string) (string, error)
}

type service struct { // потом сделать маленькую и добавить интерфейс в handler
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
		return "", fmt.Errorf("%w", err)
	}

	short := encodeURL(long)

	_, err = s.GetShortByLong(ctx, long)
	if !errors.Is(err, storage.ErrNotFound) {
		return short, nil
	}

	if err := s.store.Set(ctx, long, short); err != nil {
		log.Println("err in storage.Set")
		return "", fmt.Errorf("error writing URL in DB: %w", err)
	}

	return short, nil
}

func validateURL(long string) error {
	url, err := url.Parse(long)
	if err != nil {
		return fmt.Errorf("URL is incorrect: %w", err)
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https, got: %q", url.Scheme)
	}

	if url.Host == "" {
		return errors.New("url host is empty")
	}

	if !strings.Contains(url.Host, ".") {
		return errors.New(`URL doesn't contain "."`)
	}

	if strings.ContainsAny(url.Host, "\t \n") {
		return errors.New("url host contains whitespace")
	}
	return nil
}

func encodeURL(long string) string {
	h := fnv.New64a()
	h.Write([]byte(long))

	sum := h.Sum64()

	var b [8]byte
	binary.BigEndian.PutUint64(b[:], sum)

	encoded := base62.StdEncoding.EncodeToString(b[:]) // 11 символов
	return encoded[:8]                                 // или [:8], или [:10]
}

func (s *service) GetLong(ctx context.Context, short string) (string, error) {
	long, err := s.store.GetLong(ctx, short)
	if err != nil {
		log.Println("")
		return "", fmt.Errorf("%w", err) // тут нормально обработать
	}

	return long, nil
}

func (s *service) GetShortByLong(ctx context.Context, long string) (string, error) {
	short, err := s.store.GetShortByLong(ctx, long)
	if err != nil {
		log.Println("")
		return "", fmt.Errorf("%w", err) // тут нормально обработать
	}

	return short, nil
}
