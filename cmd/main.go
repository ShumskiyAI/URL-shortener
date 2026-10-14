package main

import (
	"errors"
	"log"
	"net/http"
	"time"
	"url-shortner/internal/handler"
)

func main() {
	mux := http.NewServeMux()

	// localhost:8080/shorten/https://youtube.com/lens
	// curl.exe -X POST http://localhost:8080/shorten/https://youtube.com/lens
	mux.HandleFunc("POST /shorten", handler.Shorting)

	// curl.exe http://localhost:8080/тут должна быть короткая ссылка
	mux.HandleFunc("GET /{short}", handler.Redirecting)

	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// go func() {
	log.Println("start listening server")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
	// }()

}
