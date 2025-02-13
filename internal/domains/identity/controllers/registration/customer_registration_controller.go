package registration

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
	AccountRegistrationService  *service.AccountCreationService
	PasswordService             *service.UserOptionalInfoService
	CustomerRegistrationService *service.CustomerRegistrationService
}

func NewCustomerRegistrationController(container *di.Container) *CustomerRegistrationController {

	accountRegistrationService := service.NewAccountCreationService(container)

	return &CustomerRegistrationController{
		AccountRegistrationService:  accountRegistrationService,
		PasswordService:             service.NewUserOptionalInfoService(container),
		CustomerRegistrationService: service.NewCustomerRegistrationService(container),
	}
}

// CreateCustomer creates a new customer account.
// @Summary Create a new customer account
// @Description Registers a new customer with provided details and creates JWT authentication token
// @Tags registration
// @Accept json
// @Produce json
// @Param customer body identity.CustomerRegistrationDto true "Customer registration details"
// @Success 201 {object} map[string]interface{} "Customer registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /register/customer [post]
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

	// Step 3: Call the service to create the customer account
	userInfo, err := c.CustomerRegistrationService.RegisterCustomer(r.Context(), valueObject)

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
		Name:     "access_token",
		Value:    signedToken,
		Path:     "/",
		HttpOnly: true, // Prevent JavaScript access
		Secure:   true, // Use HTTPS in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour), // Set expiration to 24 hours
	})

	// Step 5: Set Authorization header and respond
	w.Header().Set("Authorization", "Bearer "+signedToken)
	response_handlers.RespondWithSuccess(w, dto.UserInfoDto, http.StatusCreated)
}
