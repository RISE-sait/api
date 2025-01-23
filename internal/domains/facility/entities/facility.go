package entity

import "github.com/google/uuid"

type Facility struct {
	ID             uuid.UUID
	Name           string
	Location       string
	FacilityTypeID uuid.UUID
}
