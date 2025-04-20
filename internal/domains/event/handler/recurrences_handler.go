package event

import (
	dto "api/internal/domains/event/dto"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"
	"net/http"

	"github.com/go-chi/chi"
)

// CreateRecurrences creates new events given its recurrence information.
// @Tags events
// @Accept json
// @Produce json
// @Security Bearer
// @Param event body dto.RecurrenceRequestDto true "Event details"
// @Success 201 {object} map[string]interface{} "Event created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/recurring [post]
func (h *EventsHandler) CreateRecurrences(w http.ResponseWriter, r *http.Request) {

	userID, ctxErr := contextUtils.GetUserID(r.Context())

	if ctxErr != nil {
		responseHandlers.RespondWithError(w, ctxErr)
		return
	}

	var targetBody dto.RecurrenceRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if recurrenceValues, err := targetBody.ToCreateRecurrenceValues(userID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		if err = h.EventsService.CreateEvents(r.Context(), recurrenceValues); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// UpdateRecurrences updates existing events by filters.
// @Tags events
// @Accept json
// @Produce json
// @Security Bearer
// @Param event body dto.RecurrenceRequestDto true "Update events details"
// @Success 204 {object} map[string]interface{} "No Content: Events updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Events not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/recurring/{id} [put]
func (h *EventsHandler) UpdateRecurrences(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	recurrenceID, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userID, err := contextUtils.GetUserID(r.Context())

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var targetBody dto.RecurrenceRequestDto

	if err = validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Convert to domain values

	params, err := targetBody.ToUpdateRecurrenceValues(userID, recurrenceID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.EventsService.UpdateRecurringEvents(r.Context(), params); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
