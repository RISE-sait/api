package identity

import (
	"api/cmd/server/di"
	errLib "api/internal/libs/errors"
	response_handlers "api/internal/libs/responses"
	"encoding/json"
	"io"
	"net/http"

	dto "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
)

type ChildAccountConfirmationController struct {
	ConfirmChildService *service.ConfirmChildService
}

func NewChildAccountConfirmationController(container *di.Container) *ChildAccountConfirmationController {
	return &ChildAccountConfirmationController{
		ConfirmChildService: service.NewConfirmChildService(container),
	}
}

func (c *ChildAccountConfirmationController) ConfirmChild(w http.ResponseWriter, r *http.Request) {

	body, ioErr := io.ReadAll(r.Body)

	if ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Failed to read request body", http.StatusBadRequest))
		return
	}

	var dto dto.ConfirmChildDto

	if ioErr := json.Unmarshal(body, &dto); ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid JSON format for credentials", http.StatusBadRequest))
		return
	}

	// Step 2: Call the service to create the account
	err := c.ConfirmChildService.ConfirmChild(r.Context(), dto.ChildEmail, dto.ParentEmail)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
