package main

import (
	_interface "api/cmd/server/interface"
	"api/cmd/server/router"
	"api/configs"
	courseDb "api/internal/domains/course/infra/persistence/sqlc/generated"
	facilityDb "api/internal/domains/facility/infra/persistence/sqlc/generated"
	identityDb "api/internal/domains/identity/authentication/infra/sqlc/generated"
	membershipDb "api/internal/domains/membership/infra/persistence/sqlc/generated"
	membershipPlanDb "api/internal/domains/membership/plans/infra/persistence/sqlc/generated"

	"api/internal/middlewares"
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
)

func main() {

	// Build the connection string
	dbConn := configs.GetDBConnection()
	defer func(dbConn *sql.DB) {
		err := dbConn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(dbConn)

	identityQueries := identityDb.New(dbConn)
	courseQueries := courseDb.New(dbConn)
	membershipQueries := membershipDb.New(dbConn)
	membershipPlanQueries := membershipPlanDb.New(dbConn)
	facilityQueries := facilityDb.New(dbConn)

	// Create the cRouter and apply middlewares first
	cRouter := chi.NewRouter()

	queries := _interface.QueriesType{
		IdentityDb:       identityQueries,
		CoursesDb:        courseQueries,
		MembershipDb:     membershipQueries,
		MembershipPlanDb: membershipPlanQueries,
		FacilityDb:       facilityQueries,
	}

	setupMiddlewares(cRouter)
	router.RegisterRoutes(cRouter, queries)

	// Define routes
	cRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("helloererererererererererern"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
	router.Use(middlewares.SetJSONContentType)
}
