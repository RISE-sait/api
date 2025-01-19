package facilityType

import (
	db "api/sqlc"

	"github.com/google/uuid"
)

type UpdateFacilityTypeRequest struct {
	Id   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required_and_notwhitespace"`
}

func (r *UpdateFacilityTypeRequest) ToDBParams() *db.UpdateFacilityTypeParams {

	dbParams := db.UpdateFacilityTypeParams{
		ID:   r.Id,
		Name: r.Name,
	}

	return &dbParams
}
