package event_staff

import (
	values "api/internal/domains/event_staff/values"
)

// RequestDto represents the request payload for creating an event-staff relationship.
type RequestDto struct {
	Base EventStaffBase
}

func (dto *RequestDto) ToDetails() values.EventStaff {
	return values.EventStaff{
		EventID: dto.Base.EventID,
		StaffID: dto.Base.StaffID,
	}
}
