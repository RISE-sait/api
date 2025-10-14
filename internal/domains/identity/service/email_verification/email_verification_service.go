package email_verification

import (
	"api/internal/di"
	errLib "api/internal/libs/errors"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// EmailVerificationService handles email verification operations
type EmailVerificationService struct {
	DB *sql.DB
}

// NewEmailVerificationService initializes a new EmailVerificationService instance
func NewEmailVerificationService(container *di.Container) *EmailVerificationService {
	return &EmailVerificationService{
		DB: container.DB,
	}
}

// GenerateVerificationToken generates a cryptographically secure random token
func (s *EmailVerificationService) GenerateVerificationToken() (string, *errLib.CommonError) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Failed to generate verification token: %v", err)
		return "", errLib.New("Failed to generate verification token", http.StatusInternalServerError)
	}

	// Convert to hex string (64 characters)
	token := hex.EncodeToString(bytes)
	return token, nil
}

// StoreVerificationToken stores the verification token for a user
// Tokens are valid for 24 hours
func (s *EmailVerificationService) StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string) *errLib.CommonError {
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	query := `
		UPDATE users.users
		SET email_verification_token = $1,
		    email_verification_token_expires_at = $2,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND email_verified = FALSE
	`

	result, err := s.DB.ExecContext(ctx, query, token, expiresAt, userID)
	if err != nil {
		log.Printf("Failed to store verification token for user %s: %v", userID, err)
		return errLib.New("Failed to store verification token", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("User %s not found or already verified", userID)
		return errLib.New("User not found or already verified", http.StatusNotFound)
	}

	log.Printf("Stored verification token for user %s (expires at %s)", userID, expiresAt.Format(time.RFC3339))
	return nil
}

// VerifyEmailToken verifies a token and marks the user's email as verified
func (s *EmailVerificationService) VerifyEmailToken(ctx context.Context, token string) *errLib.CommonError {
	// First, find the user with this token and check if it's valid
	var userID uuid.UUID
	var expiresAt time.Time
	var emailVerified bool

	query := `
		SELECT id, email_verification_token_expires_at, email_verified
		FROM users.users
		WHERE email_verification_token = $1
	`

	err := s.DB.QueryRowContext(ctx, query, token).Scan(&userID, &expiresAt, &emailVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Invalid verification token: %s", token)
			return errLib.New("Invalid verification token", http.StatusBadRequest)
		}
		log.Printf("Failed to query verification token: %v", err)
		return errLib.New("Failed to verify token", http.StatusInternalServerError)
	}

	// Check if already verified
	if emailVerified {
		log.Printf("User %s email already verified", userID)
		return errLib.New("Email already verified", http.StatusConflict)
	}

	// Check if token has expired
	if time.Now().UTC().After(expiresAt) {
		log.Printf("Verification token expired for user %s (expired at %s)", userID, expiresAt.Format(time.RFC3339))
		return errLib.New("Verification token has expired. Please request a new verification email.", http.StatusGone)
	}

	// Mark email as verified and clear the token
	updateQuery := `
		UPDATE users.users
		SET email_verified = TRUE,
		    email_verified_at = CURRENT_TIMESTAMP,
		    email_verification_token = NULL,
		    email_verification_token_expires_at = NULL,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, updateErr := s.DB.ExecContext(ctx, updateQuery, userID)
	if updateErr != nil {
		log.Printf("Failed to mark email as verified for user %s: %v", userID, updateErr)
		return errLib.New("Failed to verify email", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Failed to update verification status for user %s", userID)
		return errLib.New("Failed to verify email", http.StatusInternalServerError)
	}

	log.Printf("Successfully verified email for user %s", userID)
	return nil
}

// ResendVerificationEmail generates a new token and resends the verification email
func (s *EmailVerificationService) ResendVerificationEmail(ctx context.Context, email string) (uuid.UUID, string, string, *errLib.CommonError) {
	// Find the user by email
	var userID uuid.UUID
	var firstName string
	var emailVerified bool

	query := `
		SELECT id, first_name, email_verified
		FROM users.users
		WHERE email = $1 AND deleted_at IS NULL
	`

	err := s.DB.QueryRowContext(ctx, query, email).Scan(&userID, &firstName, &emailVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found with email: %s", email)
			return uuid.Nil, "", "", errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Failed to query user by email: %v", err)
		return uuid.Nil, "", "", errLib.New("Failed to resend verification email", http.StatusInternalServerError)
	}

	// Check if already verified
	if emailVerified {
		log.Printf("Email already verified for user %s", userID)
		return uuid.Nil, "", "", errLib.New("Email already verified", http.StatusConflict)
	}

	// Generate new token
	token, tokenErr := s.GenerateVerificationToken()
	if tokenErr != nil {
		return uuid.Nil, "", "", tokenErr
	}

	// Store the new token
	if storeErr := s.StoreVerificationToken(ctx, userID, token); storeErr != nil {
		return uuid.Nil, "", "", storeErr
	}

	return userID, firstName, token, nil
}

// IsEmailVerified checks if a user's email is verified
func (s *EmailVerificationService) IsEmailVerified(ctx context.Context, userID uuid.UUID) (bool, *errLib.CommonError) {
	var emailVerified bool

	query := `
		SELECT email_verified
		FROM users.users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&emailVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found: %s", userID)
			return false, errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Failed to check email verification status: %v", err)
		return false, errLib.New("Failed to check verification status", http.StatusInternalServerError)
	}

	return emailVerified, nil
}

// GetVerificationURL generates the full verification URL with the token
func (s *EmailVerificationService) GetVerificationURL(token string, baseURL string) string {
	return fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)
}
