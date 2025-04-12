package event

import (
	"api/internal/di"
	repo "api/internal/domains/event/persistence/repository"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	"time"
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

func (s *Service) GetEvents(ctx context.Context, programTypeStr string, programID, locationID, userID, teamID, createdBy, updatedBy uuid.UUID, before, after time.Time) ([]values.ReadEventValues, *errLib.CommonError) {

	return s.repo.GetEvents(ctx, programTypeStr, programID, locationID, userID, teamID, createdBy, updatedBy, before, after)
}

func (s *Service) CreateEvent(ctx context.Context, details values.CreateEventValues) *errLib.CommonError {
	return s.repo.CreateEvent(ctx, details)
}

func (s *Service) UpdateEvent(ctx context.Context, details values.UpdateEventValues) *errLib.CommonError {
	return s.repo.UpdateEvent(ctx, details)
}

func (s *Service) DeleteEvent(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.repo.DeleteEvent(ctx, id)
}
