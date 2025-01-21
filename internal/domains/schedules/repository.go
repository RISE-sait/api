package schedules

import (
	"api/internal/types"
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

func (r *Repository) CreateSchedule(c context.Context, schedule *db.CreateScheduleParams) *types.HTTPError {
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

func (r *Repository) GetSchedule(c context.Context, id uuid.NullUUID) (*db.Schedule, *types.HTTPError) {
	schedule, err := r.Queries.GetScheduleByCourseID(c, id)

	if err != nil {

		log.Printf("Failed to retrieve schedule with ID: %s", id.UUID)
		return nil, utils.MapDatabaseError(err)
	}

	return &schedule, nil
}

func (r *Repository) GetAllSchedules(c context.Context) (*[]db.Schedule, *types.HTTPError) {
	schedules, err := r.Queries.GetAllSchedules(c)

	if err != nil {
		return &[]db.Schedule{}, utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}

	return &schedules, nil
}

func (r *Repository) UpdateSchedule(c context.Context, schedule *db.UpdateScheduleParams) *types.HTTPError {
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

func (r *Repository) DeleteSchedule(c context.Context, id uuid.NullUUID) *types.HTTPError {
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
