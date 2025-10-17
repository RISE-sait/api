package enrollment

import (
	"net/http"

	"api/internal/di"
	enrollmentService "api/internal/domains/enrollment/service"
	responseHandlers "api/internal/libs/responses"
)

type CustomerEnrollmentHandler struct {
	Service *enrollmentService.CustomerEnrollmentService
}

func NewCustomerEnrollmentHandler(container *di.Container) *CustomerEnrollmentHandler {
	return &CustomerEnrollmentHandler{Service: enrollmentService.NewCustomerEnrollmentService(container)}
}

// RemoveCustomerFromEvent removes a customer completely from an event.
// @Summary Remove a customer from an event
// @Description Completely removes a customer's enrollment from an event (deletes the record).
// @Tags event_enrollment
// @Accept json
// @Produce json
// @Param event_id path string true "Event ID" Format(uuid)
// @Param customer_id path string true "Customer ID" Format(uuid)
// @Success 200 {object} map[string]interface{} "Customer successfully removed from event"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Enrollment not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /events/{event_id}/customers/{customer_id} [delete]
func (h *CustomerEnrollmentHandler) RemoveCustomerFromEvent(w http.ResponseWriter, r *http.Request) {

	eventId, err := parseUUIDParam(r, "event_id")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	customerId, err := parseUUIDParam(r, "customer_id")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.RemoveCustomerFromEvent(r.Context(), eventId, customerId); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusOK)
}
