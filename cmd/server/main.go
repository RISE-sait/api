package main

import (
	"api/cmd/server/router"
	"api/configs"
	db "api/internal/domains/identity/authentication/infra/sqlc/generated"
	"github.com/go-chi/cors"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
)

func main() {

	// Build the connection string
	dbConn := configs.GetDBConnection()
	defer dbConn.Close()

	queries := db.New(dbConn)

	// Create the cRouter and apply middlewares first
	cRouter := chi.NewRouter()

	setupMiddlewares(cRouter)
	router.RegisterRoutes(cRouter, queries)

	// Define routes
	cRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("helloererererererererererern"))
	})

	// Start the server
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", cRouter))
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
}
