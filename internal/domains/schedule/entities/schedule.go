package entity

import (
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	ID            uuid.UUID
	Course        string
	Facility      string
	BeginDatetime time.Time
	EndDatetime   time.Time
	Day           string
}
