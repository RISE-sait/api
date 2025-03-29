package values

import (
	"github.com/shopspring/decimal"
	"time"

	"github.com/google/uuid"
)

type ProgramDetails struct {
	Name        string
	Description string
	Level       string
	Type        string
	PayGPrice   decimal.NullDecimal
}

type CreateProgramValues struct {
	ProgramDetails
}

type UpdateProgramValues struct {
	ID uuid.UUID
	ProgramDetails
}

type GetProgramValues struct {
	ProgramDetails
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
