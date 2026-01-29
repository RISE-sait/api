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

	// Find accounts that have passed their scheduled deletion date
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
		log.Printf("[ACCOUNT-DELETION] Failed to query accounts for deletion: %v", err)
		return err
	}
	defer rows.Close()

	var (
		deleted int
		errors  int
	)

	for rows.Next() {
		var (
			userID               uuid.UUID
			email                sql.NullString
			firstName            string
			lastName             string
			deletedAt            time.Time
			scheduledDeletionAt  time.Time
		)

		if err := rows.Scan(&userID, &email, &firstName, &lastName, &deletedAt, &scheduledDeletionAt); err != nil {
			log.Printf("[ACCOUNT-DELETION] Failed to scan row: %v", err)
			errors++
			continue
		}

		log.Printf("[ACCOUNT-DELETION] Processing permanent deletion for user %s (%s %s, scheduled: %s)",
			userID, firstName, lastName, scheduledDeletionAt.Format(time.RFC3339))

		if err := j.permanentlyDeleteAccount(ctx, userID, email.String); err != nil {
			log.Printf("[ACCOUNT-DELETION] Failed to permanently delete account %s: %v", userID, err)
			errors++
			continue
		}

		deleted++
	}

	log.Printf("[ACCOUNT-DELETION] Summary: deleted=%d, errors=%d", deleted, errors)
	return nil
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
			// Continue with database deletion even if Firebase deletion fails
			// The account might not exist in Firebase or was already deleted
		}
	}

	// Delete related data in order (respecting foreign key constraints)
	// Note: Many tables have ON DELETE CASCADE, but we explicitly delete for audit purposes

	// 1. Delete push notification tokens
	_, err = tx.ExecContext(ctx, `DELETE FROM notifications.push_tokens WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete push tokens for %s: %v", userID, err)
	}

	// 2. Delete credit transactions
	_, err = tx.ExecContext(ctx, `DELETE FROM users.credit_transactions WHERE customer_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete credit transactions for %s: %v", userID, err)
	}

	// 3. Delete weekly credit usage
	_, err = tx.ExecContext(ctx, `DELETE FROM users.weekly_credit_usage WHERE customer_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete weekly credit usage for %s: %v", userID, err)
	}

	// 4. Delete customer active credit packages
	_, err = tx.ExecContext(ctx, `DELETE FROM users.customer_active_credit_packages WHERE customer_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete active credit packages for %s: %v", userID, err)
	}

	// 5. Delete customer membership plans
	_, err = tx.ExecContext(ctx, `DELETE FROM users.customer_membership_plans WHERE customer_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete membership plans for %s: %v", userID, err)
	}

	// 6. Delete event enrollments
	_, err = tx.ExecContext(ctx, `DELETE FROM events.customer_enrollment WHERE customer_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete event enrollments for %s: %v", userID, err)
	}

	// 7. Delete event attendance records
	_, err = tx.ExecContext(ctx, `DELETE FROM events.attendance WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete attendance records for %s: %v", userID, err)
	}

	// 8. Delete discount usage records
	_, err = tx.ExecContext(ctx, `DELETE FROM discount.customer_discount_usage WHERE customer_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete discount usage for %s: %v", userID, err)
	}

	// 9. Delete waiver signings
	_, err = tx.ExecContext(ctx, `DELETE FROM users.waiver_signings WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete waiver signings for %s: %v", userID, err)
	}

	// 10. Delete subsidy records
	_, err = tx.ExecContext(ctx, `DELETE FROM subsidies.customer_subsidies WHERE customer_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete subsidies for %s: %v", userID, err)
	}

	// 11. Delete athlete record (if exists)
	_, err = tx.ExecContext(ctx, `DELETE FROM athletic.athletes WHERE id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to delete athlete record for %s: %v", userID, err)
	}

	// 12. Handle child accounts - set parent_id to NULL for any children
	_, err = tx.ExecContext(ctx, `UPDATE users.users SET parent_id = NULL WHERE parent_id = $1`, userID)
	if err != nil {
		log.Printf("[ACCOUNT-DELETION] Warning: Failed to orphan child accounts for %s: %v", userID, err)
	}

	// 13. Finally, delete the user record
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
