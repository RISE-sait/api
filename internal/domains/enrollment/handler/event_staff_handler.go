package enrollment

import (
	"api/internal/di"
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

func NewEventStaffsHandler(container *di.Container) *StaffsHandler {
	return &StaffsHandler{Repo: repository.NewEventStaffsRepository(container)}
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

	eventId, err := parseUUIDParam(r, "event_id")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	staffId, err := parseUUIDParam(r, "staff_id")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.AssignStaffToEvent(r.Context(), eventId, staffId); err != nil {
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

	eventId, err := parseUUIDParam(r, "event_id")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	staffId, err := parseUUIDParam(r, "staff_id")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.UnassignedStaffFromEvent(r.Context(), eventId, staffId); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusOK)
}

func parseUUIDParam(r *http.Request, param string) (uuid.UUID, *errLib.CommonError) {
	idStr := chi.URLParam(r, param)
	if idStr == "" {
		return uuid.Nil, errLib.New(param+" must be provided", http.StatusBadRequest)
	}
	return validators.ParseUUID(idStr)
}
