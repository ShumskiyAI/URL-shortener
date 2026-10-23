package service

import (
	"context"
	"errors"
	"testing"

	"url-shortner/internal/apperr"
)

var (
	redisDown = errors.New("redis down")
)

type mockStore struct {
	SetFunc            func(ctx context.Context, long, short string) error
	GetLongFunc        func(ctx context.Context, short string) (string, error)
	GetShortByLongFunc func(ctx context.Context, long string) (string, error)
	NextIDFunc         func(ctx context.Context) (int64, error)
}

func (m *mockStore) Set(ctx context.Context, long, short string) error {
	if m.SetFunc == nil {
		return nil
	}
	return m.SetFunc(ctx, long, short)
}

func (m *mockStore) GetLong(ctx context.Context, short string) (string, error) {
	if m.GetLongFunc == nil {
		return "", apperr.ErrNotFound
	}
	return m.GetLongFunc(ctx, short)
}

func (m *mockStore) GetShortByLong(ctx context.Context, long string) (string, error) {
	if m.GetShortByLongFunc == nil {
		return "", apperr.ErrNotFound
	}
	return m.GetShortByLongFunc(ctx, long)
}

func (m *mockStore) NextID(ctx context.Context) (int64, error) {
	if m.NextIDFunc == nil {
		return 1, nil
	}
	return m.NextIDFunc(ctx)
}

func TestShorten(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		mockStore *mockStore
		wantErr   error
		wantShort string
	}{
		{
			name:     "new valid URL",
			inputURL: "https://example.com",
			mockStore: &mockStore{
				GetShortByLongFunc: func(ctx context.Context, long string) (string, error) {
					return "", apperr.ErrNotFound
				},
				NextIDFunc: func(ctx context.Context) (int64, error) { return 1, nil },
				SetFunc:    func(ctx context.Context, long, short string) error { return nil },
			},
			wantErr:   nil,
			wantShort: "n",
		},
		{
			name:     "existing URL returns same short",
			inputURL: "https://example.com",
			mockStore: &mockStore{
				GetShortByLongFunc: func(ctx context.Context, long string) (string, error) {
					return "abc123", nil
				},
			},
			wantErr:   nil,
			wantShort: "abc123",
		},
		{
			name:      "invalid URL",
			inputURL:  "not a url",
			mockStore: &mockStore{},
			wantErr:   apperr.ErrInvalidURL,
			wantShort: "",
		},
		{
			name:      "empty URL",
			inputURL:  "",
			mockStore: &mockStore{},
			wantErr:   apperr.ErrInvalidURL,
			wantShort: "",
		},
		{
			name:     "store error on GetShortByLong",
			inputURL: "https://example.com",
			mockStore: &mockStore{
				GetShortByLongFunc: func(ctx context.Context, long string) (string, error) {
					return "", redisDown
				},
			},
			wantErr:   redisDown,
			wantShort: "",
		},
		{
			name:     "store error on NextID",
			inputURL: "https://example.com",
			mockStore: &mockStore{
				GetShortByLongFunc: func(ctx context.Context, long string) (string, error) {
					return "", apperr.ErrNotFound
				},
				NextIDFunc: func(ctx context.Context) (int64, error) {
					return 0, redisDown
				},
			},
			wantErr:   redisDown,
			wantShort: "",
		},
		{
			name:     "store error on Set",
			inputURL: "https://example.com",
			mockStore: &mockStore{
				GetShortByLongFunc: func(ctx context.Context, long string) (string, error) {
					return "", apperr.ErrNotFound
				},
				NextIDFunc: func(ctx context.Context) (int64, error) { return 42, nil },
				SetFunc: func(ctx context.Context, long, short string) error {
					return redisDown
				},
			},
			wantErr:   redisDown,
			wantShort: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(tc.mockStore)
			short, err := svc.Shorten(context.Background(), tc.inputURL)

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) && err.Error() != tc.wantErr.Error() {
					t.Fatalf("expected error %v, got: %v", tc.wantErr, err)
				}
			}

			if short != tc.wantShort {
				t.Errorf("expected short %q, got %q", tc.wantShort, short)
			}
		})
	}
}

func TestGetLongURL(t *testing.T) {
	tests := []struct {
		name      string
		short     string
		mockStore *mockStore
		wantLong  string
		wantErr   error
	}{
		{
			name:  "found",
			short: "abc123",
			mockStore: &mockStore{
				GetLongFunc: func(ctx context.Context, short string) (string, error) {
					return "https://example.com", nil
				},
			},
			wantLong: "https://example.com",
			wantErr:  nil,
		},
		{
			name:  "not found",
			short: "missing",
			mockStore: &mockStore{
				GetLongFunc: func(ctx context.Context, short string) (string, error) {
					return "", apperr.ErrNotFound
				},
			},
			wantLong: "",
			wantErr:  apperr.ErrNotFound,
		},
		{
			name:  "store error",
			short: "abc123",
			mockStore: &mockStore{
				GetLongFunc: func(ctx context.Context, short string) (string, error) {
					return "", errors.New("redis down")
				},
			},
			wantLong: "",
			wantErr:  errors.New("redis down"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(tc.mockStore)
			long, err := svc.GetLongURL(context.Background(), tc.short)

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
			}

			if long != tc.wantLong {
				t.Errorf("expected long %q, got %q", tc.wantLong, long)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://example.com", false},
		{"valid http", "http://example.com", false},
		{"valid with path", "https://example.com/path?q=1", false},
		{"empty", "", true},
		{"no scheme", "example.com", true},
		{"wrong scheme", "ftp://example.com", true},
		{"no host", "https://", true},
		{"spaces", "not a url", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateURL(tc.url)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateURL(%q) error = %v, wantErr %v", tc.url, err, tc.wantErr)
			}
		})
	}
}

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		id   int64
		want string
	}{
		{0, "0"},
		{1, "n"},
		{61, "3bl"},
		{62, "3bm"},
		{12345, "3idarWH"},
	}

	for _, tc := range tests {
		got := encodeBase62(tc.id)
		if got != tc.want {
			t.Errorf("encodeBase62(%d) = %q, want %q", tc.id, got, tc.want)
		}
	}
}
