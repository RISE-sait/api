package values

import (
	"github.com/google/uuid"
	"time"
)

type PracticeDetails struct {
	Name        string
	Description string
}

type CreatePracticeValues struct {
	PracticeDetails
}

type UpdatePracticeValues struct {
	ID              uuid.UUID
	PracticeDetails PracticeDetails
}

type GetPracticeValues struct {
	PracticeDetails PracticeDetails
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
