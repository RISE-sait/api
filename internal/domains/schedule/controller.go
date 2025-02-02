package schedule

import (
	"api/cmd/server/di"
	dto "api/internal/domains/schedule/dto"
	"api/internal/domains/schedule/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

// SchedulesController provides HTTP handlers for managing schedules.
type SchedulesController struct {
	Service *SchedulesService
}

// NewController creates a new instance of SchedulesController.
func NewSchedulesController(container *di.Container) *SchedulesController {
	return &SchedulesController{Service: NewSchedulesService(container)}
}

// GetAllSchedules retrieves all schedules from the database.
func (c *SchedulesController) GetSchedules(w http.ResponseWriter, r *http.Request) {

	courseIdStr := r.URL.Query().Get("course_id")
	facilityIdStr := r.URL.Query().Get("facility_id")

	begin_datetimeStr := r.URL.Query().Get("begin_datetime")
	end_datetimeStr := r.URL.Query().Get("end_datetime")

	courseId, err := validators.ParseUUID(courseIdStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	facilityId, err := validators.ParseUUID(facilityIdStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	// parse datetime
	beginDatetime, err := validators.ParseDateTime(begin_datetimeStr)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	endDatetime, err := validators.ParseDateTime(end_datetimeStr)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	details := &values.ScheduleDetails{
		BeginDatetime: *beginDatetime,
		EndDatetime:   *endDatetime,
		CourseID:      courseId,
		FacilityID:    facilityId,
	}

	schedules, err := c.Service.GetSchedules(r.Context(), details)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.ScheduleResponse, len(schedules))

	for i, schedule := range schedules {
		result[i] = mapEntityToResponse(&schedule)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (c *SchedulesController) CreateSchedule(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.ScheduleRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	scheduleCreate, err := targetBody.ToScheduleDetails()

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := c.Service.CreateSchedule(r.Context(), scheduleCreate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (c *SchedulesController) UpdateSchedule(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var targetBody dto.ScheduleRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	params, err := targetBody.ToScheduleAllFields(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := c.Service.UpdateSchedule(r.Context(), params); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (c *SchedulesController) DeleteSchedule(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
	}

	if err = c.Service.DeleteSchedule(r.Context(), id); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func mapEntityToResponse(schedule *values.ScheduleAllFields) dto.ScheduleResponse {
	return dto.ScheduleResponse{
		ID:            schedule.ID,
		BeginDatetime: schedule.BeginDatetime,
		EndDatetime:   schedule.EndDatetime,
		CourseID:      schedule.CourseID,
		FacilityID:    schedule.FacilityID,
		Day:           schedule.Day,
	}
}
