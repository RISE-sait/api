package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api/cmd/server/router"
	"api/internal/di"
	healthHandler "api/internal/domains/health/handler"

	"github.com/go-chi/cors"

	"api/internal/middlewares"

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

// setupServer initializes and configures the HTTP server router.
//
// It creates a new Chi router, sets up middleware, registers the endpoints
// including the root handler and Swagger documentation, and registers application routes.
//
// Parameters:
//   - container: Dependency injection container that holds application services like db connections and Gcp service
//   - swaggerUrl: The URL where Swagger documentation will be served from
//
// Returns:
//   - An http.Handler that can be used with an HTTP server
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

	// Register health check endpoints at root level for load balancers
	setupHealthCheckRoutes(r, container)

	router.RegisterRoutes(r, container)
	return r
}

// setupMiddlewares configures the middleware stack for the Chi router.
// It sets up:
//   - Standard logging of HTTP requests
//   - Panic recovery to prevent application crashes
//   - Automatic JSON content type header for responses
//   - CORS configuration allowing requests from specific origins with
//     support for credentials, authorized methods, and custom headers
//
// Parameters:
//   - router: The Chi router instance to which middleware will be attached
func setupMiddlewares(router *chi.Mux) {
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middlewares.SetJSONContentType)

	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"https://rise-web-461776259687.us-west2.run.app", "http://localhost:3000", "https://www.rise-basketball.com", "https://www.risesportscomplex.com", "https://www.riseup-hoops.com"}, // Added all production domains
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}, // Added PATCH method
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
		Debug:            true,
	}).Handler)
}

// setupHealthCheckRoutes configures health check endpoints for load balancer integration
// These endpoints are available at the root level without authentication:
//   - GET /health - Comprehensive health check with database and external service status
//   - GET /ready - Kubernetes readiness probe
//   - GET /live - Kubernetes liveness probe
func setupHealthCheckRoutes(router *chi.Mux, container *di.Container) {
	h := healthHandler.NewHealthHandler(container)
	
	// Comprehensive health check endpoint
	router.Get("/health", h.HealthCheck)
	
	// Kubernetes probes
	router.Get("/ready", h.ReadinessCheck)
	router.Get("/live", h.LivenessCheck)
	
	// Additional monitoring endpoints
	router.Get("/health/webhook-retries", h.WebhookRetryStats)
	router.Get("/health/security-audit", h.SecurityAudit)
	
	log.Println("Health check endpoints registered:")
	log.Println("  - GET /health - Comprehensive health status")
	log.Println("  - GET /ready - Readiness check for Kubernetes")
	log.Println("  - GET /live - Liveness check for Kubernetes")
	log.Println("  - GET /health/webhook-retries - Webhook retry statistics")
	log.Println("  - GET /health/security-audit - Comprehensive security assessment")
}
