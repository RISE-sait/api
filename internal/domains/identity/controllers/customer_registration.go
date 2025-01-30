package identity

import (
	"api/cmd/server/di"
	identity "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	errLib "api/internal/libs/errors"
	lib "api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"encoding/json"
	"io"
	"net/http"
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

	body, ioErr := io.ReadAll(r.Body)

	if ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Failed to read request body", http.StatusBadRequest))
		return
	}

	var credentials identity.Credentials
	var customerDto identity.CustomerWaiverCreateDto

	if ioErr := json.Unmarshal(body, &credentials); ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid JSON format for credentials", http.StatusBadRequest))
		return
	}

	if ioErr := json.Unmarshal(body, &customerDto); ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid JSON format for customer data", http.StatusBadRequest))
		return
	}

	userPasswordCreate := identity.NewCredentials(credentials.Email, credentials.Password)
	waiverCreate := identity.NewCustomerWaiverCreateDto(customerDto.WaiverUrl, customerDto.IsWaiverSigned)

	// Step 2: Call the service to create the account
	userInfo, err := c.AccountRegistrationService.CreateCustomer(r.Context(), userPasswordCreate, waiverCreate)
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

	// Step 4: Set Authorization header and respond
	w.Header().Set("Authorization", "Bearer "+signedToken)
	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
