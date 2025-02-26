package registration

import (
	"api/internal/di"
	"api/internal/domains/identity/dto/customer"
	service "api/internal/domains/identity/service/firebase"
	"api/internal/domains/identity/service/registration"
	errLib "api/internal/libs/errors"
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
// @Param customer body customer.RegistrationDto true "Child account registration details" // Details for child account registration
// @Param firebase_token header string true "Firebase token for user verification" // Firebase token in the Authorization header
// @Success 201 {object} map[string]interface{} "Child account registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or missing Firebase token"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid Firebase token or insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to register child account or associate with parent"
// @Router /register/child [post]
func (h *ChildRegistrationHandlers) RegisterChild(w http.ResponseWriter, r *http.Request) {

	firebaseToken := r.Header.Get("firebase_token")

	if firebaseToken == "" {
		responseHandlers.RespondWithError(w, errLib.New("Missing Firebase token", http.StatusBadRequest))
		return
	}

	var dto customer.RegistrationDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	parent, err := h.FirebaseService.GetUserInfo(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
	}

	valueObject, err := dto.ToCreateChildValueObject(parent.Email)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Step 3: Call the service to create the customer account
	err = h.ChildAccountRegistrationService.CreateChildAccount(r.Context(), valueObject)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
