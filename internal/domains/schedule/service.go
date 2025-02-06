package schedule

import (
	"api/internal/di"
	entity "api/internal/domains/schedule/entities"
	"api/internal/domains/schedule/persistence"
	"api/internal/domains/schedule/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

// SchedulesService provides HTTP handlers for managing schedules.
type SchedulesService struct {
	Repo *persistence.SchedulesRepository
}

// NewController creates a new instance of SchedulesController.
func NewSchedulesService(container *di.Container) *SchedulesService {
	return &SchedulesService{Repo: persistence.NewScheduleRepository(container)}
}

// GetAllSchedules retrieves all schedules from the database.
func (s *SchedulesService) GetSchedules(ctx context.Context, fields values.ScheduleDetails) ([]entity.Schedule, *errLib.CommonError) {
	return s.Repo.GetSchedules(ctx, fields)
}

func (s *SchedulesService) CreateSchedule(ctx context.Context, fields *values.ScheduleDetails) *errLib.CommonError {

	return s.Repo.CreateSchedule(ctx, fields)
}

func (s *SchedulesService) UpdateSchedule(ctx context.Context, fields *values.ScheduleAllFields) *errLib.CommonError {

	return s.Repo.UpdateSchedule(ctx, fields)
}

func (s *SchedulesService) DeleteSchedule(ctx context.Context, id uuid.UUID) *errLib.CommonError {

	return s.Repo.DeleteSchedule(ctx, id)
}
