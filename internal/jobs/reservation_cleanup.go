package jobs

import (
	"context"
	"database/sql"
	"log"
	"time"

	"api/internal/di"
)

// ReservationCleanupJob deletes expired pending reservations to prevent table bloat
type ReservationCleanupJob struct {
	db *sql.DB
}

// NewReservationCleanupJob creates a new reservation cleanup job
func NewReservationCleanupJob(container *di.Container) *ReservationCleanupJob {
	return &ReservationCleanupJob{
		db: container.DB,
	}
}

// Name returns the job name
func (j *ReservationCleanupJob) Name() string {
	return "ReservationCleanup"
}

// Interval returns how often this job runs (every hour)
func (j *ReservationCleanupJob) Interval() time.Duration {
	return 1 * time.Hour
}

// Run executes the cleanup logic
func (j *ReservationCleanupJob) Run(ctx context.Context) error {
	log.Printf("[RESERVATION-CLEANUP] Starting cleanup of expired pending reservations")

	var (
		eventDeleted   int64
		programDeleted int64
	)

	// Delete expired pending event reservations (older than 1 hour to be safe)
	result, err := j.db.ExecContext(ctx, `
		DELETE FROM events.customer_enrollment
		WHERE payment_status = 'pending'
		  AND payment_expired_at < CURRENT_TIMESTAMP - interval '1 hour'
	`)
	if err != nil {
		log.Printf("[RESERVATION-CLEANUP] Failed to delete expired event reservations: %v", err)
		return err
	}
	eventDeleted, _ = result.RowsAffected()

	// Delete expired pending program reservations (older than 1 hour to be safe)
	result, err = j.db.ExecContext(ctx, `
		DELETE FROM program.customer_enrollment
		WHERE payment_status = 'pending'
		  AND payment_expired_at < CURRENT_TIMESTAMP - interval '1 hour'
	`)
	if err != nil {
		log.Printf("[RESERVATION-CLEANUP] Failed to delete expired program reservations: %v", err)
		return err
	}
	programDeleted, _ = result.RowsAffected()

	log.Printf("[RESERVATION-CLEANUP] Summary: events=%d, programs=%d deleted", eventDeleted, programDeleted)

	return nil
}
