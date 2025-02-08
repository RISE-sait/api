package events

import (
	"api/internal/di"
	entity "api/internal/domains/events/entities"
	"api/internal/domains/events/persistence"
	"api/internal/domains/events/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

// EventsService provides HTTP handlers for managing schedules (events).
type EventsService struct {
	Repo *persistence.EventsRepository
}

// NewEventsService creates a new instance of EventsService.
func NewEventsService(container *di.Container) *EventsService {
	return &EventsService{Repo: persistence.NewEventsRepository(container)}
}

// GetAllSchedules retrieves all events from the database.
func (s *EventsService) GetEvents(ctx context.Context, fields values.EventDetails) ([]entity.Event, *errLib.CommonError) {
	return s.Repo.GetEvents(ctx, fields)
}

func (s *EventsService) CreateEvent(ctx context.Context, fields *values.EventDetails) *errLib.CommonError {
	return s.Repo.CreateEvent(ctx, fields)
}

func (s *EventsService) UpdateEvent(ctx context.Context, fields *values.EventAllFields) *errLib.CommonError {
	return s.Repo.UpdateEvent(ctx, fields)
}

func (s *EventsService) DeleteEvent(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteEvent(ctx, id)
}

func (s *EventsService) GetCustomersCountByEventId(ctx context.Context, id uuid.UUID) (int64, *errLib.CommonError) {
	return s.Repo.GetCustomersCountByEventId(ctx, id)
}
