package firebase

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
	"log"
	"net/http"
)

type Service struct {
	FirebaseAuthClient *auth.Client
}

func NewFirebaseService() (*Service, *errLib.CommonError) {

	authClient, err := getFirebaseAuthClient()

	if err != nil {
		return nil, errLib.New("Failed to get firebase auth client", http.StatusInternalServerError)
	}

	return &Service{
		FirebaseAuthClient: authClient,
	}, nil
}

func getFirebaseAuthClient() (*auth.Client, *errLib.CommonError) {

	// Load the Firebase service account key from an environment variable or a file
	firebaseCredentials := config.Envs.FirebaseCredentials
	if firebaseCredentials == "" {
		log.Printf("Firebase credentials not found in environment variables")
		return nil, errLib.New("Internal server error: Firebase credentials not found", http.StatusInternalServerError)
	}

	opt := option.WithCredentialsJSON([]byte(firebaseCredentials))
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
