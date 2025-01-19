package types

import (
	"api/internal/services"
	db "api/sqlc"
)

type Dependencies struct {
	Queries        *db.Queries
	HubSpotService *services.HubSpotService
}
