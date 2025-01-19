package types

import (
	"api/internal/services"
	db "api/sqlc"
	"database/sql"
)

type Dependencies struct {
	Queries        *db.Queries
	HubSpotService *services.HubSpotService
	DB             *sql.DB
}
