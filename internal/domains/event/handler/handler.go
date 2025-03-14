package event

import (
	dto "api/internal/domains/event/dto"
	repository "api/internal/domains/event/persistence/repository"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// EventsHandler provides HTTP handlers for managing events.
type EventsHandler struct {
	Repo *repository.Repository
}

func NewEventsHandler(repo *repository.Repository) *EventsHandler {
	return &EventsHandler{Repo: repo}
}

// GetEvents retrieves all events based on filter criteria.
// @Summary Get all events
// @Description Retrieve all events within a specific date range, with optional filters by course, location, game, and practice.
// @Tags events
// @Param after query string true "Retrieve events after this date (format: YYYY-MM-DD)" example("2024-05-01")
// @Param before query string true "Retrieve events before this date (format: YYYY-MM-DD)" example("2024-06-01")
// @Param game_id query string false "Filter by game ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param course_id query string false "Filter by course ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param practice_id query string false "Filter by practice ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param location_id query string false "Filter by location ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Accept json
// @Produce json
// @Success 200 {array} event.ResponseDto "List of events retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input format or missing required parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events [get]
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	var after, before time.Time

	var courseID, gameID, locationID, practiceID uuid.UUID

	if afterStr := query.Get("after"); afterStr != "" {
		afterDate, err := time.Parse("2006-01-02", afterStr)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'after' date format, expected YYYY-MM-DD", http.StatusBadRequest))
			return
		}
		after = afterDate
	} else {
		responseHandlers.RespondWithError(w, errLib.New("'after' date is required", http.StatusBadRequest))
		return
	}

	if beforeStr := query.Get("before"); beforeStr != "" {
		beforeDate, err := time.Parse("2006-01-02", beforeStr)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'before' date format, expected YYYY-MM-DD", http.StatusBadRequest))
			return
		}
		before = beforeDate
	} else {
		responseHandlers.RespondWithError(w, errLib.New("'before' date is required", http.StatusBadRequest))
		return
	}

	if gameIDStr := query.Get("game_id"); gameIDStr != "" {
		id, err := validators.ParseUUID(gameIDStr)

		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'game_id' format, expected uuid format", http.StatusBadRequest))
			return
		}

		gameID = id
	}

	if courseIDStr := query.Get("course_id"); courseIDStr != "" {
		id, err := validators.ParseUUID(courseIDStr)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'course_id' format, expected uuid format", http.StatusBadRequest))
			return
		}
		courseID = id
	}

	if practiceIDStr := query.Get("practice_id"); practiceIDStr != "" {
		id, err := validators.ParseUUID(practiceIDStr)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid practice_id format, expected uuid format", http.StatusBadRequest))
			return
		}
		practiceID = id
	}

	if locationIDStr := query.Get("location_id"); locationIDStr != "" {
		id, err := validators.ParseUUID(locationIDStr)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'location_id' format, expected uuid format", http.StatusBadRequest))
			return
		}
		locationID = id
	}

	events, err := h.Repo.GetEvents(r.Context(), after, before, courseID, practiceID, gameID, locationID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.ResponseDto, len(events))

	for i, event := range events {
		result[i] = dto.NewEventResponse(event)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// GetEvent retrieves detailed information about a specific event.
// @Summary Get event details
// @Description Retrieves details of a specific event based on its ID.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} event.ResponseDto "Event details retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{id} [get]
func (h *EventsHandler) GetEvent(w http.ResponseWriter, r *http.Request) {

	eventIdStr := chi.URLParam(r, "id")

	var eventId uuid.UUID

	if eventIdStr != "" {
		id, err := validators.ParseUUID(eventIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		eventId = id
	}

	event, err := h.Repo.GetEvent(r.Context(), eventId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.NewEventResponse(event)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// CreateEvent creates a new event.
// @Summary Create a new event
// @Description Registers a new event with the provided details.
// @Tags events
// @Accept json
// @Produce json
// @Param event body dto.RequestDto true "Event details"
// @Success 201 {object} map[string]interface{} "Event created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events [post]
func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.RequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	eventCreate, err := targetBody.ToCreateEventValues()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	createdEvent, err := h.Repo.CreateEvent(r.Context(), eventCreate)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.NewEventResponse(createdEvent)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// UpdateEvent updates an existing event by ID.
// @Summary Update an event
// @Description Updates the details of an existing event.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param event body dto.RequestDto true "Updated event details"
// @Success 204 {object} map[string]interface{} "No Content: Event updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{id} [put]
func (h *EventsHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var targetBody dto.RequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	params, err := targetBody.ToUpdateEventValues(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	event, err := h.Repo.UpdateEvent(r.Context(), params)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.NewEventResponse(event)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusNoContent)
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
	}

	if err = h.Repo.DeleteEvent(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
