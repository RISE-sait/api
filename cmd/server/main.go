package main

import (
	routes "api/internal"
	"api/internal/dependencies"
	"api/internal/middlewares"
	"log"
	"net/http"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
)

func main() {

	// Build the connection string
	dependencies := dependencies.InitDependencies()

	router := chi.NewRouter()

	setupMiddlewares(router)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello"))
	})

	// Auth routes
	routes.RegisterRoutes(router, dependencies)

	// Start the server
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func setupMiddlewares(router *chi.Mux) {
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	router.Use(cors.Handler)
	router.Use(middleware.Logger)
	router.Use(middlewares.SetJSONContentType)
}
