package identity

import (
	"api/internal/di"
	identity "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type CreatePendingChildAccountController struct {
	ChildAccountRegistrationService *service.ChildAccountRequestService
}

func NewCreatePendingChildAccountController(container *di.Container) *CreatePendingChildAccountController {

	childAccountRegistrationService := service.NewChildAccountRegistrationRequestService(container)
	return &CreatePendingChildAccountController{
		ChildAccountRegistrationService: childAccountRegistrationService,
	}
}

func (c *CreatePendingChildAccountController) CreatePendingChildAccount(w http.ResponseWriter, r *http.Request) {

	var dto identity.CreatePendingChildAccountDto

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
	err = c.ChildAccountRegistrationService.CreatePendingAccount(r.Context(), valueObject)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
