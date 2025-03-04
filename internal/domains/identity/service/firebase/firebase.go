package firebase

import (
	"api/internal/di"
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

func (s *Service) GetUserEmail(ctx context.Context, firebaseIdToken string) (string, *errLib.CommonError) {

	token, firebaseErr := s.FirebaseAuthClient.VerifyIDToken(ctx, firebaseIdToken)

	if firebaseErr != nil {
		log.Println("failed to verify: ", firebaseErr.Error())
		return "", errLib.New("Invalid Firebase token", http.StatusUnauthorized)
	}

	user, firebaseErr := s.FirebaseAuthClient.GetUser(ctx, token.UID)

	if firebaseErr != nil {
		return "", errLib.New("User not found", http.StatusUnauthorized)
	}

	return user.Email, nil
}
