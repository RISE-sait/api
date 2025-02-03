package dto

import (
	"github.com/google/uuid"
)

type FacilityResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Location     string    `json:"location"`
	FacilityType string    `json:"facility_type"`
}

type FacilityTypeResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
