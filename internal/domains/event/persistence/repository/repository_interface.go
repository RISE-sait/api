package event

import (
	entity "api/internal/domains/event/entity"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type EventsRepositoryInterface interface {
	CreateEvent(ctx context.Context, event *values.Details) (entity.Event, *errLib.CommonError)
	GetEvents(ctx context.Context, courseId, locationId, practiceId, gameId *uuid.UUID) ([]entity.Event, *errLib.CommonError)
	UpdateEvent(ctx context.Context, event *entity.Event) (*entity.Event, *errLib.CommonError)
	DeleteEvent(ctx context.Context, id uuid.UUID) *errLib.CommonError
	GetEventDetails(ctx context.Context, id uuid.UUID) (*entity.Event, *errLib.CommonError)
}
