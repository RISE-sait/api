package values

import "github.com/google/uuid"

type ScheduleAllFields struct {
	ScheduleDetails
	ID uuid.UUID
}
