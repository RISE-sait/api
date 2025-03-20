package event

import (
	repository "api/internal/domains/event/persistence/repository"
	staffDto "api/internal/domains/user/dto/staff"
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

// GetStaffsAssignedToEvent retrieves the list of staff assigned to a specific event.
// @Summary Get staff assigned to an event
// @Description Retrieve all staff assigned to an event using event_id as a query parameter.
// @Tags event_staff
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {array} staff.ResponseDto "GetMemberships of staff assigned to the event"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{event_id}/staffs [get]
func (h *StaffsHandler) GetStaffsAssignedToEvent(w http.ResponseWriter, r *http.Request) {

	var id uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		eventId, err := validators.ParseUUID(idStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		id = eventId
	} else {
		responseHandlers.RespondWithError(w, errLib.New("id must be provided", http.StatusBadRequest))
		return
	}

	staffs, err := h.Repo.GetStaffsAssignedToEvent(r.Context(), id)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := make([]staffDto.ResponseDto, len(staffs))

	for i, staff := range staffs {
		responseBody[i] = staffDto.NewStaffResponse(staff)
	}

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
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
