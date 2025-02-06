package schedule

import (
	"api/internal/di"
	dto "api/internal/domains/schedule/dto"
	entity "api/internal/domains/schedule/entities"
	"api/internal/domains/schedule/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
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

	beginTimeStr := r.URL.Query().Get("begin_datetime")
	endTimeStr := r.URL.Query().Get("end_datetime")

	var courseId uuid.UUID
	var facilityId uuid.UUID
	var beginDatetime time.Time
	var endDatetime time.Time

	if courseIdStr != "" {
		id, err := validators.ParseUUID(courseIdStr)

		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		courseId = id
	}

	if facilityIdStr != "" {

		id, err := validators.ParseUUID(facilityIdStr)

		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		facilityId = id
	}

	if beginTimeStr != "" {
		datetime, err := validators.ParseTime(beginTimeStr)
		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		beginDatetime = datetime
	}

	if endTimeStr != "" {
		datetime, err := validators.ParseTime(endTimeStr)
		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		endDatetime = datetime
	}

	details := values.ScheduleDetails{
		BeginTime:  beginDatetime,
		EndTime:    endDatetime,
		CourseID:   courseId,
		FacilityID: facilityId,
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

	params, err := (&targetBody).ToScheduleAllFields(idStr)

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

func mapEntityToResponse(schedule *entity.Schedule) dto.ScheduleResponse {
	return dto.ScheduleResponse{
		ID:        schedule.ID,
		BeginTime: schedule.BeginTime.Format("15:04"), // Convert to "HH:MM:SS"
		EndTime:   schedule.EndTime.Format("15:04"),
		Course:    schedule.Course,
		Facility:  schedule.Facility,
		Day:       schedule.Day,
	}
}
