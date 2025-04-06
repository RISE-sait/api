package enrollment

import (
	repository "api/internal/domains/enrollment/persistence/repository"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
)

type StaffsHandler struct {
	Repo *repository.StaffsRepository
}

func NewEventStaffsHandler(repo *repository.StaffsRepository) *StaffsHandler {
	return &StaffsHandler{Repo: repo}
}

// AssignStaffToEvent assigns a staff member to an event.
// @Summary Assign a staff member to an event
// @Description Assign a staff member to an event using event_id and staff_id in the request body.
// @Tags event_staff
// @Accept json
// @Produce json
// @Param event_id path string true "Event ID"
// @Param staff_id path string true "Staff ID"
// @Success 200 {object} map[string]interface{} "Staff successfully assigned to event"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{event_id}/staffs/{staff_id} [post]
func (h *StaffsHandler) AssignStaffToEvent(w http.ResponseWriter, r *http.Request) {

	var eventId, staffId uuid.UUID

	if idStr := chi.URLParam(r, "event_id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			eventId = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("event_id must be provided", http.StatusBadRequest))
		return
	}

	if idStr := chi.URLParam(r, "staff_id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			staffId = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("staff_id must be provided", http.StatusBadRequest))
		return
	}

	if err := h.Repo.AssignStaffToEvent(r.Context(), eventId, staffId); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusOK)
}

// UnassignStaffFromEvent removes a staff member from an event.
// @Summary Unassign a staff member from an event
// @Description Remove a staff member from an event using event_id and staff_id in the request body.
// @Tags event_staff
// @Accept json
// @Produce json
// @Param event_id path string true "Event ID"
// @Param staff_id path string true "Staff ID"
// @Success 200 {object} map[string]interface{} "Staff successfully unassigned from event"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{event_id}/staffs/{staff_id} [delete]
func (h *StaffsHandler) UnassignStaffFromEvent(w http.ResponseWriter, r *http.Request) {

	var eventId, staffId uuid.UUID

	if idStr := chi.URLParam(r, "event_id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			eventId = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("event_id must be provided", http.StatusBadRequest))
		return
	}

	if idStr := chi.URLParam(r, "staff_id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			staffId = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("staff_id must be provided", http.StatusBadRequest))
		return
	}

	if err := h.Repo.UnassignedStaffFromEvent(r.Context(), eventId, staffId); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusOK)
}
