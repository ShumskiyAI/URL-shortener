package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"url-shortner/internal/apperr"
)

type mockService struct {
	ShortenFunc    func(ctx context.Context, long string) (string, error)
	GetLongURLFunc func(ctx context.Context, short string) (string, error)
}

func (m *mockService) Shorten(ctx context.Context, long string) (string, error) {
	if m.ShortenFunc == nil {
		return "", nil
	}
	return m.ShortenFunc(ctx, long)
}

func (m *mockService) GetLongURL(ctx context.Context, short string) (string, error) {
	if m.GetLongURLFunc == nil {
		return "", apperr.ErrNotFound
	}
	return m.GetLongURLFunc(ctx, short)
}

func TestShorting(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockSvc    *mockService
		wantStatus int
		wantShort  string
	}{
		{
			name: "success",
			body: `{"url": "https://example.com"}`,
			mockSvc: &mockService{
				ShortenFunc: func(ctx context.Context, long string) (string, error) {
					return "abc123", nil
				},
			},
			wantStatus: http.StatusOK,
			wantShort:  "abc123",
		},
		{
			name:       "invalid JSON",
			body:       `not json`,
			mockSvc:    &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty URL",
			body:       `{"url": ""}`,
			mockSvc:    &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown field",
			body:       `{"url": "https://example.com", "extra": "x"}`,
			mockSvc:    &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid URL from service",
			body: `{"url": "https://example.com"}`,
			mockSvc: &mockService{
				ShortenFunc: func(ctx context.Context, long string) (string, error) {
					return "", apperr.ErrInvalidURL
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal error from service",
			body: `{"url": "https://example.com"}`,
			mockSvc: &mockService{
				ShortenFunc: func(ctx context.Context, long string) (string, error) {
					return "", errors.New("redis down")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(tc.mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()

			h.CreateShort(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantShort != "" {
				var resp struct {
					Short string `json:"short"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Short != tc.wantShort {
					t.Errorf("expected short %q, got %q", tc.wantShort, resp.Short)
				}
			}
		})
	}
}

func TestRedirecting(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		mockSvc      *mockService
		wantStatus   int
		wantLocation string
	}{
		{
			name: "found - redirect",
			path: "/abc123",
			mockSvc: &mockService{
				GetLongURLFunc: func(ctx context.Context, short string) (string, error) {
					return "https://example.com", nil
				},
			},
			wantStatus:   http.StatusFound,
			wantLocation: "https://example.com",
		},
		{
			name: "not found",
			path: "/missing",
			mockSvc: &mockService{
				GetLongURLFunc: func(ctx context.Context, short string) (string, error) {
					return "", apperr.ErrNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "internal error",
			path: "/abc123",
			mockSvc: &mockService{
				GetLongURLFunc: func(ctx context.Context, short string) (string, error) {
					return "", errors.New("redis down")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(tc.mockSvc)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.SetPathValue("short", strings.TrimPrefix(tc.path, "/")) // ← добавить
			rec := httptest.NewRecorder()

			h.Redirecting(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantLocation != "" {
				if got := rec.Header().Get("Location"); got != tc.wantLocation {
					t.Errorf("expected Location %q, got %q", tc.wantLocation, got)
				}
			}
		})
	}
}

func TestRecoverMiddleware(t *testing.T) {
	panicky := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	wrapped := RecoverMiddleware(panicky)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestLoggingMiddleware_CallsNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	wrapped := LoggingMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
}
