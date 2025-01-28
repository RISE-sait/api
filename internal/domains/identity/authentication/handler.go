package authentication

import (
	"api/internal/domains/identity/authentication/dto"
	"api/internal/domains/identity/authentication/values"
	"api/internal/libs/jwt"
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
	var dto dto.GetUserRequest
	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	credentials := values.NewCredentials(dto.Email, dto.Password)

	userInfo, err := h.AuthService.AuthenticateUser(r.Context(), credentials)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	token, err := jwt.SignJWT(*userInfo)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusNoContent)
}
