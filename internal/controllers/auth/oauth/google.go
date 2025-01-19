package oauth

import (
	"api/config"
	"api/internal/repositories"
	"api/internal/types/auth"
	"api/internal/utils"
	"api/internal/utils/validators"
	db "api/sqlc"
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Structure to store the access token response
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

func HandleOAuthCallback(w http.ResponseWriter, r *http.Request, staffRepo *repositories.StaffRepository) {
	var targetBody struct {
		Code string `json:"code" validate:"required_and_notwhitespace"`
	}

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	token, err := exchangeCodeForToken(r.Context(), targetBody.Code)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	userInfo, err := getUserInfo(token.AccessToken, staffRepo, r.Context())
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	signedToken, err := utils.SignJWT(*userInfo)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+signedToken)
	utils.RespondWithSuccess(w, nil, http.StatusOK)
}

func getUserInfo(accessToken string, staffRepo *repositories.StaffRepository, c context.Context) (*auth.UserInfo, *utils.HTTPError) {
	userInfoEndpoint := "https://www.googleapis.com/oauth2/v2/userinfo"
	resp, err := http.Get(fmt.Sprintf("%s?access_token=%s", userInfoEndpoint, accessToken))
	if err != nil {
		log.Printf("Failed to get user info using access token: %s", err.Error())
		return nil, utils.CreateHTTPError("Authentication error", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	var userInfo auth.UserInfo
	if err := validators.DecodeAndValidateRequestBody(resp.Body, &userInfo); err != nil {
		return nil, err
	}

	staff, getStaffErr := staffRepo.GetStaffByEmail(c, userInfo.Email)

	var staffInfo *auth.StaffInfo = nil

	if getStaffErr != nil {
		staffInfo = &auth.StaffInfo{
			Role:     string(db.StaffRoleEnum(staff.Role)),
			IsActive: staff.IsActive,
		}
	}

	userInfo = auth.UserInfo{
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		StaffInfo: staffInfo,
	}

	return &userInfo, nil
}

func exchangeCodeForToken(c context.Context, code string) (*oauth2.Token, *utils.HTTPError) {
	googleOauthConfig := &oauth2.Config{
		ClientID:     config.Envs.GoogleAuthConfig.ClientId,
		ClientSecret: config.Envs.GoogleAuthConfig.ClientSecret,
		RedirectURL:  config.Envs.GoogleAuthConfig.GoogleRedirectUrl,
		Scopes:       []string{"profile", "email"}, // Adjust scopes as needed
		Endpoint:     google.Endpoint,
	}

	token, err := googleOauthConfig.Exchange(c, code)
	if err != nil {
		log.Printf("Failed to exchange authorization code for exchange token: %s", err.Error())
		return nil, utils.CreateHTTPError("Authentication error", http.StatusInternalServerError)
	}
	return token, nil
}
