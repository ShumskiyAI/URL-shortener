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
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	st, err := storage.NewStorage(pingCtx)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer st.Close()

	log.Println("connected to Redis")

	svc := service.NewService(st)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("GET /", h.Index)
	mux.HandleFunc("POST /shorten", h.CreateShort)
	mux.HandleFunc("GET /{short}", h.Redirecting)

	middlewareCheck := handler.RecoverMiddleware(handler.LoggingMiddleware(mux))

	server := http.Server{
		Addr:         ":8080",
		Handler:      middlewareCheck,
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
