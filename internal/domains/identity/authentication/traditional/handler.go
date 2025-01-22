package traditional

import (
	"api/internal/libs/errors"
	handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type Handler struct {
	AuthService *Service
}

func NewHandler(authService *Service) *Handler {
	return &Handler{AuthService: authService}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var targetBody GetUserRequest
	if err := validators.ParseRequestBodyToJSON(r.Body, &targetBody); err != nil {

		newErr := errLib.New("Invalid request body", http.StatusBadRequest)
		handlers.RespondWithError(w, newErr)
		return
	}

	token, err := h.AuthService.AuthenticateUser(r.Context(), targetBody.Email, targetBody.Password)
	if err != nil {
		handlers.RespondWithError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
