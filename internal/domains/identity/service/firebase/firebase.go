package firebase

import (
	"api/internal/di"
	"api/internal/domains/identity/entity"
	errLib "api/internal/libs/errors"
	"context"
	"log"
	"net/http"

	"firebase.google.com/go/auth"
)

type Service struct {
	FirebaseAuthClient *auth.Client
}

func NewFirebaseService(container *di.Container) *Service {

	firebaseAuthClient := container.FirebaseService.FirebaseAuthClient

	return &Service{
		FirebaseAuthClient: firebaseAuthClient,
	}
}

func (s *Service) GetUserInfo(ctx context.Context, firebaseIdToken string) (*entity.UserInfo, *errLib.CommonError) {

	token, firebaseErr := s.FirebaseAuthClient.VerifyIDToken(ctx, firebaseIdToken)

	if firebaseErr != nil {
		log.Println("failed to verify: ", firebaseErr.Error())
		return nil, errLib.New("Invalid Firebase token", http.StatusUnauthorized)
	}

	user, firebaseErr := s.FirebaseAuthClient.GetUser(ctx, token.UID)

	if firebaseErr != nil {
		return nil, errLib.New("User not found", http.StatusUnauthorized)
	}

	email := user.Email

	userInfo := entity.UserInfo{
		Email:     email,
		FirstName: user.DisplayName,
		LastName:  "",
	}

	return &userInfo, nil
}
