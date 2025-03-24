package event

import (
	service "api/internal/domains/event/service"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"net/http"

	"github.com/go-chi/chi"
)

type CustomerEnrollmentHandler struct {
	Service *service.CustomerEnrollmentService
}

func NewCustomerEnrollmentHandler(service *service.CustomerEnrollmentService) *CustomerEnrollmentHandler {
	return &CustomerEnrollmentHandler{Service: service}
}

// EnrollCustomer creates a new enrollment.
// @Tags enrollments
// @Accept json
// @Produce json
// @Security Bearer
// @Param event_id path string true "Event ID"
// @Param customer_id path string true "Customer ID"
// @Success 201 "Enrollment created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{event_id}/customers/{customer_id} [post]
func (h *CustomerEnrollmentHandler) EnrollCustomer(w http.ResponseWriter, r *http.Request) {

	var eventId, customerId uuid.UUID

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

	if idStr := chi.URLParam(r, "customer_id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			customerId = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("customer_id must be provided", http.StatusBadRequest))
		return
	}

	if err := h.Service.EnrollCustomer(r.Context(), eventId, customerId); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// UnenrollCustomer deletes an enrollment by ID.
// @Summary Delete an enrollment
// @Description Delete an enrollment by ID
// @Tags enrollments
// @Accept json
// @Produce json
// @Param id path string true "Enrollment ID"
// @Security Bearer
// @Success 204 "No Content: Enrollment deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Enrollment not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{event_id}/customers/{customer_id} [delete]
func (h *CustomerEnrollmentHandler) UnenrollCustomer(w http.ResponseWriter, r *http.Request) {

	var eventId, customerId uuid.UUID

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

	if idStr := chi.URLParam(r, "customer_id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			customerId = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("customer_id must be provided", http.StatusBadRequest))
		return
	}

	if err := h.Service.UnEnrollCustomer(r.Context(), eventId, customerId); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
