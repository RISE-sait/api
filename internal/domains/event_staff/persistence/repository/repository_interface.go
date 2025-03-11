package event_staff

import (
	values "api/internal/domains/event_staff/values"
	staffValues "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type EventStaffsRepositoryInterface interface {
	AssignStaffToEvent(c context.Context, input values.EventStaff) *errLib.CommonError
	GetStaffsAssignedToEvent(ctx context.Context, eventId uuid.UUID) ([]staffValues.ReadValues, *errLib.CommonError)
	UnassignedStaffFromEvent(c context.Context, input values.EventStaff) *errLib.CommonError
}
