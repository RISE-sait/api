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

// DisableUser disables a Firebase user account (for soft delete - user cannot log in but data is preserved)
func (s *Service) DisableUser(ctx context.Context, userEmail string) *errLib.CommonError {
	// Get user by email first
	user, firebaseErr := s.FirebaseAuthClient.GetUserByEmail(ctx, userEmail)
	if firebaseErr != nil {
		log.Printf("Failed to get Firebase user by email %s: %v", userEmail, firebaseErr)
		return errLib.New("Firebase user not found", http.StatusNotFound)
	}

	// Disable the user account
	params := (&auth.UserToUpdate{}).Disabled(true)
	_, firebaseErr = s.FirebaseAuthClient.UpdateUser(ctx, user.UID, params)
	if firebaseErr != nil {
		log.Printf("Failed to disable Firebase user %s: %v", user.UID, firebaseErr)
		return errLib.New("Failed to disable Firebase user", http.StatusInternalServerError)
	}

	log.Printf("Successfully disabled Firebase user: %s (%s)", user.UID, userEmail)
	return nil
}

// EnableUser re-enables a disabled Firebase user account (for account recovery)
func (s *Service) EnableUser(ctx context.Context, userEmail string) *errLib.CommonError {
	// Get user by email first
	user, firebaseErr := s.FirebaseAuthClient.GetUserByEmail(ctx, userEmail)
	if firebaseErr != nil {
		log.Printf("Failed to get Firebase user by email %s: %v", userEmail, firebaseErr)
		return errLib.New("Firebase user not found", http.StatusNotFound)
	}

	// Enable the user account
	params := (&auth.UserToUpdate{}).Disabled(false)
	_, firebaseErr = s.FirebaseAuthClient.UpdateUser(ctx, user.UID, params)
	if firebaseErr != nil {
		log.Printf("Failed to enable Firebase user %s: %v", user.UID, firebaseErr)
		return errLib.New("Failed to enable Firebase user", http.StatusInternalServerError)
	}

	log.Printf("Successfully enabled Firebase user: %s (%s)", user.UID, userEmail)
	return nil
}

// UpdateUserEmail updates a Firebase user's email address
func (s *Service) UpdateUserEmail(ctx context.Context, currentEmail string, newEmail string) *errLib.CommonError {
	// Get user by current email first
	user, firebaseErr := s.FirebaseAuthClient.GetUserByEmail(ctx, currentEmail)
	if firebaseErr != nil {
		log.Printf("Failed to get Firebase user by email %s: %v", currentEmail, firebaseErr)
		return errLib.New("Firebase user not found", http.StatusNotFound)
	}

	// Update the email
	params := (&auth.UserToUpdate{}).Email(newEmail)
	_, firebaseErr = s.FirebaseAuthClient.UpdateUser(ctx, user.UID, params)
	if firebaseErr != nil {
		log.Printf("Failed to update Firebase user email from %s to %s: %v", currentEmail, newEmail, firebaseErr)
		return errLib.New("Failed to update Firebase user email", http.StatusInternalServerError)
	}

	log.Printf("Successfully updated Firebase user email from %s to %s", currentEmail, newEmail)
	return nil
}
