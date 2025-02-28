package haircut_event

import (
	dto "api/internal/domains/haircut/dto"
	repository "api/internal/domains/haircut/persistence/repository/event"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
)

// EventsHandler provides HTTP handlers for managing events.
type EventsHandler struct {
	Repo repository.IBarberEventsRepository
}

func NewEventsHandler(repo repository.IBarberEventsRepository) *EventsHandler {
	return &EventsHandler{Repo: repo}
}

// GetEvents retrieves all barber events based on filter criteria.
// @Summary Get all barber events
// @Description Retrieve all barber events, with optional filters by barber ID and customer ID.
// @Tags barber_events
// @Accept json
// @Produce json
// @Param barber_id query string false "Filter by barber ID"
// @Param customer_id query string false "Filter by customer ID"
// @Param begin_date_time query string false "Filter by start date (ISO 8601 format)"
// @Param end_date_time query string false "Filter by end date (ISO 8601 format)"
// @Success 200 {array} haircut.ResponseDto "List of barber events retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events [get]
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {

	events, err := h.Repo.GetEvents(r.Context())

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

// CreateEvent creates a new barber event.
// @Summary Create a new barber event
// @Description Registers a new barber event with the provided details.
// @Tags barber_events
// @Accept json
// @Produce json
// @Param event body dto.RequestDto true "Barber event details"
// @Success 201 {object} haircut.ResponseDto "Barber event created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events [post]
func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.RequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	eventCreate, err := targetBody.ToCreateEventValue()

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

// UpdateEvent updates an existing barber event by ID.
// @Summary Update a barber event
// @Description Updates the details of an existing barber event.
// @Tags barber_events
// @Accept json
// @Produce json
// @Param id path string true "Barber event ID"
// @Param event body dto.RequestDto true "Updated barber event details"
// @Success 200 {object} haircut.ResponseDto "Barber event updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Barber event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events/{id} [put]
func (h *EventsHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var targetBody dto.RequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	params, err := targetBody.ToUpdateEventValue(idStr)

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

// DeleteEvent deletes a barber event by ID.
// @Summary Delete a barber event
// @Description Deletes a barber event by its ID.
// @Tags barber_events
// @Accept json
// @Produce json
// @Param id path string true "Barber event ID"
// @Success 204 "No Content: Barber event deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Barber event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events/{id} [delete]
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

// GetEventDetails retrieves detailed information about a specific barber event.
// @Summary Get barber event details
// @Description Retrieves details of a specific barber event based on its ID.
// @Tags barber_events
// @Accept json
// @Produce json
// @Param id path string true "Barber event ID"
// @Success 200 {object} haircut.ResponseDto "Barber event details retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Barber event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events/{id} [get]
func (h *EventsHandler) GetEventDetails(w http.ResponseWriter, r *http.Request) {

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

	event, err := h.Repo.GetEventDetails(r.Context(), eventId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.NewEventResponse(event)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusOK)
}
