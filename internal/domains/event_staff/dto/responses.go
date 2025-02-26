package event_staff

import (
	entity "api/internal/domains/event_staff/values"
)

// ResponseDto represents the response payload for event-staff-related operations.
type ResponseDto struct {
	EventStaffBase
}

// NewEventStaffResponse creates a new ResponseDto from the domain model.
func NewEventStaffResponse(entity entity.EventStaff) *ResponseDto {
	return &ResponseDto{
		EventStaffBase: EventStaffBase{
			EventID: entity.EventID,
			StaffID: entity.StaffID,
		},
	}
}
