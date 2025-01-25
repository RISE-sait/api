package entities

import (
	"time"

	"github.com/google/uuid"
)

type Course struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}
