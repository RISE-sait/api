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
)

// EventsHandler provides HTTP handlers for managing events.
type EventsHandler struct {
	Repo repository.IEventsRepository
}

func NewEventsHandler(repo repository.IEventsRepository) *EventsHandler {
	return &EventsHandler{Repo: repo}
}

// GetEvents retrieves all events based on filter criteria.
// @Summary Get all events
// @Description Retrieve all events, with optional filters by course, location, and practice.
// @Tags events
// @Accept json
// @Produce json
// @Param courseId query string false "Filter by course ID (UUID)"
// @Param locationId query string false "Filter by location ID (UUID)"
// @Param practiceId query string false "Filter by practice ID (UUID)"
// @Success 200 {array} event.ResponseDto "GetMemberships of events retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events [get]
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {

	courseIdStr := r.URL.Query().Get("courseId")
	locationIdStr := r.URL.Query().Get("locationId")
	practiceIdStr := r.URL.Query().Get("practiceId")
	gameIdStr := r.URL.Query().Get("gameId")

	var courseId, locationId, practiceId, gameId uuid.UUID

	if courseIdStr != "" {
		id, err := validators.ParseUUID(courseIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("Invalid course ID", http.StatusBadRequest))
			return
		}

		courseId = id
	}

	if locationIdStr != "" {
		id, err := validators.ParseUUID(locationIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("Invalid location ID", http.StatusBadRequest))
			return
		}

		locationId = id
	}

	if practiceIdStr != "" {
		id, err := validators.ParseUUID(practiceIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("Invalid practice ID", http.StatusBadRequest))
			return
		}

		practiceId = id
	}

	if gameIdStr != "" {
		id, err := validators.ParseUUID(gameIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("Invalid game ID", http.StatusBadRequest))
			return
		}

		gameId = id
	}

	events, err := h.Repo.GetEvents(r.Context(), courseId, locationId, practiceId, gameId)

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
// @Description Retrieves details of a specific event based on its HubSpotId.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} event.ResponseDto "Event details retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
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

// UpdateEvent updates an existing event by HubSpotId.
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

// DeleteEvent deletes an event by HubSpotId.
// @Summary Delete an event
// @Description Deletes an event by its HubSpotId.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 204 {object} map[string]interface{} "No Content: Event deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
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
