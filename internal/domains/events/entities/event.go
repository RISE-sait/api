package entity

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID
	Course    string
	Facility  string
	BeginTime time.Time
	EndTime   time.Time
	Day       string
}
