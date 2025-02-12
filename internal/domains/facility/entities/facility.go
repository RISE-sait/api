package entity

import (
	"api/internal/domains/facility/values"

	"github.com/google/uuid"
)

type Facility struct {
	ID uuid.UUID
	values.FacilityDetails
}

type FacilityType struct {
	ID   uuid.UUID
	Name string
}
