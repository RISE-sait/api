package values

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type PracticeDetails struct {
	Name        string
	Description string
	Level       string
	PayGPrice   *decimal.Decimal
}

type CreatePracticeValues struct {
	PracticeDetails
}

type UpdatePracticeValues struct {
	ID uuid.UUID
	PracticeDetails
}

type GetPracticeValues struct {
	PracticeDetails PracticeDetails
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
