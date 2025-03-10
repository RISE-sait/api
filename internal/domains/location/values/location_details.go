package facility

import (
	"github.com/google/uuid"
)

type BaseDetails struct {
	Name    string
	Address string
}

type CreateDetails struct {
	BaseDetails
}

type UpdateDetails struct {
	BaseDetails
	ID uuid.UUID
}

type ReadValues struct {
	ID uuid.UUID
	BaseDetails
}
