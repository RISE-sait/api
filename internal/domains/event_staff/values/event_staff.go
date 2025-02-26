package event_staff

import (
	"github.com/google/uuid"
)

type EventStaff struct {
	EventID uuid.UUID
	StaffID uuid.UUID
}
