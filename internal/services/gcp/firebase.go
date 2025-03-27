package gcp

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
)

type Service struct {
	FirebaseAuthClient *auth.Client
}

func NewFirebaseService() (*Service, *errLib.CommonError) {

	authClient, err := getFirebaseAuthClient()

	if err != nil {
		return nil, err
	}

	return &Service{
		FirebaseAuthClient: authClient,
	}, nil
}

func getFirebaseAuthClient() (*auth.Client, *errLib.CommonError) {

	var opt option.ClientOption

	if gcpServiceAccountCredentials := config.Env.GcpServiceAccountCredentials; gcpServiceAccountCredentials != "" {
		opt = option.WithCredentialsJSON([]byte(gcpServiceAccountCredentials))
	} else if _, err := os.Stat("/app/config/gcp-service-account.json"); err == nil {
		opt = option.WithCredentialsFile("/app/config/gcp-service-account.json")
	} else {
		log.Printf("Firebase credentials not found in environment variables or file")
		return nil, errLib.New("Internal server error: Firebase credentials not found", http.StatusInternalServerError)
	}

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Printf("error initializing app: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Printf("error initializing Firebase Auth client: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return authClient, nil
}

func (s *Service) GetUserFromFirebase(ctx context.Context, idToken string) (*auth.UserRecord, *errLib.CommonError) {
	token, firebaseErr := s.FirebaseAuthClient.VerifyIDToken(ctx, idToken)

	if firebaseErr != nil {
		return nil, errLib.New("Invalid Firebase token", http.StatusUnauthorized)
	}

	user, firebaseErr := s.FirebaseAuthClient.GetUser(ctx, token.UID)

	if firebaseErr != nil {
		return nil, errLib.New("User not found", http.StatusUnauthorized)
	}

	return user, nil
}
