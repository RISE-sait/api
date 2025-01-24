package entity

import (
	"time"

	"github.com/google/uuid"
)

type Membership struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}
