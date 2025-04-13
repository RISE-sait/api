package event

import (
	"api/internal/di"
	repo "api/internal/domains/event/persistence/repository"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type Service struct {
	repo *repo.EventsRepository
}

func NewEventService(container *di.Container) *Service {
	return &Service{
		repo: repo.NewEventsRepository(container),
	}
}

func (s *Service) GetEvent(ctx context.Context, eventID uuid.UUID) (values.ReadEventValues, *errLib.CommonError) {
	return s.repo.GetEvent(ctx, eventID)
}

func (s *Service) GetEvents(ctx context.Context, filter values.GetEventsFilter) ([]values.ReadEventValues, *errLib.CommonError) {
	return s.repo.GetEvents(ctx, filter)
}

func (s *Service) CreateEvents(ctx context.Context, details values.CreateEventsValues) *errLib.CommonError {
	return s.repo.CreateEvents(ctx, details)
}

func (s *Service) UpdateEvent(ctx context.Context, details values.UpdateEventValues) *errLib.CommonError {
	return s.repo.UpdateEvent(ctx, details)
}

func (s *Service) DeleteEvents(ctx context.Context, ids []uuid.UUID) *errLib.CommonError {
	return s.repo.DeleteEvent(ctx, ids)
}
