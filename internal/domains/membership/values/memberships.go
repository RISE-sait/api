package membership

import (
	"github.com/google/uuid"
	"time"
)

type CreateValues struct {
	Name        string
	Description string
}

type UpdateValues struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type ReadValues struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
