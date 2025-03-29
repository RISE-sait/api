package membership

import (
	"github.com/google/uuid"
	"time"
)

type BaseValue struct {
	Name        string
	Description string
	Benefits    string
}

type CreateValues struct {
	BaseValue
}

type UpdateValues struct {
	ID uuid.UUID
	BaseValue
}

type ReadValues struct {
	ID uuid.UUID
	BaseValue
	CreatedAt time.Time
	UpdatedAt time.Time
}
