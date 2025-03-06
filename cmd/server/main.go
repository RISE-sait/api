package main

import (
	"api/cmd/server/router"
	"api/internal/di"
	"api/internal/services/gcp"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api/internal/middlewares"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"

	_ "api/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @host localhost:80

// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
func main() {

	diContainer := di.NewContainer()
	defer diContainer.Cleanup()

	urls, err := gcp.GetFilesInBucket("rise-sports", "haircut")

	if err != nil {
		log.Fatalf("Failed to get files in bucket: %v", err)
	}

	log.Println(urls)

	server := &http.Server{
		Addr:         ":80",
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
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
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

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	router.RegisterRoutes(r, container)
	return r
}

func setupMiddlewares(router *chi.Mux) {
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middlewares.SetJSONContentType)
}
