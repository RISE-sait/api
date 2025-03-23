package practice

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type Details struct {
	Name        string
	Description string
	Level       string
	PayGPrice   *decimal.Decimal
}

type CreatePracticeValues struct {
	Details
}

type UpdatePracticeValues struct {
	ID uuid.UUID
	Details
}

type GetPracticeValues struct {
	Details
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
