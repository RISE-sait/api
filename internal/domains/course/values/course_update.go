package values

import (
	"time"

	"github.com/google/uuid"
)

type CourseUpdate struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}
