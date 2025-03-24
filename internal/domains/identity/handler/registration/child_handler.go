package registration

import (
	"api/internal/di"
	commonDto "api/internal/domains/identity/dto/common"
	dto "api/internal/domains/identity/dto/customer"
	service "api/internal/domains/identity/service/firebase"
	"api/internal/domains/identity/service/registration"
	identityUtils "api/internal/domains/identity/utils"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type ChildRegistrationHandlers struct {
	ChildAccountRegistrationService *registration.ChildRegistrationService
	FirebaseService                 *service.Service
}

func NewChildRegistrationHandlers(container *di.Container) *ChildRegistrationHandlers {

	return &ChildRegistrationHandlers{
		ChildAccountRegistrationService: registration.NewChildAccountRegistrationService(container),
		FirebaseService:                 service.NewFirebaseService(container),
	}
}

// RegisterChild registers a new child account under a parent's account.
// @Summary Register a new child account and associate it with the parent
// @Description Registers a new child account using the provided details and associates it with the parent based on the Firebase authentication token.
// @Tags registration
// @Accept json
// @Produce json
// @Param customer body customer.ChildRegistrationRequestDto true "Child account registration details" // BaseDetails for child account registration
// @Param Authorization header string true "Firebase token for user verification" // Firebase token in the Authorization header
// @Success 201 {object} map[string]interface{} "Child account registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or missing Firebase token"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid Firebase token or insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to register child account or associate with parent"
// @Router /register/child [post]
func (h *ChildRegistrationHandlers) RegisterChild(w http.ResponseWriter, r *http.Request) {

	firebaseToken, err := identityUtils.GetFirebaseTokenFromAuthorizationHeader(r)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var requestDto dto.ChildRegistrationRequestDto

	if err = validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	parentEmail, err := h.FirebaseService.GetUserEmail(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	valueObject, err := requestDto.ToCreateChildValueObject(parentEmail)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Step 3: Call the service to create the customer account
	if userInfo, err := h.ChildAccountRegistrationService.CreateChildAccount(r.Context(), valueObject); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		responseBody := commonDto.UserAuthenticationResponseDto{
			FirstName:   userInfo.FirstName,
			LastName:    userInfo.LastName,
			Email:       userInfo.Email,
			Role:        userInfo.Role,
			Phone:       userInfo.Phone,
			Age:         userInfo.Age,
			CountryCode: userInfo.CountryCode,
		}

		responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
	}

}
