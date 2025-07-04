package event

import (
	"time"

	"github.com/google/uuid"
)

type BaseRecurrenceValues struct {
	DayOfWeek       time.Weekday
	FirstOccurrence time.Time
	LastOccurrence  time.Time
	StartTime       string
	EndTime         string
}

type ReadRecurrenceValues struct {
	BaseRecurrenceValues
	ID      uuid.UUID
	Program struct {
		ID          uuid.UUID
		Name        string
		Description string
		Type        string
	}
	Team *struct {
		ID   uuid.UUID
		Name string
	}
	Location struct {
		ID      uuid.UUID
		Name    string
		Address string
	}
	EventCount int64
}

type CreateRecurrenceValues struct {
	BaseRecurrenceValues
	ProgramID  uuid.UUID
	TeamID     uuid.UUID
	LocationID uuid.UUID
	CourtID    uuid.UUID
	CreatedBy  uuid.UUID
}

type UpdateRecurrenceValues struct {
	BaseRecurrenceValues
	ID         uuid.UUID
	ProgramID  uuid.UUID
	TeamID     uuid.UUID
	LocationID uuid.UUID
	CourtID    uuid.UUID
	UpdatedBy  uuid.UUID
}
