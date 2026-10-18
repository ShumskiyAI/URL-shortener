package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortner/internal/handler"
	"url-shortner/internal/service"
	"url-shortner/internal/storage"
)

func main() {
	storage := storage.NewStorage()
	service := service.NewService(storage)
	handler := handler.NewHandler(service)

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

	errCh := make(chan error, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("start listening server")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		log.Printf("server error: %v", err)
		os.Exit(1)

	case sig := <-quit:
		log.Printf("received signal %v, shutting down...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
			if closeErr := server.Close(); closeErr != nil {
				log.Printf("close error: %v", closeErr)
			}
		}

		log.Println("server stopped gracefully")
	}
}
