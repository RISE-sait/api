package jobs

import (
	"context"
	"database/sql"
	"log"
	"time"

	"api/internal/di"

	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
)

// FirebaseCleanupJob finds and removes Firebase users that don't exist in the database
type FirebaseCleanupJob struct {
	db                 *sql.DB
	firebaseAuthClient *auth.Client
	dryRun             bool
}

// excludedEmails contains emails that should never be deleted from Firebase
// even if they don't exist in the database (e.g., test accounts)
var excludedEmails = map[string]bool{
	"testadmin@rise.com": true,
}

// FirebaseCleanupResult contains the results of a cleanup operation
type FirebaseCleanupResult struct {
	TotalFirebaseUsers int      `json:"total_firebase_users"`
	TotalDBUsers       int      `json:"total_db_users"`
	OrphanedCount      int      `json:"orphaned_count"`
	DeletedCount       int      `json:"deleted_count"`
	FailedCount        int      `json:"failed_count"`
	DryRun             bool     `json:"dry_run"`
	OrphanedEmails     []string `json:"orphaned_emails,omitempty"`
	Errors             []string `json:"errors,omitempty"`
}

// NewFirebaseCleanupJob creates a new Firebase cleanup job
func NewFirebaseCleanupJob(container *di.Container) *FirebaseCleanupJob {
	return &FirebaseCleanupJob{
		db:                 container.DB,
		firebaseAuthClient: container.FirebaseService.FirebaseAuthClient,
		dryRun:             true, // Default to dry run for safety
	}
}

// Name returns the job name
func (j *FirebaseCleanupJob) Name() string {
	return "FirebaseCleanup"
}

// Interval returns how often this job runs (every hour as a safety net)
func (j *FirebaseCleanupJob) Interval() time.Duration {
	return 1 * time.Hour
}

// SetDryRun allows toggling dry run mode
func (j *FirebaseCleanupJob) SetDryRun(dryRun bool) {
	j.dryRun = dryRun
}

// Run executes the Firebase cleanup logic
func (j *FirebaseCleanupJob) Run(ctx context.Context) error {
	result, err := j.RunWithResult(ctx)
	if err != nil {
		return err
	}

	log.Printf("[FIREBASE-CLEANUP] Summary: firebase_users=%d, db_users=%d, orphaned=%d, deleted=%d, failed=%d, dry_run=%v",
		result.TotalFirebaseUsers, result.TotalDBUsers, result.OrphanedCount, result.DeletedCount, result.FailedCount, result.DryRun)

	return nil
}

// RunWithResult executes the cleanup and returns detailed results
func (j *FirebaseCleanupJob) RunWithResult(ctx context.Context) (*FirebaseCleanupResult, error) {
	result := &FirebaseCleanupResult{
		DryRun:         j.dryRun,
		OrphanedEmails: []string{},
		Errors:         []string{},
	}

	if j.dryRun {
		log.Printf("[FIREBASE-CLEANUP] Starting cleanup in DRY RUN mode - no users will be deleted")
	} else {
		log.Printf("[FIREBASE-CLEANUP] Starting cleanup - orphaned Firebase users will be deleted")
	}

	// 1. Get all emails from database
	dbEmails, err := j.getAllDBEmails(ctx)
	if err != nil {
		log.Printf("[FIREBASE-CLEANUP] Failed to get database emails: %v", err)
		return nil, err
	}
	result.TotalDBUsers = len(dbEmails)
	log.Printf("[FIREBASE-CLEANUP] Found %d users in database", len(dbEmails))

	// Create a set for fast lookup
	dbEmailSet := make(map[string]bool)
	for _, email := range dbEmails {
		dbEmailSet[email] = true
	}

	// 2. Iterate through Firebase users and find orphaned ones
	orphanedUsers := []struct {
		UID   string
		Email string
	}{}

	iter := j.firebaseAuthClient.Users(ctx, "")
	for {
		user, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("[FIREBASE-CLEANUP] Error iterating Firebase users: %v", err)
			result.Errors = append(result.Errors, "Iterator error: "+err.Error())
			break
		}

		result.TotalFirebaseUsers++

		// Check if Firebase user exists in database (and not in exclusion list)
		if user.Email != "" && !dbEmailSet[user.Email] && !excludedEmails[user.Email] {
			orphanedUsers = append(orphanedUsers, struct {
				UID   string
				Email string
			}{UID: user.UID, Email: user.Email})
			result.OrphanedEmails = append(result.OrphanedEmails, user.Email)
			log.Printf("[FIREBASE-CLEANUP] Found orphaned Firebase user: %s (%s)", user.UID, user.Email)
		}
	}

	result.OrphanedCount = len(orphanedUsers)
	log.Printf("[FIREBASE-CLEANUP] Found %d orphaned Firebase users out of %d total", result.OrphanedCount, result.TotalFirebaseUsers)

	// 3. Delete orphaned users (if not dry run)
	if !j.dryRun {
		for _, orphan := range orphanedUsers {
			if err := j.firebaseAuthClient.DeleteUser(ctx, orphan.UID); err != nil {
				log.Printf("[FIREBASE-CLEANUP] Failed to delete Firebase user %s (%s): %v", orphan.UID, orphan.Email, err)
				result.Errors = append(result.Errors, "Delete failed for "+orphan.Email+": "+err.Error())
				result.FailedCount++
			} else {
				log.Printf("[FIREBASE-CLEANUP] Successfully deleted orphaned Firebase user: %s (%s)", orphan.UID, orphan.Email)
				result.DeletedCount++
			}
		}
	} else {
		log.Printf("[FIREBASE-CLEANUP] DRY RUN: Would have deleted %d orphaned Firebase users", result.OrphanedCount)
	}

	return result, nil
}

// getAllDBEmails fetches all user emails from the database (including pending staff)
func (j *FirebaseCleanupJob) getAllDBEmails(ctx context.Context) ([]string, error) {
	// Query emails from users table and pending_staff table
	// Staff are already in users.users (staff.staff references users.users)
	// Pending staff have registered in Firebase but aren't approved yet
	// We include them here so they are NOT deleted from Firebase
	rows, err := j.db.QueryContext(ctx, `
		SELECT email FROM users.users
		WHERE email IS NOT NULL
		  AND email != ''
		  AND deleted_at IS NULL
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
