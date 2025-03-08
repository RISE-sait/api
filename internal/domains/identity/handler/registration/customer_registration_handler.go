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

type CustomerRegistrationHandlers struct {
	CustomerRegistrationService *registration.CustomerRegistrationService
	FirebaseService             *service.Service
}

func NewCustomerRegistrationHandlers(container *di.Container) *CustomerRegistrationHandlers {

	return &CustomerRegistrationHandlers{
		CustomerRegistrationService: registration.NewCustomerRegistrationService(container),
		FirebaseService:             service.NewFirebaseService(container),
	}
}

// RegisterCustomer registers a new customer.
// @Summary Register a new customer
// @Description Registers a new customer using the provided details, creates a customer account. The Firebase token is used for user verification.
// @Tags registration
// @Accept json
// @Produce json
// @Param customer body customer.RegistrationRequestDto true "Customer registration details" // Details for customer registration
// @Param firebase_token header string true "Firebase token for user verification" // Firebase token in the Authorization header
// @Success 201 {object} map[string]interface{} "Customer registered and JWT token issued successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or missing Firebase token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to register customer or create JWT token"
// @Router /register/customer [post]
func (h *CustomerRegistrationHandlers) RegisterCustomer(w http.ResponseWriter, r *http.Request) {

	firebaseToken := r.Header.Get("firebase_token")

	if firebaseToken == "" {
		responseHandlers.RespondWithError(w, errLib.New("Missing Firebase token", http.StatusBadRequest))
		return
	}

	var dto customer.RegistrationRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	email, err := h.FirebaseService.GetUserEmail(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	valueObject, err := dto.ToCreateRegularCustomerValueObject(email)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Step 3: Call the service to create the customer account
	if err = h.CustomerRegistrationService.RegisterCustomer(r.Context(), valueObject); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
