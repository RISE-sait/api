package registration

import (
	"api/internal/di"
	dto "api/internal/domains/identity/dto/staff"
	firebaseService "api/internal/domains/identity/service/firebase"
	registrationService "api/internal/domains/identity/service/registration"
	identityUtils "api/internal/domains/identity/utils"

	"github.com/google/uuid"

	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type StaffHandlers struct {
	StaffRegistrationService *registrationService.StaffsRegistrationService
	FirebaseService          *firebaseService.Service
}

func NewStaffRegistrationHandlers(container *di.Container) *StaffHandlers {

	staffRegistrationService := registrationService.NewStaffRegistrationService(container)

	return &StaffHandlers{
		StaffRegistrationService: staffRegistrationService,
		FirebaseService:          firebaseService.NewFirebaseService(container),
	}
}

// RegisterStaff registers a new staff member account.
// @Summary Register a new staff member
// @Description Creates a new staff account in the system using the provided registration details.
// @Tags registration
// @Accept json
// @Produce json
// @Param Authorization header string true "Firebase token for user verification" // Firebase token in the Authorization header
// @Param staff body dto.RegistrationRequestDto true "Staff registration details"
// @Success 201 {object} map[string]interface{} "Staff registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid or missing authentication token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to register staff"
// @Router /register/staff [post]
func (h *StaffHandlers) RegisterStaff(w http.ResponseWriter, r *http.Request) {

	firebaseToken, err := identityUtils.GetFirebaseTokenFromAuthorizationHeader(r)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var requestDto dto.RegistrationRequestDto

	if err = validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	email, err := h.FirebaseService.GetUserEmail(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	valueObject, err := requestDto.ToCreateStaffValues(email)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	err = h.StaffRegistrationService.RegisterPendingStaff(r.Context(), valueObject)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// ApproveStaff approves a pending staff member.
// @Summary Approve a pending staff member
// @Description Approves a pending staff member's account in the system
// @Tags registration
// @Accept json
// @Produce json
// @Security Bearer
// @Param staff_id path string true "ID of staff member to approve"
// @Success 200 {object} map[string]interface{} "Staff approved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid or missing authentication token"
// @Failure 403 {object} map[string]interface{} "Forbidden: User does not have admin privileges"
// @Failure 404 {object} map[string]interface{} "Not Found: Staff member not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to approve staff"
// @Router /register/staff/approve/{id} [post]
func (h *StaffHandlers) ApproveStaff(w http.ResponseWriter, r *http.Request) {

	var staffID uuid.UUID

	if staffIdStr := chi.URLParam(r, "id"); staffIdStr == "" {
		responseHandlers.RespondWithError(w, errLib.New("staff ID is required", http.StatusBadRequest))
		return
	} else {
		id, err := validators.ParseUUID(staffIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		staffID = id
	}

	err := h.StaffRegistrationService.ApproveStaff(r.Context(), staffID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusOK)
}

// GetPendingStaff retrieves a pending staff member's details.
// @Summary Get pending staff member details
// @Tags registration
// @produce json
// @Security Bearer
// @Success 200 {object} dto.PendingStaffResponseDto "Pending staff member details"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to retrieve pending staff"
// @Router /register/staff/pending [get]
func (h *StaffHandlers) GetPendingStaffs(w http.ResponseWriter, r *http.Request) {
	staffs, err := h.StaffRegistrationService.GetPendingStaffs(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	result := make([]dto.PendingStaffResponseDto, len(staffs))
	for i, staff := range staffs {
		result[i] = dto.NewPendingStaffResponse(staff)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// DeletePendingStaff deletes or rejects a pending staff member.
// @Summary Delete/Reject a pending staff member
// @Description Deletes a pending staff member's application from the system
// @Tags registration
// @Accept json
// @Produce json
// @Security Bearer
// @Param staff_id path string true "ID of staff member to delete"
// @Success 200 {object} map[string]interface{} "Staff deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid or missing authentication token"
// @Failure 403 {object} map[string]interface{} "Forbidden: User does not have admin privileges"
// @Failure 404 {object} map[string]interface{} "Not Found: Pending staff member not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to delete staff"
// @Router /register/staff/reject/{id} [delete]
func (h *StaffHandlers) DeletePendingStaff(w http.ResponseWriter, r *http.Request) {
	var staffID uuid.UUID

	if staffIdStr := chi.URLParam(r, "id"); staffIdStr == "" {
		responseHandlers.RespondWithError(w, errLib.New("staff ID is required", http.StatusBadRequest))
		return
	} else {
		id, err := validators.ParseUUID(staffIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		staffID = id
	}

	err := h.StaffRegistrationService.DeletePendingStaff(r.Context(), staffID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusOK)
}
