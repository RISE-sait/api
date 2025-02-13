package registration

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
	AccontRegistrationService       *service.AccountCreationService
}

func NewCreatePendingChildAccountController(container *di.Container) *CreatePendingChildAccountController {

	return &CreatePendingChildAccountController{
		ChildAccountRegistrationService: service.NewChildAccountRegistrationRequestService(container),
		AccontRegistrationService:       service.NewAccountCreationService(container),
	}
}

// CreatePendingChildAccount creates a pending child account.
// @Summary Create a pending child account
// @Description Registers a child account that requires parental confirmation before activation
// @Tags registration
// @Accept json
// @Produce json
// @Param child body identity.CreatePendingChildAccountDto true "Pending child account details"
// @Success 201 {object} map[string]interface{} "Child account request created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /register/child/pending [post]
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

	if err := c.ChildAccountRegistrationService.CreatePendingChildAccount(r.Context(), nil, valueObject); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
