package discount

import (
	"time"

	"github.com/google/uuid"
)

type CreateValues struct {
	Name            string
	Description     string
	DiscountPercent int
	IsUseUnlimited  bool
	UsePerClient    int
	IsActive        bool
	ValidFrom       time.Time
	ValidTo         time.Time
}

type UpdateValues struct {
	ID uuid.UUID
	CreateValues
}

type ReadValues struct {
	ID uuid.UUID
	CreateValues
	CreatedAt time.Time
	UpdatedAt time.Time
}
