package identity

import (
	"api/cmd/server/di"
	identity "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	errLib "api/internal/libs/errors"
	response_handlers "api/internal/libs/responses"
	"encoding/json"
	"io"
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

	body, ioErr := io.ReadAll(r.Body)

	if ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Failed to read request body", http.StatusBadRequest))
		return
	}

	var customerWaiverDto identity.CustomerWaiverCreateDto
	var childAccountDto identity.CreateChildAccountDto
	var credentialsDto identity.Credentials

	if ioErr := json.Unmarshal(body, &customerWaiverDto); ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid JSON format for waiver data", http.StatusBadRequest))
		return
	}

	if ioErr := json.Unmarshal(body, &childAccountDto); ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid format for parent email", http.StatusBadRequest))
		return
	}

	if ioErr := json.Unmarshal(body, &credentialsDto); ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid format for parent email", http.StatusBadRequest))
		return
	}

	// Step 2: Call the service to create the account
	err := c.ChildAccountRegistrationService.CreatePendingAccount(r.Context(), &credentialsDto, &customerWaiverDto, &childAccountDto)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
