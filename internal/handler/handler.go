package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type service interface {
	Shorten(ctx context.Context, long string) (string, error)
	GetLong(ctx context.Context, short string) (string, error)
	GetShortByLong(ctx context.Context, long string) (string, error)
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
func (h *handler) Shorting(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL string `json:"url"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("error of reading URL")
		http.Error(w, "error of reading URL", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Printf("err unmarshaling: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if input.URL == "" {
		log.Println("input.URL is empty")
		http.Error(w, "inputed URL is empty", http.StatusBadRequest)
		return
	}

	shortURL, err := h.svc.Shorten(r.Context(), input.URL)
	if err != nil {
		log.Println("URL is incorrect")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&shortURL)
	if err != nil {
		log.Printf("err encode: %v", err)
		return
	}
}

// GET /{short} — редиректит (302) на оригинальный URL
func (h handler) Redirecting(w http.ResponseWriter, r *http.Request) {
	short := r.PathValue("short")
	if short == "" {
		log.Println("short is empty")
		http.Error(w, "inputed URL is empty", http.StatusBadRequest)
		return
	}

	long, err := h.svc.GetLong(r.Context(), short)
	if err != nil {
		// тут какая-то обработка
		log.Println()
		http.Error(w, "", 400)
		return
	}
	if long == "" {
		log.Println("long is not found")
		http.Error(w, "needed URL is not found", http.StatusNotFound)
		return
	}

	// http.StatusFound - 302 статус код
	http.Redirect(w, r, long, http.StatusFound)
}
