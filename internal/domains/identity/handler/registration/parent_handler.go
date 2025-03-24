package registration

import (
	"api/internal/di"
	dto "api/internal/domains/identity/dto/customer"
	service "api/internal/domains/identity/service/firebase"
	"api/internal/domains/identity/service/registration"
	identityUtils "api/internal/domains/identity/utils"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type ParentRegistrationHandlers struct {
	CustomerRegistrationService *registration.CustomerRegistrationService
	FirebaseService             *service.Service
	StaffRegistrationService    *registration.StaffsRegistrationService
}

func NewParentRegistrationHandlers(container *di.Container) *ParentRegistrationHandlers {

	return &ParentRegistrationHandlers{
		CustomerRegistrationService: registration.NewCustomerRegistrationService(container),
		StaffRegistrationService:    registration.NewStaffRegistrationService(container),
		FirebaseService:             service.NewFirebaseService(container),
	}
}

// RegisterParent registers a new parent.
// It verifies the Firebase token and uses the provided parent details to create a new parent account.
//
// @Summary Register a new parent
// @Description Registers a new parent by verifying the Firebase token and creating an account based on the provided details.
// @Tags registration
// @Accept json
// @Produce json
// @Param parent body dto.ParentRegistrationRequestDto true "Parent registration details" // The parent registration data, including name, age, email, phone, consent, etc.
// @Param Authorization header string true "Firebase token for user verification" // Firebase token required for verifying the user
// @Success 201 {object} map[string]interface{} "Parent registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Missing or invalid Firebase token or request body"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to register parent"
// @Router /register/parent [post]
func (h *ParentRegistrationHandlers) RegisterParent(w http.ResponseWriter, r *http.Request) {

	firebaseToken, err := identityUtils.GetFirebaseTokenFromAuthorizationHeader(r)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var requestDto dto.ParentRegistrationRequestDto

	if err = validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	email, err := h.FirebaseService.GetUserEmail(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	vo, parentErr := requestDto.ToParent(email)

	if parentErr != nil {
		responseHandlers.RespondWithError(w, parentErr)
		return
	}

	if userInfo, err := h.CustomerRegistrationService.RegisterParent(r.Context(), vo); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		responseHandlers.RespondWithSuccess(w, userInfo, http.StatusCreated)
	}
}
