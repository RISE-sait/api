package event

import (
	"api/internal/custom_types"
	db "api/internal/domains/event/persistence/sqlc/generated"
	"github.com/google/uuid"
)

type Event struct {
	ID         uuid.UUID
	PracticeID *uuid.UUID
	CourseID   *uuid.UUID
	LocationID uuid.UUID
	BeginTime  custom_types.TimeWithTimeZone
	EndTime    custom_types.TimeWithTimeZone
	Day        db.DayEnum
}
