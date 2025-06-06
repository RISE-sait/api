package event

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"api/internal/di"
	dto "api/internal/domains/event/dto"
	"api/internal/domains/event/service"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"

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
// @Tags events
// @Param after query string false "Start date of the events range (YYYY-MM-DD)" Format(date) example("2025-03-01")
// @Param before query string false "End date of the events range (YYYY-MM-DD)" Format(date) example("2025-03-31")
// @Param month query string false "Convenience month filter in YYYY-MM format"
// @Param day query string false "Convenience day filter in YYYY-MM-DD format"
// @Param program_id query string false "Filter by program ID" Format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param participant_id query string false "Filter by participant ID" Format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param team_id query string false "Filter by team ID" Format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param location_id query string false "Filter by location ID" Format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param program_type query string false "Filter by program type" Enums(game,practice,course,others) example(practice)
// @Param response_type query string false "Response format type" Enums(date,day) default(date) example(date)
// @Param created_by query string false "ID of person who created the event" Format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param updated_by query string false "ID of person who updated the event" Format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param limit query int false "Number of items per page" minimum(1) example(10)
// @Param offset query int false "Number of items to skip (for pagination)" minimum(0) example(20)
// @Param page query int false "Page number (alternative to offset)" minimum(1) example(2)
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Returns a list of events with pagination metadata"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input format or missing required parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events [get]
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	type FilterKey string

	const (
		FilterKeyParticipantID FilterKey = "participant_id"
		FilterKeyTeamID        FilterKey = "team_id"
		FilterKeyProgramID     FilterKey = "program_id"
		FilterKeyLocationID    FilterKey = "location_id"
		FilterKeyCreatedBy     FilterKey = "created_by"
		FilterKeyUpdatedBy     FilterKey = "updated_by"
	)

	query := r.URL.Query()

	limit := 20
	page := 1
	offset := 0
	if val := query.Get("limit"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if val := query.Get("page"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
			page = parsed
		}

	}

	if val := query.Get("offset"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			offset = parsed
		}
	} else {
		// use page based pagination if offset is not provided
		offset = (page - 1) * limit
	}

	var (
		after, before             time.Time
		uuidFilters               = make(map[string]uuid.UUID)
		programType, responseType string
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

	// Convenience parameters for calendar views
	monthStr := query.Get("month") // format YYYY-MM
	dayStr := query.Get("day")     // format YYYY-MM-DD
	if monthStr != "" && dayStr != "" {
		responseHandlers.RespondWithError(w, errLib.New("cannot provide both month and day parameters", http.StatusBadRequest))
		return
	}
	if after.IsZero() && before.IsZero() {
		if monthStr != "" {
			month, err := time.Parse("2006-01", monthStr)
			if err != nil {
				responseHandlers.RespondWithError(w, errLib.New("invalid 'month' format, expected YYYY-MM", http.StatusBadRequest))
				return
			}
			after = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
			before = after.AddDate(0, 1, -1)
		} else if dayStr != "" {
			day, err := time.Parse("2006-01-02", dayStr)
			if err != nil {
				responseHandlers.RespondWithError(w, errLib.New("invalid 'day' format, expected YYYY-MM-DD", http.StatusBadRequest))
				return
			}
			after = day
			before = day
		}
	}

	uuidFilterFields := []FilterKey{
		FilterKeyParticipantID,
		FilterKeyTeamID,
		FilterKeyProgramID,
		FilterKeyLocationID,
		FilterKeyCreatedBy,
		FilterKeyUpdatedBy,
	}

	for _, filterKey := range uuidFilterFields {
		if paramValue := query.Get(string(filterKey)); paramValue != "" {
			if id, err := validators.ParseUUID(paramValue); err != nil {
				responseHandlers.RespondWithError(w, errLib.New(
					fmt.Sprintf("invalid '%s' format, expected uuid format", filterKey),
					http.StatusBadRequest,
				))
				return
			} else {
				// Use the same key for both query param and struct field
				uuidFilters[string(filterKey)] = id
			}
		}
	}

	programType = query.Get("program_type")
	responseType = query.Get("response_type")

	if responseType == "" {
		responseType = "date"
	}
	if responseType != "date" && responseType != "day" {
		responseHandlers.RespondWithError(w, errLib.New("invalid 'response_type', must be 'date' or 'day'", http.StatusBadRequest))
		return
	}

	if (after.IsZero() || before.IsZero()) &&
		(len(uuidFilters) == 0 && programType == "") {
		responseHandlers.RespondWithError(w,
			errLib.New(`at least one of (before and after) or 
(program_id, participant_id, location_id, team_id, program_type, created_by, updated_by), must be provided`, http.StatusBadRequest))
		return
	}

	filter := values.GetEventsFilter{
		ProgramType:   programType,
		ProgramID:     uuidFilters[string(FilterKeyProgramID)],
		LocationID:    uuidFilters[string(FilterKeyLocationID)],
		ParticipantID: uuidFilters[string(FilterKeyParticipantID)],
		TeamID:        uuidFilters[string(FilterKeyTeamID)],
		CreatedBy:     uuidFilters[string(FilterKeyCreatedBy)],
		UpdatedBy:     uuidFilters[string(FilterKeyUpdatedBy)],
		Before:        before,
		After:         after,
		Limit:         limit,
		Offset:        offset,
	}

	userID, idErr := contextUtils.GetUserID(r.Context())
	role, roleErr := contextUtils.GetUserRole(r.Context())

	if responseType == "date" {
		var events []values.ReadEventValues

		if idErr == nil && roleErr == nil && (role == contextUtils.RoleAthlete || role == contextUtils.RoleCoach) {
			events, _ = h.EventsService.GetEventsForUser(r.Context(), userID, role, filter)
		} else {
			events, _ = h.EventsService.GetEvents(r.Context(), filter)
		}

		// empty list instead of var responseDto []dto.EventResponseDto, so that it would return empty list instead of nil if no events found
		responseDto := make([]dto.EventResponseDto, len(events))

		for i, retrievedEvent := range events {

			eventDto := dto.NewEventResponseDto(retrievedEvent, false)

			responseDto[i] = eventDto
		}

		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
		return
	}

	var schedules []values.ReadRecurrenceValues
	var err *errLib.CommonError

	if idErr == nil && roleErr == nil && (role == contextUtils.RoleAthlete || role == contextUtils.RoleCoach) {
		schedules, err = h.EventsService.GetEventsRecurrencesForUser(r.Context(), userID, role, filter)
	} else {
		schedules, err = h.EventsService.GetEventsRecurrences(r.Context(), filter)
	}
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// empty list instead of var responseDto []dto.EventResponseDto, so that it would return empty list instead of nil if no events found
	responseDto := []dto.RecurrenceResponseDto{}

	for _, schedule := range schedules {

		scheduleDto := dto.NewRecurrenceResponseDto(schedule)

		responseDto = append(responseDto, scheduleDto)
	}

	response := map[string]interface{}{
		"data":  responseDto,
		"page":  page,
		"limit": limit,
		"count": len(responseDto),
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)

}

// GetEvent retrieves detailed information about a specific event based on its ID.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID" Format(uuid) example(550e8400-e29b-41d4-a716-446655440000)
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

// CreateEvent creates new events given its recurrence information.
// @Tags events
// @Accept json
// @Produce json
// @Security Bearer
// @Param event body dto.EventRequestDto true "Event details"
// @Success 201 {object} map[string]interface{} "Event created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/one-time [post]
func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID, ctxErr := contextUtils.GetUserID(r.Context())

	if ctxErr != nil {
		responseHandlers.RespondWithError(w, ctxErr)
		return
	}

	var targetBody dto.EventRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if eventCreate, err := targetBody.ToCreateEventValues(userID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		if err = h.EventsService.CreateEvent(r.Context(), eventCreate); err != nil {
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
// @Param id path string true "Event ID" Format(uuid) example(550e8400-e29b-41d4-a716-446655440000)
// @Param event body dto.EventRequestDto true "Updated event details"
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

	var targetBody dto.EventRequestDto

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
