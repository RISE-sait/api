package main

import (
	"api/cmd/server/router"
	"api/internal/di"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/cors"
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

// @title			Rise API
// @version		1.0
//
//	@contact.email	klintlee1@gmail.com
//
// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
func main() {

	swaggerUrl := os.Getenv("SWAGGER_URL")
	if swaggerUrl == "" {
		swaggerUrl = "http://localhost/swagger/doc.json"
	}

	diContainer := di.NewContainer()
	defer diContainer.Cleanup()

	server := &http.Server{
		Addr:         ":80",
		Handler:      setupServer(diContainer, swaggerUrl),
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

func setupServer(container *di.Container, swaggerUrl string) http.Handler {
	r := chi.NewRouter()
	setupMiddlewares(r)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello, welcome to Rise API",
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(swaggerUrl), // Use the dynamic host
	))

	router.RegisterRoutes(r, container)
	return r
}

func setupMiddlewares(router *chi.Mux) {
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middlewares.SetJSONContentType)

	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"https://rise-web-461776259687.us-west2.run.app", "*"}, // Allow this specific origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},             // Allowed HTTP methods
		AllowedHeaders:   []string{"Content-Type", "Authorization"},                       // Allowed headers
		ExposedHeaders:   []string{"Authorization"},                                       // Add this line
		AllowCredentials: true,                                                            // Allow cookies and credentials
		Debug:            true,                                                            // Enable CORS debugging
	}).Handler)
}
