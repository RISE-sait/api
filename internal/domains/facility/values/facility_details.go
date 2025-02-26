package facility

import (
	"github.com/google/uuid"
)

type Details struct {
	Name                 string
	Address              string
	FacilityCategoryName string
	FacilityCategoryID   uuid.UUID
}
