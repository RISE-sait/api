package haircut

import (
	dto "api/internal/domains/haircut/dto"
	repository "api/internal/domains/haircut/persistence/repository"
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

// GetEvents retrieves all barber events based on filter criteria.
// @Summary Get all barber events
// @Description Retrieve all barber events, with optional filters by barber ID and customer ID.
// @Tags haircut
// @Accept json
// @Produce json
// @Param after query string false "Start date of the events range (YYYY-MM-DD)" example("2025-03-01")
// @Param before query string false "End date of the events range (YYYY-MM-DD)" example("2025-03-31")
// @Param barber_id query string false "Filter by barber ID"
// @Param customer_id query string false "Filter by customer ID"
// @Success 200 {array} haircut.EventResponseDto "List of barber events retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events [get]
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	var (
		barberID, customerID uuid.UUID
		before, after        time.Time
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

	if barberIdStr := query.Get("barber_id"); barberIdStr != "" {
		if id, err := validators.ParseUUID(barberIdStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			barberID = id
		}
	}

	if customerIdStr := query.Get("customer_id"); customerIdStr != "" {
		if id, err := validators.ParseUUID(customerIdStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			customerID = id
		}
	}

	if (after.IsZero() || before.IsZero()) && (barberID == uuid.Nil && customerID == uuid.Nil) {
		responseHandlers.RespondWithError(w, errLib.New("at least one of (before and after) or one of (barber_id, customer_id) must be provided", http.StatusBadRequest))
		return
	}

	events, err := h.Repo.GetEvents(r.Context(), barberID, customerID, before, after)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.EventResponseDto, len(events))

	for i, event := range events {
		result[i] = dto.NewEventResponse(event)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// CreateEvent creates a new barber event.
// @Summary Create a new barber event
// @Description Registers a new barber event with the provided details.
// @Tags haircut
// @Accept json
// @Produce json
// @Param event body dto.RequestDto true "Barber event details"
// @Success 201 {object} haircut.EventResponseDto "Barber event created successfully"
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
// @Tags haircut
// @Accept json
// @Produce json
// @Param id path string true "Barber event ID"
// @Param event body dto.RequestDto true "Updated barber event details"
// @Success 200 {object} haircut.EventResponseDto "Barber event updated successfully"
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
// @Tags haircut
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
// @Tags haircut
// @Accept json
// @Produce json
// @Param id path string true "Barber event ID"
// @Success 200 {object} haircut.EventResponseDto "Barber event details retrieved successfully"
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

	event, err := h.Repo.GetEvent(r.Context(), eventId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.NewEventResponse(event)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusOK)
}
