package event_capacity

import (
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type EventCapacityRepositoryInterface interface {
	GetEventIsFull(c context.Context, eventId uuid.UUID) (*bool, *errLib.CommonError)
}
