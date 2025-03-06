package event

import (
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type IEventsRepository interface {
	CreateEvent(ctx context.Context, event values.CreateEventValues) (values.ReadEventValues, *errLib.CommonError)
	GetEvents(ctx context.Context, courseId, locationId, practiceId, gameId uuid.UUID) ([]values.ReadEventValues, *errLib.CommonError)
	UpdateEvent(ctx context.Context, event values.UpdateEventValues) (values.ReadEventValues, *errLib.CommonError)
	DeleteEvent(ctx context.Context, id uuid.UUID) *errLib.CommonError
	GetEvent(ctx context.Context, id uuid.UUID) (values.ReadEventValues, *errLib.CommonError)
}
