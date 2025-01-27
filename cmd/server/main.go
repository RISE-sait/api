package main

import (
	"api/cmd/server/di"
	"api/cmd/server/router"
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api/internal/middlewares"
	"log"
	"net/http"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
)

func main() {

	diContainer := di.NewContainer()
	defer diContainer.Cleanup()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      setupServer(diContainer),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on %s", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func setupServer(container *di.Container) http.Handler {
	r := chi.NewRouter()
	setupMiddlewares(r)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello, welcome to Rise API",
			"version": "1.0.0",
		})
	})

	router.RegisterRoutes(r, container)
	return r
}

func setupMiddlewares(router *chi.Mux) {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	router.Use(corsHandler.Handler)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middlewares.SetJSONContentType)
}
