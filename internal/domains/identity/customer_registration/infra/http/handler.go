package customer_registration

import (
	"api/internal/domains/identity/customer_registration"
	"api/internal/domains/identity/customer_registration/infra/http/dto"
	"api/internal/domains/identity/customer_registration/values"
	lib "api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type Handler struct {
	AccountRegistrationService *customer_registration.AccountRegistrationService
}

func NewHandler(accountRegistrationService *customer_registration.AccountRegistrationService) *Handler {
	return &Handler{
		AccountRegistrationService: accountRegistrationService,
	}
}

func (c *Handler) CreateTraditionalAccount(w http.ResponseWriter, r *http.Request) {
	var dto dto.CreateUserRequest

	// Step 1: Decode and validate the request body.
	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	userPasswordCreate := values.NewUserPasswordCreate(dto.Email, dto.Password)
	waiverCreate := values.NewWaiverCreate(dto.Email, dto.WaiverUrl, dto.IsSignedWaiver)

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
