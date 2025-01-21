package dependencies

import (
	"api/internal/services/hubspot"
	db "api/sqlc"
	"database/sql"
)

type Dependencies struct {
	Queries        *db.Queries
	HubSpotService *hubspot.HubSpotService
	DB             *sql.DB
}
