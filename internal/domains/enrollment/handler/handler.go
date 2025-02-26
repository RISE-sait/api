package handler

import (
	"api/internal/domains/enrollment/dto"
	"api/internal/domains/enrollment/entity"
	"api/internal/domains/enrollment/service"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service enrollment_service.EnrollmentService
}

func NewHandler(service enrollment_service.EnrollmentService) *Handler {
	return &Handler{Service: service}
}

// CreateEnrollment creates a new enrollment.
// @Summary Create a new enrollment
// @Description Create a new enrollment
// @Tags enrollments
// @Accept json
// @Produce json
// @Param enrollment body dto.EnrollmentRequestDto true "Enrollment details"
// @Security Bearer
// @Success 201 {object} dto.EnrollmentResponse "Enrollment created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /enrollments [post]
func (h *Handler) CreateEnrollment(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.EnrollmentRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	enrollmentDetails, err := requestDto.ToCreateValueObjects()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	createdEnrollment, err := h.Service.EnrollCustomer(r.Context(), *enrollmentDetails)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := mapEntityToResponse(createdEnrollment)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// GetEnrollments retrieves enrollments.
// @Summary Get enrollments by customer and event HubSpotId
// @Description Get enrollments by customer and event HubSpotId
// @Tags enrollments
// @Accept json
// @Produce json
// @Param customerId path string true "Customer HubSpotId"
// @Param eventId path string true "Event HubSpotId"
// @Success 200 {array} dto.EnrollmentResponse "Enrollments retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
// @Failure 404 {object} map[string]interface{} "Not Found: Enrollments not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /enrollments/{customerId}/{eventId} [get]
func (h *Handler) GetEnrollments(w http.ResponseWriter, r *http.Request) {

	var customerId *uuid.UUID

	customerIdStr := chi.URLParam(r, "customerId")

	if customerIdStr != "" {
		id, err := validators.ParseUUID(customerIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		customerId = &id
	}

	var eventId *uuid.UUID

	eventIdStr := chi.URLParam(r, "eventId")

	if eventIdStr != "" {
		id, err := validators.ParseUUID(customerIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		eventId = &id
	}

	enrollments, err := h.Service.GetEnrollments(r.Context(), eventId, customerId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseData := make([]dto.EnrollmentResponse, len(enrollments))

	for i, enrollment := range enrollments {
		responseData[i] = mapEntityToResponse(&enrollment)
	}

	responseHandlers.RespondWithSuccess(w, responseData, http.StatusOK)
}

// DeleteEnrollment deletes an enrollment by HubSpotId.
// @Summary Delete an enrollment
// @Description Delete an enrollment by HubSpotId
// @Tags enrollments
// @Accept json
// @Produce json
// @Param id path string true "Enrollment HubSpotId"
// @Security Bearer
// @Success 204 "No Content: Enrollment deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
// @Failure 404 {object} map[string]interface{} "Not Found: Enrollment not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /enrollments/{id} [delete]
func (h *Handler) DeleteEnrollment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.UnEnrollCustomer(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func mapEntityToResponse(enrollment *entity.Enrollment) dto.EnrollmentResponse {
	return dto.EnrollmentResponse{
		ID:          enrollment.ID,
		CustomerID:  enrollment.CustomerID,
		EventID:     enrollment.EventID,
		CreatedAt:   enrollment.CreatedAt,
		UpdatedAt:   enrollment.UpdatedAt,
		CheckedInAt: enrollment.CheckedInAt,
		IsCancelled: enrollment.IsCancelled,
	}
}
