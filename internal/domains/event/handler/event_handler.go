package event

import (
	"api/internal/di"
	dto "api/internal/domains/event/dto"
	"api/internal/domains/event/service"
	values "api/internal/domains/event/values"
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
	EventsService *service.Service
}

func NewEventsHandler(container *di.Container) *EventsHandler {
	return &EventsHandler{EventsService: service.NewEventService(container)}
}

// GetEvents retrieves all events based on filter criteria.
// @Summary Get events
// @Description Retrieve all events within a specific date range, with optional filters by course, location, game, and practice.
// @Tags events
// @Param after query string false "Start date of the events range (YYYY-MM-DD)" example("2025-03-01")
// @Param before query string false "End date of the events range (YYYY-MM-DD)" example("2025-03-31")
// @Param program_id query string false "Filter by program ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param participant_id query string false "Filter by participant ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
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
		after, before                                                      time.Time
		locationID, programID, participantID, teamID, createdBy, updatedBy uuid.UUID
		programType                                                        string
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

	if participantIDStr := query.Get("participant_id"); participantIDStr != "" {
		if id, err := validators.ParseUUID(participantIDStr); err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'participant_id' format, expected uuid format", http.StatusBadRequest))
			return
		} else {
			participantID = id
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

	if createdByStr := query.Get("created_by"); createdByStr != "" {
		if id, err := validators.ParseUUID(createdByStr); err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'created_by' format, expected uuid format", http.StatusBadRequest))
			return
		} else {
			createdBy = id
		}
	}

	if updatedByStr := query.Get("updated_by"); updatedByStr != "" {
		if id, err := validators.ParseUUID(updatedByStr); err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'updated_by' format, expected uuid format", http.StatusBadRequest))
			return
		} else {
			updatedBy = id
		}
	}

	if (after.IsZero() || before.IsZero()) &&
		(programID == uuid.Nil && participantID == uuid.Nil && locationID == uuid.Nil && teamID == uuid.Nil && programType == "") {
		responseHandlers.RespondWithError(w,
			errLib.New(`at least one of (before and after) or 
(program_id, participant_id, location_id, team_id, program_type, created_by, updated_by), must be provided`, http.StatusBadRequest))
		return
	}

	filter := values.GetEventsFilter{
		ProgramType:   programType,
		ProgramID:     programID,
		LocationID:    locationID,
		ParticipantID: participantID,
		TeamID:        teamID,
		CreatedBy:     createdBy,
		UpdatedBy:     updatedBy,
		Before:        before,
		After:         after,
	}

	events, err := h.EventsService.GetEvents(r.Context(), filter)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// empty list instead of var responseDto []dto.EventResponseDto, so that it would return empty list instead of nil if no events found
	responseDto := []dto.EventResponseDto{}

	for _, retrievedEvent := range events {

		eventDto := dto.NewEventResponseDto(retrievedEvent, false)

		responseDto = append(responseDto, eventDto)
	}

	responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)

}

// GetEvent retrieves detailed information about a specific event based on its ID.
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

	if retrievedEvent, err := h.EventsService.GetEvent(r.Context(), eventId); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {

		responseDto := dto.NewEventResponseDto(retrievedEvent, true)

		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}

// CreateEvents creates new events given its recurrence information.
// @Tags events
// @Accept json
// @Produce json
// @Security Bearer
// @Param event body dto.CreateRequestDto true "Event details"
// @Success 201 {object} map[string]interface{} "Event created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events [post]
func (h *EventsHandler) CreateEvents(w http.ResponseWriter, r *http.Request) {

	userID, ctxErr := contextUtils.GetUserID(r.Context())

	if ctxErr != nil {
		responseHandlers.RespondWithError(w, ctxErr)
		return
	}

	var targetBody dto.CreateRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if eventCreate, err := targetBody.ToCreateEventsValues(userID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		if err = h.EventsService.CreateEvents(r.Context(), eventCreate); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// UpdateEvent updates an existing event by ID.
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

	retrievedEvent, err := h.EventsService.GetEvent(r.Context(), params.ID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	isAdmin := userRole == contextUtils.RoleAdmin || userRole == contextUtils.RoleSuperAdmin
	isCreator := retrievedEvent.CreatedBy.ID == loggedInUserID

	// Check if the user is an admin or the creator of the retrievedEvent, if not, return forbidden
	if !isAdmin && !isCreator {
		responseHandlers.RespondWithError(w, errLib.New("You do not have permission to access this resource", http.StatusForbidden))
		return
	}

	if err = h.EventsService.UpdateEvent(r.Context(), params); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteEvents deletes multiple events by IDs.
// @Tags events
// @Accept json
// @Produce json
// @Security Bearer
// @Param ids body dto.DeleteRequestDto true "Array of Event IDs to delete"
// @Success 204 {object} map[string]interface{} "No Content: Events deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: One or more events not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events [delete]
func (h *EventsHandler) DeleteEvents(w http.ResponseWriter, r *http.Request) {

	// Get auth context
	userRole := r.Context().Value(contextUtils.RoleKey).(contextUtils.CtxRole)
	loggedInUserID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var targetBody dto.DeleteRequestDto

	if err = validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = validators.ValidateDto(&targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	filter := values.GetEventsFilter{
		Ids: targetBody.IDs,
	}

	retrievedEvents, err := h.EventsService.GetEvents(r.Context(), filter)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	isAdmin := userRole == contextUtils.RoleAdmin || userRole == contextUtils.RoleSuperAdmin

	for _, retrievedEvent := range retrievedEvents {
		if !isAdmin && retrievedEvent.CreatedBy.ID != loggedInUserID {
			responseHandlers.RespondWithError(w,
				errLib.New("You don't have permission to delete one or more events", http.StatusForbidden))
			return
		}
	}

	if err = h.EventsService.DeleteEvents(r.Context(), targetBody.IDs); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
