package course

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type Details struct {
	Name        string
	Description string
	PayGPrice   *decimal.Decimal
}

type CreateCourseDetails struct {
	Details
}

type UpdateCourseDetails struct {
	ID uuid.UUID
	Details
}

type ReadDetails struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Details
}
