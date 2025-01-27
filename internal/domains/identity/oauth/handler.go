package oauth

import (
	"api/internal/domains/identity/entities"
	"api/internal/domains/identity/lib"
	errLib "api/internal/libs/errors"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"log"
	"net/http"
)

type Handler struct {
	AuthService *Service
}

func NewHandler(authService *Service) *Handler {
	return &Handler{AuthService: authService}
}

// TokenResponse Structure to store the access token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token" validate:"omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")

	if code == "" {
		response_handlers.RespondWithError(w, errLib.New("Authorization code is missing", http.StatusBadRequest))
		return
	}

	token, err := ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	userInfoRespBody, err := h.AuthService.GetUserInfoRespBodyFromGoogleAPI(token.AccessToken)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	var userInfo *entities.UserInfo

	if err := validators.ParseJSON(userInfoRespBody, &userInfo); err != nil {
		log.Println("Error getting user info", err)
		response_handlers.RespondWithError(w, errLib.New("Failed to parse user info from Google", http.StatusInternalServerError))
		return
	}

	userInfo, err = h.AuthService.SetUserInfoWithStaffDetails(r.Context(), *userInfo)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	signedToken, err := lib.SignJWT(*userInfo)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+signedToken)
	w.WriteHeader(http.StatusOK)
}
