package event

import (
	dto "api/internal/domains/schedule/dto"
	repository "api/internal/domains/schedule/persistence/repository"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// SchedulesHandler provides HTTP handlers for managing events.
type SchedulesHandler struct {
	Repo *repository.ScheduleRepository
}

func NewSchedulesHandler(repo *repository.ScheduleRepository) *SchedulesHandler {
	return &SchedulesHandler{Repo: repo}
}

// GetSchedules retrieves all schedules based on filter criteria.
// @Summary Get schedules
// @Description Retrieve all schedules with optional filters by program, location, user, team, and program type
// @Tags schedules
// @Param program_id query string false "Filter by program ID (UUID format)" Format(uuid)
// @Param user_id query string false "Filter by user ID (UUID format)" Format(uuid)
// @Param team_id query string false "Filter by team ID (UUID format)" Format(uuid)
// @Param location_id query string false "Filter by location ID (UUID format)" Format(uuid)
// @Param program_type query string false "Filter by program type (practice, course, game, others)"
// @Accept json
// @Produce json
// @Success 200 {array} dto.ScheduleResponseDto "List of schedules retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /schedules [get]
func (h *SchedulesHandler) GetSchedules(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	var (
		locationID, programID, userID, teamID uuid.UUID
		programType                           string
	)

	if userIDStr := query.Get("user_id"); userIDStr != "" {
		if id, err := validators.ParseUUID(userIDStr); err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'user_id' format, expected uuid format", http.StatusBadRequest))
			return
		} else {
			userID = id
		}
	}

	if teamIDStr := query.Get("team_id"); teamIDStr != "" {
		if id, err := validators.ParseUUID(teamIDStr); err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'team_id' format, expected uuid format", http.StatusBadRequest))
			return
		} else {
			teamID = id
		}
	}

	if programIDStr := query.Get("program_id"); programIDStr != "" {
		if id, err := validators.ParseUUID(programIDStr); err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'program_id' format, expected uuid format", http.StatusBadRequest))
			return
		} else {
			programID = id
		}
	}

	programType = query.Get("program_type")

	if locationIDStr := query.Get("location_id"); locationIDStr != "" {
		if id, err := validators.ParseUUID(locationIDStr); err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'location_id' format, expected uuid format", http.StatusBadRequest))
			return
		} else {
			locationID = id
		}
	}

	schedules, err := h.Repo.GetSchedules(r.Context(), programID, locationID, userID, teamID, programType)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseDto := []dto.ScheduleResponseDto{}

	for _, schedule := range schedules {

		scheduleResponseDto := dto.NewScheduleResponseDto(schedule)

		responseDto = append(responseDto, scheduleResponseDto)
	}

	responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
}

// GetSchedule retrieves detailed information about a specific schedule.
// @Summary Get schedule details
// @Description Retrieves details of a specific schedule by ID
// @Tags schedules
// @Param id path string true "Schedule ID" Format(uuid)
// @Accept json
// @Produce json
// @Success 200 {object} dto.ScheduleResponseDto "Schedule details retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid schedule ID format"
// @Failure 404 {object} map[string]interface{} "Schedule not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /schedules/{id} [get]
func (h *SchedulesHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {

	var scheduleID uuid.UUID

	if eventIdStr := chi.URLParam(r, "id"); eventIdStr != "" {

		if id, err := validators.ParseUUID(eventIdStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			scheduleID = id
		}
	}

	if schedule, err := h.Repo.GetSchedule(r.Context(), scheduleID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {

		responseDto := dto.NewScheduleResponseDto(schedule)

		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}

// CreateSchedule creates a new schedule.
// @Summary Create schedule
// @Description Creates a new schedule with the provided details
// @Tags schedules
// @Security BearerAuth
// @Param request body dto.ScheduleRequestDto true "Schedule creation data"
// @Accept json
// @Produce json
// @Success 201 {object} dto.ScheduleResponseDto "Schedule created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /schedules [post]
func (h *SchedulesHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.ScheduleRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if scheduleValues, err := targetBody.ToCreateScheduleValues(); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		if err = h.Repo.CreateSchedule(r.Context(), scheduleValues); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// UpdateSchedule updates an existing schedule.
// @Summary Update schedule
// @Description Updates the specified schedule with new data
// @Tags schedules
// @Security BearerAuth
// @Param id path string true "Schedule ID" Format(uuid)
// @Param request body dto.ScheduleRequestDto true "Schedule update data"
// @Accept json
// @Produce json
// @Success 204 "Schedule updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Schedule not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /schedules/{id} [put]
func (h *SchedulesHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {

	// Extract and validate input

	idStr := chi.URLParam(r, "id")

	var targetBody dto.ScheduleRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Convert to domain values

	params, err := targetBody.ToUpdateScheduleValues(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.UpdateSchedule(r.Context(), params); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteSchedule deletes a schedule.
// @Summary Delete schedule
// @Description Deletes the specified schedule
// @Tags schedules
// @Security BearerAuth
// @Param id path string true "Schedule ID" Format(uuid)
// @Accept json
// @Produce json
// @Success 204 "Schedule deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid schedule ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Schedule not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /schedules/{id} [delete]
func (h *SchedulesHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.DeleteSchedule(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
