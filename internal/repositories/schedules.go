package repositories

import (
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type SchedulesRepository struct {
	Queries *db.Queries
}

func (r *SchedulesRepository) CreateSchedule(c context.Context, schedule *db.CreateScheduleParams) *utils.HTTPError {
	row, err := r.Queries.CreateSchedule(c, *schedule)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to create schedule: %+v", *schedule)
		return utils.CreateHTTPError("Failed to create schedule", http.StatusInternalServerError)
	}

	return nil
}

func (r *SchedulesRepository) GetSchedule(c context.Context, id uuid.NullUUID) (*db.Schedule, *utils.HTTPError) {
	schedule, err := r.Queries.GetScheduleByCourseID(c, id)

	if err != nil {

		log.Printf("Failed to retrieve schedule with ID: %s", id.UUID)
		return nil, utils.MapDatabaseError(err)
	}

	return &schedule, nil
}

func (r *SchedulesRepository) GetAllSchedules(c context.Context) (*[]db.Schedule, *utils.HTTPError) {
	schedules, err := r.Queries.GetAllSchedules(c)

	if err != nil {
		return &[]db.Schedule{}, utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}

	return &schedules, nil
}

func (r *SchedulesRepository) UpdateSchedule(c context.Context, schedule *db.UpdateScheduleParams) *utils.HTTPError {
	row, err := r.Queries.UpdateSchedule(c, *schedule)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to update schedule: %+v", *schedule)
		return utils.CreateHTTPError("Schedule not found", http.StatusNotFound)
	}

	return nil
}

func (r *SchedulesRepository) DeleteSchedule(c context.Context, id uuid.NullUUID) *utils.HTTPError {
	row, err := r.Queries.DeleteSchedule(c, id)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to delete schedule with ID: %s", id.UUID)
		return utils.CreateHTTPError("Schedule not found", http.StatusNotFound)
	}

	return nil
}
