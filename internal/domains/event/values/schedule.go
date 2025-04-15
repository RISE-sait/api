package event

import (
	"github.com/google/uuid"
	"time"
)

type Schedule struct {
	DayOfWeek string
	StartTime string
	EndTime   string
	Program   struct {
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
	EventCount      int64
	FirstOccurrence time.Time
	LastOccurrence  time.Time
}
