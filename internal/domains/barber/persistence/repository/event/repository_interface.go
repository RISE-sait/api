package event

import (
	values "api/internal/domains/barber/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type BarberEventsRepositoryInterface interface {
	CreateEvent(ctx context.Context, eventDetails values.CreateEventValues) (values.EventReadValues, *errLib.CommonError)
	GetEvents(ctx context.Context) ([]values.EventReadValues, *errLib.CommonError)
	UpdateEvent(ctx context.Context, event values.UpdateEventValues) (values.EventReadValues, *errLib.CommonError)
	DeleteEvent(c context.Context, id uuid.UUID) *errLib.CommonError
	GetEventDetails(ctx context.Context, id uuid.UUID) (values.EventReadValues, *errLib.CommonError)
}
