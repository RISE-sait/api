package event_staff

import (
	eventStaffDto "api/internal/domains/event_staff/dto"
	repository "api/internal/domains/event_staff/persistence/repository"
	staffDto "api/internal/domains/staff/dto"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
)

// EventStaffsHandler provides HTTP handlers for managing events.
type EventStaffsHandler struct {
	Repo repository.EventStaffsRepositoryInterface
}

func NewEventStaffsHandler(repo repository.EventStaffsRepositoryInterface) *EventStaffsHandler {
	return &EventStaffsHandler{Repo: repo}
}

// AssignStaffToEvent assigns a staff member to an event.
// @Summary Assign a staff member to an event
// @Description Assign a staff member to an event using event_id and staff_id in the request body.
// @Tags event_staff
// @Accept json
// @Produce json
// @Param request body eventStaffDto.RequestDto true "Event and staff assignment details"
// @Success 200 {object} map[string]interface{} "Staff successfully assigned to event"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /event-staff [post]
func (h *EventStaffsHandler) AssignStaffToEvent(w http.ResponseWriter, r *http.Request) {

	var targetBody eventStaffDto.RequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	details := targetBody.ToDetails()

	err := h.Repo.AssignStaffToEvent(r.Context(), details)

	if err != nil {
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
// @Param event_id query string true "Event ID (UUID)"
// @Success 200 {array} staff.ResponseDto "GetMemberships of staff assigned to the event"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /event-staff [get]
func (h *EventStaffsHandler) GetStaffsAssignedToEvent(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var id uuid.UUID

	if idStr != "" {
		eventId, err := validators.ParseUUID(idStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
		}

		id = eventId
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
// @Param request body eventStaffDto.RequestDto true "Event and staff unassignment details"
// @Success 200 {object} map[string]interface{} "Staff successfully unassigned from event"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /event-staff [delete]
func (h *EventStaffsHandler) UnassignStaffFromEvent(w http.ResponseWriter, r *http.Request) {

	var targetBody eventStaffDto.RequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	details := targetBody.ToDetails()

	err := h.Repo.UnassignedStaffFromEvent(r.Context(), details)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusOK)
}
