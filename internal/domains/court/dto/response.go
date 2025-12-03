package court

import (
	values "api/internal/domains/court/values"
	"github.com/google/uuid"
)

// ResponseDto represents court data returned to clients
// swagger:model CourtResponse
type ResponseDto struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	LocationID   uuid.UUID `json:"location_id"`
	LocationName string    `json:"location_name"`
}

// NewResponse converts domain values to response dto
func NewResponse(v values.ReadValues) ResponseDto {
	return ResponseDto{
		ID:           v.ID,
		Name:         v.Name,
		LocationID:   v.LocationID,
		LocationName: v.LocationName,
	}
}