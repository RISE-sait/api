package traditional_auth

import (
	"api/internal/domains/identity/values"
	handlers "api/internal/libs/responses"
	response_handlers "api/internal/libs/responses"
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
	var dto GetUserRequest
	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	credentials := values.NewCredentials(dto.Email, dto.Password)

	token, err := h.AuthService.AuthenticateUser(r.Context(), credentials)
	if err != nil {
		handlers.RespondWithError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusNoContent)
}
