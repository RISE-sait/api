package jobs

import (
	"context"
	"database/sql"
	"log"
	"time"

	"api/internal/di"
	firebaseService "api/internal/domains/identity/service/firebase"
	"github.com/google/uuid"
)

// AccountDeletionJob permanently deletes accounts that have passed their scheduled deletion date
type AccountDeletionJob struct {
	db              *sql.DB
	firebaseService *firebaseService.Service
}

// NewAccountDeletionJob creates a new account deletion job
func NewAccountDeletionJob(container *di.Container) *AccountDeletionJob {
	return &AccountDeletionJob{
		db:              container.DB,
		firebaseService: firebaseService.NewFirebaseService(container),
	}
}

// Name returns the job name
func (j *AccountDeletionJob) Name() string {
	return "AccountDeletion"
}

// Interval returns how often this job runs (every hour)
func (j *AccountDeletionJob) Interval() time.Duration {
	return 1 * time.Hour
}

// Run executes the permanent deletion logic for accounts past their grace period
func (j *AccountDeletionJob) Run(ctx context.Context) error {
	log.Printf("[ACCOUNT-DELETION] Starting account deletion check")

	var (
		deleted int
		errors  int
	)

	// 1. Find soft-deleted accounts that have passed their scheduled deletion date
	softDeletedCount, softDeletedErrors := j.deleteSoftDeletedAccounts(ctx)
	deleted += softDeletedCount
	errors += softDeletedErrors

	// 2. Find archived accounts that have been archived for more than 30 days
	archivedCount, archivedErrors := j.deleteArchivedAccounts(ctx)
	deleted += archivedCount
	errors += archivedErrors

	log.Printf("[ACCOUNT-DELETION] Summary: deleted=%d, errors=%d", deleted, errors)
	return nil
}

// deleteSoftDeletedAccounts deletes accounts that have passed their scheduled deletion date
func (j *AccountDeletionJob) deleteSoftDeletedAccounts(ctx context.Context) (deleted int, errors int) {
	rows, err := j.db.QueryContext(ctx, `
		SELECT id, email, first_name, last_name, deleted_at, scheduled_deletion_at
		FROM users.users
		WHERE deleted_at IS NOT NULL
		  AND scheduled_deletion_at IS NOT NULL
		  AND scheduled_deletion_at < CURRENT_TIMESTAMP
		ORDER BY scheduled_deletion_at ASC
		LIMIT 50
	`)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Failed to query soft-deleted accounts: %v", err)
		return 0, 1
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userID              uuid.UUID
			email               sql.NullString
			firstName           string
			lastName            string
			deletedAt           time.Time
			scheduledDeletionAt time.Time
		)

		if err := rows.Scan(&userID, &email, &firstName, &lastName, &deletedAt, &scheduledDeletionAt); err != nil {
			log.Printf("[ACCOUNT-DELETION] Failed to scan soft-deleted row: %v", err)
			errors++
			continue
		}

		log.Printf("[ACCOUNT-DELETION] Processing soft-deleted account %s (%s %s, scheduled: %s)",
			userID, firstName, lastName, scheduledDeletionAt.Format(time.RFC3339))

		if err := j.permanentlyDeleteAccount(ctx, userID, email.String); err != nil {
			log.Printf("[ACCOUNT-DELETION] Failed to delete soft-deleted account %s: %v", userID, err)
			errors++
			continue
		}

		deleted++
	}

	return deleted, errors
}

