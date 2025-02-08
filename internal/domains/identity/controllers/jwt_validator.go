package identity

import (
	"api/internal/di"
	"api/internal/domains/identity/entities"
	errLib "api/internal/libs/errors"
	"api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"net/http"
)

type TokenValidationController struct{}

func NewTokenValidationController(container *di.Container) *TokenValidationController {
	return &TokenValidationController{}
}

func (h *TokenValidationController) ValidateToken(w http.ResponseWriter, r *http.Request) {
	cookie, cookieErr := r.Cookie("jwtToken")
	if cookieErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Missing or invalid token", http.StatusUnauthorized))
		return
	}

	claims, err := jwt.VerifyToken(cookie.Value)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	userInfo := entities.UserInfo{
		Name:  claims.Name,
		Email: claims.Email,
		StaffInfo: entities.StaffInfo{
			Role:     claims.Role,
			IsActive: claims.IsActive,
		},
	}

	response_handlers.RespondWithSuccess(w, userInfo, http.StatusOK)
}
