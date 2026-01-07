package practice

import (
	"time"

	"github.com/google/uuid"
)

type CreatePracticeValue struct {
	TeamID     uuid.UUID
	StartTime  time.Time
	EndTime    *time.Time
	LocationID uuid.UUID
	CourtID    uuid.UUID
	Status     string
	BookedBy   *uuid.UUID
}

type UpdatePracticeValue struct {
	ID               uuid.UUID
	SkipNotification bool // Skip auto-notification on update
	CreatePracticeValue
}

type ReadPracticeValue struct {
	ID             uuid.UUID
	TeamID         uuid.UUID
	TeamName       string
	TeamLogoUrl    string
	StartTime      time.Time
	EndTime        *time.Time
	LocationID     uuid.UUID
	LocationName   string
	CourtID        uuid.UUID
	CourtName      string
	Status         string
	BookedBy       *uuid.UUID
	BookedByName   string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
}

type RecurrenceValues struct {
	DayOfWeek       time.Weekday
	FirstOccurrence time.Time
	LastOccurrence  time.Time
	StartTime       string
	EndTime         string
}
