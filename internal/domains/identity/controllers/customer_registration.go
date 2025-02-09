package identity

import (
	"api/internal/di"
	identity "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	"api/internal/domains/identity/values"
	lib "api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
	"time"
)

type CustomerRegistrationController struct {
	AccountRegistrationService  *service.AccountRegistrationService
	CustomerRegistrationService *service.CustomerAccountRegistrationService
}

func NewCustomerRegistrationController(container *di.Container) *CustomerRegistrationController {

	accountRegistrationService := service.NewAccountRegistrationService(container)
	return &CustomerRegistrationController{
		AccountRegistrationService:  accountRegistrationService,
		CustomerRegistrationService: service.NewCustomerAccountRegistrationService(container),
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

	accountRegistrationCredentials := values.RegisterCredentials{
		Email:    valueObject.Email,
		Password: valueObject.Password,
	}

	// Step 2: Call the service to create the account
	tx, _, err := c.AccountRegistrationService.CreateAccount(r.Context(), &accountRegistrationCredentials)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	// Step 3: Call the service to create the customer account
	userInfo, err := c.CustomerRegistrationService.CreateCustomer(r.Context(), tx, valueObject)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	// Step 4: Create JWT claims
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
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour), // Set expiration to 24 hours
	})

	// Step 5: Set Authorization header and respond
	w.Header().Set("Authorization", "Bearer "+signedToken)
	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
