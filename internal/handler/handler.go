package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"url-shortner/internal/apperr"
)

type service interface {
	Shorten(ctx context.Context, long string) (string, error)
	GetLongURL(ctx context.Context, short string) (string, error)
}

type handler struct {
	svc service
}

func NewHandler(svc service) *handler {
	return &handler{
		svc: svc,
	}
}

// POST /shorten — принимает {"url": "..."}, возвращает {"short": "abc123"}
func (h *handler) CreateShort(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL string `json:"url"`
	}

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)) //
	dec.DisallowUnknownFields()
	err := dec.Decode(&input)
	if err != nil {
		log.Printf("err decode: %v", err)
		http.Error(w, "bad request, please try again", http.StatusBadRequest)
		return
	}

	if input.URL == "" {
		log.Println("input.URL is empty")
		http.Error(w, "inputed URL is empty", http.StatusBadRequest)
		return
	}

	shortURL, err := h.svc.Shorten(r.Context(), input.URL)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidURL) {
			log.Printf("invalid URL: %v", err)
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}
		log.Printf("error shorten or add URL in storage: %v", err)
		http.Error(w, "error on the server, please try again later", http.StatusInternalServerError)
		return
	}

	var output struct {
		Short string `json:"short"`
	}

	output.Short = shortURL

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&output)
	if err != nil {
		log.Printf("err encode: %v", err)
		return
	}
}

// GET /{short} — редиректит на оригинальный URL
func (h *handler) Redirecting(w http.ResponseWriter, r *http.Request) {
	short := r.PathValue("short")
	if short == "" {
		log.Println("short URL is empty")
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	long, err := h.svc.GetLongURL(r.Context(), short)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			http.Error(w, "needed URL is not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error on the server, please try again later", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, long, http.StatusFound)
}

func (h *handler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		log.Println(`URL path is not "/"`)
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/index.html")
}
