package course

import (
	"github.com/google/uuid"
	"time"
)

type Details struct {
	Name        string
	Description string
}

type CreateCourseDetails struct {
	Details
}

type UpdateCourseDetails struct {
	ID uuid.UUID
	Details
}

type ReadDetails struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Details
}
