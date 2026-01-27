package email_change

import (
	"api/internal/di"
	"api/internal/domains/identity/service/email_verification"
	"api/internal/domains/identity/service/firebase"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type EmailChangeService struct {
	DB              *sql.DB
	FirebaseService *firebase.Service
	VerificationSvc *email_verification.EmailVerificationService
}

func NewEmailChangeService(container *di.Container) *EmailChangeService {
	return &EmailChangeService{
		DB:              container.DB,
		FirebaseService: firebase.NewFirebaseService(container),
		VerificationSvc: email_verification.NewEmailVerificationService(container),
	}
}

// InitiateEmailChange starts the email change process by storing the pending email and sending verification
func (s *EmailChangeService) InitiateEmailChange(ctx context.Context, userID uuid.UUID, newEmail string) (string, *errLib.CommonError) {
	// Check if the new email already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users.users WHERE email = $1 AND deleted_at IS NULL)`
	if err := s.DB.QueryRowContext(ctx, checkQuery, newEmail).Scan(&exists); err != nil {
		log.Printf("Failed to check if email exists: %v", err)
		return "", errLib.New("Failed to check email availability", http.StatusInternalServerError)
	}

	if exists {
		return "", errLib.New("Email address is already in use", http.StatusConflict)
	}

	// Get user info for validation
	var currentEmail sql.NullString
	var firstName string
	getUserQuery := `SELECT email, first_name FROM users.users WHERE id = $1 AND deleted_at IS NULL`
	if err := s.DB.QueryRowContext(ctx, getUserQuery, userID).Scan(&currentEmail, &firstName); err != nil {
		if err == sql.ErrNoRows {
			return "", errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Failed to get user info: %v", err)
		return "", errLib.New("Failed to get user info", http.StatusInternalServerError)
	}

	// Check if new email is same as current
	if currentEmail.Valid && currentEmail.String == newEmail {
		return "", errLib.New("New email is the same as current email", http.StatusBadRequest)
	}

	// Generate verification token
	token, tokenErr := s.VerificationSvc.GenerateVerificationToken()
	if tokenErr != nil {
		return "", tokenErr
	}

	// Store pending email change
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	updateQuery := `
		UPDATE users.users
		SET pending_email = $2,
		    pending_email_token = $3,
		    pending_email_token_expires_at = $4,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.DB.ExecContext(ctx, updateQuery, userID, newEmail, token, expiresAt)
	if err != nil {
		log.Printf("Failed to store pending email change for user %s: %v", userID, err)
		return "", errLib.New("Failed to initiate email change", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return "", errLib.New("User not found", http.StatusNotFound)
	}

	log.Printf("Initiated email change for user %s to %s (expires at %s)", userID, newEmail, expiresAt.Format(time.RFC3339))

	return token, nil
}

// VerifyAndCompleteEmailChange verifies the token and completes the email change
func (s *EmailChangeService) VerifyAndCompleteEmailChange(ctx context.Context, token string) *errLib.CommonError {
	// Find user with this token
	var userID uuid.UUID
	var currentEmail sql.NullString
	var pendingEmail sql.NullString
	var expiresAt sql.NullTime

	query := `
		SELECT id, email, pending_email, pending_email_token_expires_at
		FROM users.users
		WHERE pending_email_token = $1 AND deleted_at IS NULL
	`

	err := s.DB.QueryRowContext(ctx, query, token).Scan(&userID, &currentEmail, &pendingEmail, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Invalid email change token: %s", token)
			return errLib.New("Invalid or expired verification token", http.StatusBadRequest)
		}
		log.Printf("Failed to query email change token: %v", err)
		return errLib.New("Failed to verify token", http.StatusInternalServerError)
	}

	// Check if pending email exists
	if !pendingEmail.Valid || pendingEmail.String == "" {
		return errLib.New("No pending email change found", http.StatusBadRequest)
	}

	// Check if token has expired
	if !expiresAt.Valid || time.Now().UTC().After(expiresAt.Time) {
		log.Printf("Email change token expired for user %s (expired at %s)", userID, expiresAt.Time.Format(time.RFC3339))
		return errLib.New("Verification token has expired. Please request a new email change.", http.StatusGone)
	}

	// Begin transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return errLib.New("Failed to complete email change", http.StatusInternalServerError)
	}
	defer tx.Rollback()

	// Update email in database
	updateQuery := `
		UPDATE users.users
		SET email = pending_email,
		    pending_email = NULL,
		    pending_email_token = NULL,
		    pending_email_token_expires_at = NULL,
		    email_changed_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND pending_email IS NOT NULL
	`

	result, updateErr := tx.ExecContext(ctx, updateQuery, userID)
	if updateErr != nil {
		log.Printf("Failed to update email for user %s: %v", userID, updateErr)
		return errLib.New("Failed to complete email change", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errLib.New("Failed to complete email change", http.StatusInternalServerError)
	}

	// Update email in Firebase
	if currentEmail.Valid {
		firebaseErr := s.FirebaseService.UpdateUserEmail(ctx, currentEmail.String, pendingEmail.String)
		if firebaseErr != nil {
			log.Printf("Failed to update Firebase email for user %s: %v", userID, firebaseErr.Message)
			return errLib.New("Failed to update email in authentication system", http.StatusInternalServerError)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit email change transaction: %v", err)
		return errLib.New("Failed to complete email change", http.StatusInternalServerError)
	}

	log.Printf("Successfully changed email for user %s from %s to %s", userID, currentEmail.String, pendingEmail.String)
	return nil
}

// ResendEmailChangeVerification generates a new token and returns info for sending verification email
func (s *EmailChangeService) ResendEmailChangeVerification(ctx context.Context, userID uuid.UUID) (string, string, string, *errLib.CommonError) {
	// Get user info including pending email
	var firstName string
	var pendingEmail sql.NullString

	query := `
		SELECT first_name, pending_email
		FROM users.users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&firstName, &pendingEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", "", errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Failed to get user info: %v", err)
		return "", "", "", errLib.New("Failed to resend verification", http.StatusInternalServerError)
	}

	// Check if there's a pending email change
	if !pendingEmail.Valid || pendingEmail.String == "" {
		return "", "", "", errLib.New("No pending email change to verify", http.StatusBadRequest)
	}

	// Generate new token
	token, tokenErr := s.VerificationSvc.GenerateVerificationToken()
	if tokenErr != nil {
		return "", "", "", tokenErr
	}

	// Update token and expiration
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	updateQuery := `
		UPDATE users.users
		SET pending_email_token = $2,
		    pending_email_token_expires_at = $3,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND pending_email IS NOT NULL AND deleted_at IS NULL
	`

	result, updateErr := s.DB.ExecContext(ctx, updateQuery, userID, token, expiresAt)
	if updateErr != nil {
		log.Printf("Failed to update email change token for user %s: %v", userID, updateErr)
		return "", "", "", errLib.New("Failed to resend verification", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return "", "", "", errLib.New("No pending email change found", http.StatusBadRequest)
	}

	log.Printf("Resent email change verification for user %s to %s", userID, pendingEmail.String)

	return firstName, pendingEmail.String, token, nil
}

// CancelPendingEmailChange cancels any pending email change for the user
func (s *EmailChangeService) CancelPendingEmailChange(ctx context.Context, userID uuid.UUID) *errLib.CommonError {
	query := `
		UPDATE users.users
		SET pending_email = NULL,
		    pending_email_token = NULL,
		    pending_email_token_expires_at = NULL,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.DB.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("Failed to cancel pending email change for user %s: %v", userID, err)
		return errLib.New("Failed to cancel email change", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errLib.New("User not found", http.StatusNotFound)
	}

	log.Printf("Cancelled pending email change for user %s", userID)
	return nil
}

// GetPendingEmail returns the pending email for a user if one exists
func (s *EmailChangeService) GetPendingEmail(ctx context.Context, userID uuid.UUID) (string, *errLib.CommonError) {
	var pendingEmail sql.NullString

	query := `SELECT pending_email FROM users.users WHERE id = $1 AND deleted_at IS NULL`
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&pendingEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Failed to get pending email for user %s: %v", userID, err)
		return "", errLib.New("Failed to get pending email", http.StatusInternalServerError)
	}

	if !pendingEmail.Valid || pendingEmail.String == "" {
		return "", nil
	}

	return pendingEmail.String, nil
}

// GetEmailChangeVerificationURL generates the full verification URL with the token
func (s *EmailChangeService) GetEmailChangeVerificationURL(token string, baseURL string) string {
	return fmt.Sprintf("%s/verify-email-change?token=%s", baseURL, token)
}
