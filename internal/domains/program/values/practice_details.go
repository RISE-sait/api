package values

import (
	"time"

	"github.com/google/uuid"
)

type ProgramDetails struct {
	Name        string
	Description string
	Level       string
	Type        string
	Capacity    *int32
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
