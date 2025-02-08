package identity

import (
	"api/internal/di"
	identity "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	lib "api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
	"time"
)

type CustomerRegistrationController struct {
	AccountRegistrationService *service.AccountRegistrationService
}

func NewCustomerRegistrationController(container *di.Container) *CustomerRegistrationController {

	accountRegistrationService := service.NewAccountRegistrationService(container)
	return &CustomerRegistrationController{
		AccountRegistrationService: accountRegistrationService,
	}
}

func (c *CustomerRegistrationController) CreateCustomer(w http.ResponseWriter, r *http.Request) {

	var dto identity.CustomerRegistrationDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	valueObject, err := dto.ToValueObjects()

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	// Step 2: Call the service to create the account
	userInfo, err := c.AccountRegistrationService.CreateCustomer(r.Context(), valueObject)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	// Step 3: Create JWT claims
	signedToken, err := lib.SignJWT(*userInfo)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwtToken",
		Value:    signedToken,
		Path:     "/",
		HttpOnly: true, // Prevent JavaScript access
		Secure:   true, // Use HTTPS in production
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(24 * time.Hour), // Set expiration to 24 hours
	})

	// Step 4: Set Authorization header and respond
	w.Header().Set("Authorization", "Bearer "+signedToken)
	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
