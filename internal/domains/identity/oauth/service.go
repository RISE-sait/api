package oauth

import (
	"api/config"
	"api/internal/domains/identity/entities"
	"api/internal/domains/identity/infra/persistence/repository"
	errors "api/internal/libs/errors"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Service struct {
	StaffRepo *repository.StaffRepository
}

func NewService(staffRepo *repository.StaffRepository) *Service {
	return &Service{
		StaffRepo: staffRepo,
	}
}

func (s *Service) GetUserInfoRespBodyFromGoogleAPI(accessToken string) (io.ReadCloser, *errors.CommonError) {
	userInfoEndpoint := "https://www.googleapis.com/oauth2/v2/userinfo"
	resp, err := http.Get(fmt.Sprintf("%s?access_token=%s", userInfoEndpoint, accessToken))
	if err != nil {
		log.Printf("Failed to get user info using access token: %s", err.Error())
		return nil, errors.New("Authentication error", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	return resp.Body, nil
}

func (s *Service) SetUserInfoWithStaffDetails(c context.Context, userInfo entities.UserInfo) (*entities.UserInfo, *errors.CommonError) {

	staff, getStaffErr := s.StaffRepo.GetStaffByEmail(c, userInfo.Email)

	var staffInfo *entities.StaffInfo = nil

	if getStaffErr != nil {
		return nil, getStaffErr
	}

	staffInfo = &entities.StaffInfo{
		Role:     string(staff.Role),
		IsActive: staff.IsActive,
	}

	userInfo = entities.UserInfo{
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		StaffInfo: staffInfo,
	}

	return &userInfo, nil
}

func ExchangeCodeForToken(c context.Context, code string) (*oauth2.Token, *errors.CommonError) {
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
		return nil, errors.New("Authentication error", http.StatusInternalServerError)
	}
	return token, nil
}
