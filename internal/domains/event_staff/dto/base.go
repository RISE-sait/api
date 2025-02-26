package event_staff

import "github.com/google/uuid"

// EventStaffBase contains common fields for event-staff-related DTOs.
type EventStaffBase struct {
	EventID uuid.UUID `json:"event_id"`
	StaffID uuid.UUID `json:"staff_id"`
}
