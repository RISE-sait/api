package identity

import (
	"api/internal/di"
	identity "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/utils/email"
	"net/http"
)

type CreatePendingChildAccountController struct {
	ChildAccountRegistrationService *service.ChildAccountRequestService
	AccontRegistrationService       *service.AccountRegistrationService
}

func NewCreatePendingChildAccountController(container *di.Container) *CreatePendingChildAccountController {

	return &CreatePendingChildAccountController{
		ChildAccountRegistrationService: service.NewChildAccountRegistrationRequestService(container),
		AccontRegistrationService:       service.NewAccountRegistrationService(container),
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

	accountRegistrationCredentials := values.RegisterCredentials{
		Email:    valueObject.Email,
		Password: valueObject.Password,
	}

	// Step 2: Call the service to create the account
	tx, _, err := c.AccontRegistrationService.CreateAccount(r.Context(), &accountRegistrationCredentials)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	// Step 2: Call the service to create the account
	err = c.ChildAccountRegistrationService.CreatePendingAccount(r.Context(), tx, valueObject)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := email.SendConfirmChildEmail(dto.ParentEmail, dto.Child.Email); err != nil {
		tx.Rollback()
		response_handlers.RespondWithError(w, errLib.New("Failed to send email", http.StatusInternalServerError))
	}

	if err := tx.Commit(); err != nil {
		response_handlers.RespondWithError(w, errLib.New("Failed to commit transaction", http.StatusInternalServerError))
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
