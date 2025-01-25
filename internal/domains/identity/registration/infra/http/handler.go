package registration

import (
	"api/internal/domains/identity/lib"
	"api/internal/domains/identity/registration"
	"api/internal/domains/identity/registration/infra/http/dto"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type AccountRegistrationController struct {
	AccountRegistrationService *registration.AccountRegistrationService
}

func NewAccountRegistrationController(accountRegistrationService *registration.AccountRegistrationService) *AccountRegistrationController {
	return &AccountRegistrationController{
		AccountRegistrationService: accountRegistrationService,
	}
}

func (c *AccountRegistrationController) CreateTraditionalAccount(w http.ResponseWriter, r *http.Request) {
	var dto dto.CreateUserRequest

	// Step 1: Decode and validate the request body.
	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	// Step 2: Call the service to create the account
	userInfo, err := c.AccountRegistrationService.CreateTraditionalAccount(r.Context(), dto.Email, dto.Password, dto.Role, dto.IsActiveStaff)
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
