package enrollment

import (
	"api/internal/domains/enrollment/dto"
	"api/internal/domains/enrollment/entity"
	service "api/internal/domains/enrollment/service"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service *service.Service
}

func NewHandler(service *service.Service) *Handler {
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
// @Param customerId query string false "Customer ID"
// @Param eventId query string false "Event ID"
// @Success 200 {array} dto.EnrollmentResponse "Enrollments retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
// @Failure 404 {object} map[string]interface{} "Not Found: Enrollments not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /enrollments [get]
func (h *Handler) GetEnrollments(w http.ResponseWriter, r *http.Request) {

	var customerId, eventId uuid.UUID

	customerIdStr := r.URL.Query().Get("customerId")

	if customerIdStr != "" {
		id, err := validators.ParseUUID(customerIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		customerId = id
	}

	eventIdStr := r.URL.Query().Get("eventId")

	if eventIdStr != "" {
		id, err := validators.ParseUUID(eventIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		eventId = id
	}

	if eventId == uuid.Nil && customerId == uuid.Nil {
		err := errLib.New("either customerId or eventId must be provided", http.StatusBadRequest)
		responseHandlers.RespondWithError(w, err)
		return
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
