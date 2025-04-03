package event

import (
	dto "api/internal/domains/event/dto"
	repository "api/internal/domains/event/persistence/repository"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// SchedulesHandler provides HTTP handlers for managing events.
type SchedulesHandler struct {
	Repo *repository.SchedulesRepository
}

func NewSchedulesHandler(schedulesRepo *repository.SchedulesRepository) *SchedulesHandler {
	return &SchedulesHandler{Repo: schedulesRepo}
}

// GetSchedules retrieves all events schedules based on filter criteria.
// @Tags Schedules
// @Accept json
// @Produce json
// @Param after query string false "Start date of the events range (YYYY-MM-DD format)" Example("2025-03-01")
// @Param before query string false "End date of the events range (YYYY-MM-DD format)" Example("2025-03-31")
// @Param program_id query string false "Filter by program ID (UUID format)" Example("550e8400-e29b-41d4-a716-446655440000")
// @Param user_id query string false "Filter by user ID (UUID format)" Example("550e8400-e29b-41d4-a716-446655440000")
// @Param team_id query string false "Filter by team ID (UUID format)" Example("550e8400-e29b-41d4-a716-446655440000")
// @Param location_id query string false "Filter by location ID (UUID format)" Example("550e8400-e29b-41d4-a716-446655440000")
// @Param program_type query string false "Filter by program type" Enums(game,practice,course,others)
// @Param created_by query string false "Filter by creator ID (UUID format)" Example("550e8400-e29b-41d4-a716-446655440000")
// @Param updated_by query string false "Filter by updater ID (UUID format)" Example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {array} dto.ScheduleResponseDto "Schedule retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /schedules [get]
func (h *SchedulesHandler) GetSchedules(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	var (
		after, before                                               time.Time
		locationID, programID, userID, teamID, createdBy, updatedBy uuid.UUID
		programType                                                 string
	)

	if afterStr := query.Get("after"); afterStr != "" {
		if afterDate, formatErr := time.Parse("2006-01-02", afterStr); formatErr != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'after' date format, expected YYYY-MM-DD", http.StatusBadRequest))
			return
		} else {
			after = afterDate
		}
	}

	if beforeStr := query.Get("before"); beforeStr != "" {
		if beforeDate, formatErr := time.Parse("2006-01-02", beforeStr); formatErr != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'before' date format, expected YYYY-MM-DD", http.StatusBadRequest))
			return
		} else {
			before = beforeDate
		}
	}

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

	if (after.IsZero() || before.IsZero()) &&
		(programID == uuid.Nil && userID == uuid.Nil && locationID == uuid.Nil && teamID == uuid.Nil && programType == "") {
		responseHandlers.RespondWithError(w,
			errLib.New(`at least one of (before and after) or 
(program_id, user_id, location_id, team_id, program_type, created_by, updated_by), must be provided`, http.StatusBadRequest))
		return
	}

	schedules, err := h.Repo.GetEventsSchedules(r.Context(), programType, programID, locationID, userID, teamID, createdBy, updatedBy, before, after)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// empty list instead of var responseDto []dto.EventResponseDto, so that it would return empty list instead of nil if no events found
	responseDto := []dto.ScheduleResponseDto{}

	for _, schedule := range schedules {

		scheduleDto := dto.NewScheduleResponseDto(schedule)

		responseDto = append(responseDto, scheduleDto)
	}

	responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)

}
