package jobs

import (
	"context"
	"database/sql"
	"log"
	"time"

	"api/internal/di"
	"github.com/google/uuid"
)

// SubsidyExpirationJob automatically expires subsidies past their valid_until date
type SubsidyExpirationJob struct {
	db *sql.DB
}

// NewSubsidyExpirationJob creates a new subsidy expiration job
func NewSubsidyExpirationJob(container *di.Container) *SubsidyExpirationJob {
	return &SubsidyExpirationJob{
		db: container.DB,
	}
}

// Name returns the job name
func (j *SubsidyExpirationJob) Name() string {
	return "SubsidyExpiration"
}

// Interval returns how often this job runs (every 30 minutes)
func (j *SubsidyExpirationJob) Interval() time.Duration {
	return 30 * time.Minute
}

// Run executes the expiration logic
func (j *SubsidyExpirationJob) Run(ctx context.Context) error {
	log.Printf("[SUBSIDY-EXPIRATION] Starting subsidy expiration check")

	// Find subsidies that should be expired
	rows, err := j.db.QueryContext(ctx, `
		SELECT
			id,
			customer_id,
			approved_amount,
			total_amount_used,
			valid_until,
			status
		FROM subsidies.customer_subsidies
		WHERE status IN ('active', 'approved')
		  AND valid_until IS NOT NULL
		  AND valid_until < CURRENT_TIMESTAMP
		ORDER BY valid_until ASC
		LIMIT 100 -- Process 100 at a time
	`)
	if err != nil {
		log.Printf("[SUBSIDY-EXPIRATION] Failed to query expired subsidies: %v", err)
		return err
	}
	defer rows.Close()

	var (
		expired int
		errors  int
	)

	for rows.Next() {
		var (
			subsidyID      uuid.UUID
			customerID     uuid.UUID
			approvedAmount float64
			usedAmount     float64
			validUntil     time.Time
			status         string
		)

		if err := rows.Scan(&subsidyID, &customerID, &approvedAmount, &usedAmount, &validUntil, &status); err != nil {
			log.Printf("[SUBSIDY-EXPIRATION] Failed to scan row: %v", err)
			errors++
			continue
		}

		remainingBalance := approvedAmount - usedAmount

		log.Printf("[SUBSIDY-EXPIRATION] Expiring subsidy %s for customer %s (valid_until: %s, remaining: $%.2f)",
			subsidyID, customerID, validUntil.Format(time.RFC3339), remainingBalance)

		// Expire the subsidy
		if err := j.expireSubsidy(ctx, subsidyID, customerID, remainingBalance); err != nil {
			log.Printf("[SUBSIDY-EXPIRATION] Failed to expire subsidy %s: %v", subsidyID, err)
			errors++
			continue
		}

		expired++

		// TODO: Send email notification to customer
		// emailService.SendSubsidyExpiredEmail(customerID, subsidyID, remainingBalance)
	}

	log.Printf("[SUBSIDY-EXPIRATION] Summary: expired=%d, errors=%d", expired, errors)

	// TODO: Add Prometheus metrics
	// metrics.SubsidiesExpired.Add(float64(expired))
	// metrics.SubsidyExpirationErrors.Add(float64(errors))

	return nil
}

// expireSubsidy marks a subsidy as expired and creates an audit log
func (j *SubsidyExpirationJob) expireSubsidy(ctx context.Context, subsidyID, customerID uuid.UUID, remainingBalance float64) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update subsidy status to expired
	_, err = tx.ExecContext(ctx, `
		UPDATE subsidies.customer_subsidies
		SET status = 'expired', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, subsidyID)
	if err != nil {
		return err
	}

	// Create audit log
	_, err = tx.ExecContext(ctx, `
		INSERT INTO subsidies.audit_log (
			customer_subsidy_id,
			action,
			previous_status,
			new_status,
			notes,
			created_at
		) VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	`, subsidyID, "auto_expired", "active", "expired",
		"Automatically expired by scheduled job - subsidy reached valid_until date")
	if err != nil {
		log.Printf("[SUBSIDY-EXPIRATION] Warning: Failed to create audit log for subsidy %s: %v", subsidyID, err)
		// Don't fail the transaction if audit log creation fails
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("[SUBSIDY-EXPIRATION] ✅ Expired subsidy %s (customer: %s, forfeited: $%.2f)",
		subsidyID, customerID, remainingBalance)

	return nil
}
