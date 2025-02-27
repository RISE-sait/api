package event

import (
	"github.com/google/uuid"
	"time"
)

type Event struct {
	ID            uuid.UUID
	PracticeID    *uuid.UUID
	CourseID      *uuid.UUID
	GameID        *uuid.UUID
	LocationID    uuid.UUID
	BeginDateTime time.Time
	EndDateTime   time.Time
}
