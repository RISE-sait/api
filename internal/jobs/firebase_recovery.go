package jobs

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"log"

	"api/internal/di"
	"api/utils/email"

	"firebase.google.com/go/auth"
)

// FirebaseRecoveryJob recreates Firebase users that exist in DB but not in Firebase
type FirebaseRecoveryJob struct {
	db                 *sql.DB
	firebaseAuthClient *auth.Client
	dryRun             bool
	limit              int    // Max users to recover (0 = unlimited)
	targetEmail        string // If set, only recover this specific email
}

// FirebaseRecoveryResult contains the results of a recovery operation
type FirebaseRecoveryResult struct {
	TotalDBUsers      int      `json:"total_db_users"`
	MissingInFirebase int      `json:"missing_in_firebase"`
	RecoveredCount    int      `json:"recovered_count"`
	AlreadyExistCount int      `json:"already_exist_count"`
	FailedCount       int      `json:"failed_count"`
	DryRun            bool     `json:"dry_run"`
	MissingEmails     []string `json:"missing_emails,omitempty"`
	RecoveredEmails   []string `json:"recovered_emails,omitempty"`
	Errors            []string `json:"errors,omitempty"`
}

// NewFirebaseRecoveryJob creates a new Firebase recovery job
func NewFirebaseRecoveryJob(container *di.Container) *FirebaseRecoveryJob {
	return &FirebaseRecoveryJob{
		db:                 container.DB,
		firebaseAuthClient: container.FirebaseService.FirebaseAuthClient,
		dryRun:             true, // Default to dry run for safety
	}
}

// SetDryRun allows toggling dry run mode
func (j *FirebaseRecoveryJob) SetDryRun(dryRun bool) {
	j.dryRun = dryRun
}

// SetLimit sets the max number of users to recover (0 = unlimited)
func (j *FirebaseRecoveryJob) SetLimit(limit int) {
	j.limit = limit
}

// SetTargetEmail sets a specific email to recover (for testing)
func (j *FirebaseRecoveryJob) SetTargetEmail(email string) {
	j.targetEmail = email
}

// Run executes the Firebase recovery logic
func (j *FirebaseRecoveryJob) Run(ctx context.Context) (*FirebaseRecoveryResult, error) {
	result := &FirebaseRecoveryResult{
		DryRun:          j.dryRun,
		MissingEmails:   []string{},
		RecoveredEmails: []string{},
		Errors:          []string{},
	}

	if j.dryRun {
		log.Printf("[FIREBASE-RECOVERY] Starting recovery in DRY RUN mode - no users will be created")
	} else {
		log.Printf("[FIREBASE-RECOVERY] Starting recovery - missing Firebase users will be recreated")
	}

	// 1. Get all emails from database (excluding children who don't have Firebase accounts)
	dbEmails, err := j.getRecoverableDBEmails(ctx)
	if err != nil {
		log.Printf("[FIREBASE-RECOVERY] Failed to get database emails: %v", err)
		return nil, err
	}
	result.TotalDBUsers = len(dbEmails)
	log.Printf("[FIREBASE-RECOVERY] Found %d users in database to check", len(dbEmails))

	// 2. Check each DB user against Firebase
	for _, email := range dbEmails {
		// If targeting a specific email, skip others
		if j.targetEmail != "" && email != j.targetEmail {
			continue
		}

		// Check if user exists in Firebase
		_, err := j.firebaseAuthClient.GetUserByEmail(ctx, email)
		if err == nil {
			// User exists in Firebase, skip
			result.AlreadyExistCount++
			continue
		}

		// User doesn't exist in Firebase - needs recovery
		result.MissingInFirebase++
		result.MissingEmails = append(result.MissingEmails, email)
		log.Printf("[FIREBASE-RECOVERY] Found missing Firebase user: %s", email)

		if !j.dryRun {
			// Check limit
			if j.limit > 0 && result.RecoveredCount >= j.limit {
				log.Printf("[FIREBASE-RECOVERY] Reached limit of %d users, stopping", j.limit)
				break
			}

			// Create user in Firebase with random password
			if err := j.createFirebaseUser(ctx, email); err != nil {
				log.Printf("[FIREBASE-RECOVERY] Failed to recover user %s: %v", email, err)
				result.Errors = append(result.Errors, "Failed to recover "+email+": "+err.Error())
				result.FailedCount++
			} else {
				log.Printf("[FIREBASE-RECOVERY] Successfully recovered user: %s", email)
				result.RecoveredEmails = append(result.RecoveredEmails, email)
				result.RecoveredCount++
			}
		}
	}

	log.Printf("[FIREBASE-RECOVERY] Summary: db_users=%d, missing=%d, recovered=%d, already_exist=%d, failed=%d, dry_run=%v",
		result.TotalDBUsers, result.MissingInFirebase, result.RecoveredCount, result.AlreadyExistCount, result.FailedCount, result.DryRun)

	return result, nil
}

// createFirebaseUser creates a new Firebase user with a random password and sends reset email
func (j *FirebaseRecoveryJob) createFirebaseUser(ctx context.Context, userEmail string) error {
	// Generate a random password (user will need to reset it)
	password, err := generateRandomPassword(32)
	if err != nil {
		return err
	}

	params := (&auth.UserToCreate{}).
		Email(userEmail).
		Password(password).
		EmailVerified(true) // Mark as verified since they were already registered

	_, err = j.firebaseAuthClient.CreateUser(ctx, params)
	if err != nil {
		return err
	}

	// Generate password reset link
	resetLink, err := j.firebaseAuthClient.PasswordResetLink(ctx, userEmail)
	if err != nil {
		log.Printf("[FIREBASE-RECOVERY] User created but failed to generate reset link for %s: %v", userEmail, err)
		// Don't fail the whole operation - user can use "forgot password" flow
		return nil
	}

	// Send the password reset email
	if err := email.SendAccountRecoveryEmail(userEmail, resetLink); err != nil {
		log.Printf("[FIREBASE-RECOVERY] User created but failed to send reset email to %s: %v", userEmail, err)
		// Log the link so it can be manually sent if needed
		log.Printf("[FIREBASE-RECOVERY] Manual reset link for %s: %s", userEmail, resetLink)
	}

	return nil
}

// generateRandomPassword generates a cryptographically secure random password
func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// getRecoverableDBEmails fetches emails of users who should have Firebase accounts
// This excludes children (who authenticate via parent) and deleted users
func (j *FirebaseRecoveryJob) getRecoverableDBEmails(ctx context.Context) ([]string, error) {
	// Get users who are NOT children (children don't have Firebase accounts)
	// Children have a non-null parent_id
	rows, err := j.db.QueryContext(ctx, `
		SELECT email FROM users.users
		WHERE email IS NOT NULL
		  AND email != ''
		  AND deleted_at IS NULL
		  AND parent_id IS NULL
		UNION
		SELECT email FROM staff.pending_staff
		WHERE email IS NOT NULL
		  AND email != ''
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}
