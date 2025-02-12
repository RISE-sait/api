package entity

import (
	"time"

	"github.com/google/uuid"
)

type Course struct {
	ID   uuid.UUID
	Name string
}

type Event struct {
	ID         uuid.UUID
	Course     *Course
	Facility   string
	FacilityID uuid.UUID
	BeginTime  time.Time
	EndTime    time.Time
	Day        string
}
