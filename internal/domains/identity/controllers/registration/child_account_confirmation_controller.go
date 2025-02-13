package registration

import (
	"api/internal/di"
	errLib "api/internal/libs/errors"
	response_handlers "api/internal/libs/responses"
	"net/http"

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

	childEmail := r.URL.Query().Get("child")
	parentEmail := r.URL.Query().Get("parent")

	if childEmail == "" || parentEmail == "" {
		response_handlers.RespondWithError(w, errLib.New("Missing required query parameters: child and parent", http.StatusBadRequest))
		return
	}

	// Step 2: Call the service to create the account
	err := c.ConfirmChildService.ConfirmChild(r.Context(), childEmail, parentEmail)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
