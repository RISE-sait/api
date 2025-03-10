package location

import (
	values "api/internal/domains/location/values"
	"github.com/google/uuid"
)

type ResponseDto struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
}

func NewLocationResponse(facility values.ReadValues) ResponseDto {
	return ResponseDto{
		ID:      facility.ID,
		Name:    facility.Name,
		Address: facility.Address,
	}
}
