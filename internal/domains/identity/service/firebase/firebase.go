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

func (s *Service) DeleteUser(ctx context.Context, userEmail string) *errLib.CommonError {
	// Get user by email first
	user, firebaseErr := s.FirebaseAuthClient.GetUserByEmail(ctx, userEmail)
	if firebaseErr != nil {
		log.Printf("Failed to get Firebase user by email %s: %v", userEmail, firebaseErr)
		return errLib.New("Firebase user not found", http.StatusNotFound)
	}

	// Delete the user from Firebase
	firebaseErr = s.FirebaseAuthClient.DeleteUser(ctx, user.UID)
	if firebaseErr != nil {
		log.Printf("Failed to delete Firebase user %s: %v", user.UID, firebaseErr)
		return errLib.New("Failed to delete Firebase user", http.StatusInternalServerError)
	}

	log.Printf("Successfully deleted Firebase user: %s (%s)", user.UID, userEmail)
	return nil
}
