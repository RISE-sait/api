package dependencies

import (
	"api/config"
	"api/internal/services/hubspot"
	db "api/sqlc"
	"database/sql"
)

type Dependencies struct {
	Queries        *db.Queries
	HubSpotService *hubspot.HubSpotService
	DB             *sql.DB
}

func InitDependencies() *Dependencies {
	// Database connection
	dbConn := config.GetDBConnection()
	queries := db.New(dbConn)

	// HubSpot service
	hubSpotService := hubspot.GetHubSpotService()

	return &Dependencies{
		DB:             dbConn,
		Queries:        queries,
		HubSpotService: hubSpotService,
	}
}
