package event

import (
	dto "api/internal/domains/event/dto"
	repository "api/internal/domains/event/persistence/repository"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// EventsHandler provides HTTP handlers for managing events.
type EventsHandler struct {
	Repo *repository.EventsRepository
}

func NewEventsHandler(repo *repository.EventsRepository) *EventsHandler {
	return &EventsHandler{Repo: repo}
}

// GetEvents retrieves all events based on filter criteria.
// @Summary Get events
// @Description Retrieve all events within a specific date range, with optional filters by course, location, game, and practice.
// @Tags events
// @Param after query string false "Start date of the events range (YYYY-MM-DD)" example("2025-03-01")
// @Param before query string false "End date of the events range (YYYY-MM-DD)" example("2025-03-31")
// @Param program_id query string false "Filter by program ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param user_id query string false "Filter by user ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param team_id query string false "Filter by team ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param location_id query string false "Filter by location ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param program_type query string false "Program Type (game, practice, course, others)"
// @Param created_by query string false "ID of person who created the event (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")"
// @Param updated_by query string false "ID of person who updated the event (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Accept json
// @Produce json
// @Success 200 {array} dto.EventResponseDto "List of events retrieved successfully"
// @Schema(oneOf={[]event.DayResponseDto,[]event.DateResponseDto})
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input format or missing required parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events [get]
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {

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

	events, err := h.Repo.GetEvents(r.Context(), programType, programID, locationID, userID, teamID, createdBy, updatedBy, before, after)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// empty list instead of var responseDto []dto.EventResponseDto, so that it would return empty list instead of nil if no events found
	responseDto := []dto.EventResponseDto{}

	for _, event := range events {

		eventDto := dto.NewEventResponseDto(event)

		responseDto = append(responseDto, eventDto)
	}

	responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
}

// GetEvent retrieves detailed information about a specific event.
// @Summary Get event details
// @Description Retrieves details of a specific event based on its ID.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param view query string false "Choose between 'date' and 'day'. Response type for the schedule, in specific dates or recurring day information. Default is 'day'."
// @Success 200 {object} dto.EventResponseDto "Event details retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{id} [get]
func (h *EventsHandler) GetEvent(w http.ResponseWriter, r *http.Request) {

	var eventId uuid.UUID

	if eventIdStr := chi.URLParam(r, "id"); eventIdStr != "" {

		if id, err := validators.ParseUUID(eventIdStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			eventId = id
		}
	}

	if event, err := h.Repo.GetEvent(r.Context(), eventId); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {

		responseDto := dto.NewEventResponseDto(event)

		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}

// CreateEvent creates a new event.
// @Summary Create a new event
// @Description Registers a new event with the provided details.
// @Tags events
// @Accept json
// @Produce json
// @Security Bearer
// @Param event body dto.CreateRequestDto true "Event details"
// @Success 201 {object} map[string]interface{} "Event created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events [post]
func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {

	userID, err := contextUtils.GetUserID(r.Context())

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var targetBody dto.CreateRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if eventCreate, err := targetBody.ToCreateEventValues(userID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		if err = h.Repo.CreateEvent(r.Context(), eventCreate); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// UpdateEvent updates an existing event by ID.
// @Summary Update an event
// @Description Updates the details of an existing event.
// @Tags events
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Event ID"
// @Param event body dto.UpdateRequestDto true "Updated event details"
// @Success 204 {object} map[string]interface{} "No Content: Event updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{id} [put]
func (h *EventsHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {

	userID, err := contextUtils.GetUserID(r.Context())

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Extract and validate input

	idStr := chi.URLParam(r, "id")

	var targetBody dto.UpdateRequestDto

	if err = validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Convert to domain values

	params, err := targetBody.ToUpdateEventValues(idStr, userID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Get auth context

	userRole := r.Context().Value(contextUtils.RoleKey).(contextUtils.CtxRole)

	loggedInUserID, err := contextUtils.GetUserID(r.Context())

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Authorization check

	creatorID, err := h.Repo.GetEventCreatedBy(r.Context(), params.ID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	isAdmin := userRole == contextUtils.RoleAdmin || userRole == contextUtils.RoleSuperAdmin
	isCreator := creatorID == loggedInUserID

	// Check if the user is an admin or the creator of the event, if not, return forbidden
	if !isAdmin && !isCreator {
		responseHandlers.RespondWithError(w, errLib.New("You do not have permission to access this resource", http.StatusForbidden))
		return
	}

	if err = h.Repo.UpdateEvent(r.Context(), params); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteEvent deletes an event by ID.
// @Summary Delete an event
// @Description Deletes an event by its ID.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 204 {object} map[string]interface{} "No Content: Event deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{id} [delete]
func (h *EventsHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Get auth context

	userRole := r.Context().Value(contextUtils.RoleKey).(contextUtils.CtxRole)

	loggedInUserID, err := contextUtils.GetUserID(r.Context())

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Authorization check

	creatorID, err := h.Repo.GetEventCreatedBy(r.Context(), id)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	isAdmin := userRole == contextUtils.RoleAdmin || userRole == contextUtils.RoleSuperAdmin
	isCreator := creatorID == loggedInUserID

	// Check if the user is an admin or the creator of the event, if not, return forbidden
	if !isAdmin && !isCreator {
		responseHandlers.RespondWithError(w, errLib.New("You do not have permission to access this resource", http.StatusForbidden))
		return
	}

	if err = h.Repo.DeleteEvent(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
