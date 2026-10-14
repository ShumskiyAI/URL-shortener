package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"url-shortner/internal/models"
	"url-shortner/internal/service"
)

// POST /shorten — принимает {"url": "..."}, возвращает {"short": "abc123"}
func Shorting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// w.Header().Set("Content-Type", "apllication/json")

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
		log.Printf("err unmarshaling: %v\n", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if input.URL == "" {
		log.Println("input.URL is empty")
		http.Error(w, "inputed URL is empty", http.StatusBadRequest)
		return
	}

	models.Mu.Lock()
	defer models.Mu.Unlock()

	shortURL, err := service.MakeShortURL(input.URL)
	if err != nil {
		log.Println("URL is incorrect")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&shortURL)
	if err != nil {
		log.Fatalf("err encode: %v\n", err)
	}
}

// GET /{short} — редиректит (302) на оригинальный URL
func Redirecting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Println("method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	target := strings.TrimPrefix(r.URL.Path, "/")

	shortURL, err := r.URL.Parse(target)
	if err != nil {
		log.Println("URL is incorrect")
		http.Error(w, "URL is incorrect", http.StatusBadRequest)
		return
	}
	if shortURL.String() == "" {
		log.Println("shortURL is empty")
		http.Error(w, "inputed URL is empty", http.StatusBadRequest)
		return
	}

	models.Mu.Lock()
	defer models.Mu.Unlock()
	fullURL := takeFullURL(shortURL.String())
	if fullURL == "" {
		log.Println("fullURL is not found")
		http.Error(w, "needed URL is not found", http.StatusNotFound)
		return
	}

	// http.StatusFound - 302 статус код
	http.Redirect(w, r, fullURL, http.StatusFound)
}

func takeFullURL(shortURL string) string {
	// models.Mu.Lock()
	// defer models.Mu.Unlock()

	for key, value := range models.URLs {
		if value == shortURL {
			return key
		}
	}
	return ""
}
