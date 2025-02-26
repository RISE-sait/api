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
	"time"
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

// RegisterCustomer registers a new customer and creates a JWT token.
// @Summary Register a new customer and create JWT token
// @Description Registers a new customer using the provided details, creates a customer account, and returns a JWT token for authentication. The Firebase token is used for user verification.
// @Tags registration
// @Accept json
// @Produce json
// @Param customer body customer.RegistrationDto true "Customer registration details" // Details for customer registration
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

	var dto customer.RegistrationDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	user, err := h.FirebaseService.GetUserInfo(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
	}

	valueObject, err := dto.ToCreateRegularCustomerValueObject(user.Email)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Step 3: Call the service to create the customer account
	jwtToken, err := h.CustomerRegistrationService.RegisterCustomer(r.Context(), valueObject)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    *jwtToken,
		Path:     "/",
		HttpOnly: true,  // Prevent JavaScript access
		Secure:   false, // Use HTTPS in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour), // Set expiration to 24 hours
	})

	// Step 5: Set Authorization header and respond
	w.Header().Set("Authorization", "Bearer "+*jwtToken)
	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