// deleteArchivedAccounts deletes accounts that have been archived for more than 30 days
func (j *AccountDeletionJob) deleteArchivedAccounts(ctx context.Context) (deleted int, errors int) {
	rows, err := j.db.QueryContext(ctx, `
		SELECT id, email, first_name, last_name, archived_at
		FROM users.users
		WHERE is_archived = TRUE
		  AND archived_at IS NOT NULL
		  AND archived_at < CURRENT_TIMESTAMP - INTERVAL '30 days'
		ORDER BY archived_at ASC
		LIMIT 50
	`)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Failed to query archived accounts: %v", err)
		return 0, 1
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userID     uuid.UUID
			email      sql.NullString
			firstName  string
			lastName   string
			archivedAt time.Time
		)

		if err := rows.Scan(&userID, &email, &firstName, &lastName, &archivedAt); err != nil {
			log.Printf("[ACCOUNT-DELETION] Failed to scan archived row: %v", err)
			errors++
			continue
		}

		log.Printf("[ACCOUNT-DELETION] Processing archived account %s (%s %s, archived: %s, 30 days passed)",
			userID, firstName, lastName, archivedAt.Format(time.RFC3339))

		if err := j.permanentlyDeleteAccount(ctx, userID, email.String); err != nil {
			log.Printf("[ACCOUNT-DELETION] Failed to delete archived account %s: %v", userID, err)
			errors++
			continue
		}

		deleted++
	}

	return deleted, errors
}

// permanentlyDeleteAccount removes all user data from the database and Firebase
func (j *AccountDeletionJob) permanentlyDeleteAccount(ctx context.Context, userID uuid.UUID, userEmail string) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete from Firebase first (if email exists)
	if userEmail != "" {
		if firebaseErr := j.firebaseService.DeleteUser(ctx, userEmail); firebaseErr != nil {
			log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete Firebase user %s: %v (continuing with DB deletion)", userEmail, firebaseErr)
		}
	}

	// Helper function to safely execute delete with savepoint
	// This prevents one failed delete from aborting the entire transaction
	safeDelete := func(description, query string, args ...interface{}) {
		savepointName := fmt.Sprintf("sp_%d", time.Now().UnixNano())
		tx.ExecContext(ctx, "SAVEPOINT "+savepointName)
		_, execErr := tx.ExecContext(ctx, query, args...)
		if execErr != nil {
			tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savepointName)
			log.Printf("[ACCOUNT-DELETION] Warning: Failed to %s for %s: %v", description, userID, execErr)
		} else {
			tx.ExecContext(ctx, "RELEASE SAVEPOINT "+savepointName)
		}
	}

	// Delete related data in order (respecting foreign key constraints)
	safeDelete("delete push tokens", `DELETE FROM notifications.push_tokens WHERE user_id = $1`, userID)
	safeDelete("delete credit transactions", `DELETE FROM users.credit_transactions WHERE customer_id = $1`, userID)
	safeDelete("delete weekly credit usage", `DELETE FROM users.weekly_credit_usage WHERE customer_id = $1`, userID)
	safeDelete("delete active credit package", `DELETE FROM users.customer_active_credit_package WHERE customer_id = $1`, userID)
	safeDelete("delete membership plans", `DELETE FROM users.customer_membership_plans WHERE customer_id = $1`, userID)
	safeDelete("delete event enrollments", `DELETE FROM events.customer_enrollment WHERE customer_id = $1`, userID)
	safeDelete("delete attendance records", `DELETE FROM events.attendance WHERE user_id = $1`, userID)
	safeDelete("delete discount usage", `DELETE FROM users.customer_discount_usage WHERE customer_id = $1`, userID)
	safeDelete("delete waiver signings", `DELETE FROM waiver.waiver_signing WHERE user_id = $1`, userID)
	safeDelete("delete waiver uploads", `DELETE FROM waiver.waiver_uploads WHERE user_id = $1`, userID)
	safeDelete("delete subsidies", `DELETE FROM subsidies.customer_subsidies WHERE customer_id = $1`, userID)
	safeDelete("delete athlete record", `DELETE FROM athletic.athletes WHERE id = $1`, userID)
	safeDelete("orphan child accounts", `UPDATE users.users SET parent_id = NULL WHERE parent_id = $1`, userID)

	// Finally, delete the user record - this one must succeed
	result, err := tx.ExecContext(ctx, `DELETE FROM users.users WHERE id = $1`, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("[ACCOUNT-DELETION] Warning: User %s was already deleted", userID)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("[ACCOUNT-DELETION] ✅ Permanently deleted account %s", userID)
	return nil
}
