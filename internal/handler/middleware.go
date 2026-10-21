package handler

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("[%s], %s - %v", r.Method, r.URL.Path, duration)
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic %v\n%s", rec, debug.Stack())
				http.Error(w, "error on server, try again later", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
