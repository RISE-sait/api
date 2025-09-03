package payment

import (
	"context"
	"database/sql"
	"time"

	"api/internal/di"
	errLib "api/internal/libs/errors"
	"github.com/lib/pq"
	"net/http"
)

type WebhookIdempotencyRepository struct {
	Container *di.Container
}

func NewWebhookIdempotencyRepository(container *di.Container) *WebhookIdempotencyRepository {
	return &WebhookIdempotencyRepository{
		Container: container,
	}
}

// CheckAndStoreWebhookEvent checks if an event has been processed and stores it if not
// Returns true if the event has already been processed (idempotent)
func (r *WebhookIdempotencyRepository) CheckAndStoreWebhookEvent(ctx context.Context, eventID string, eventType string) (bool, *errLib.CommonError) {
	
	query := `
		INSERT INTO webhook_events (event_id, event_type, processed_at, created_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (event_id) 
		DO NOTHING
		RETURNING event_id
	`

	var returnedEventID string
	err := r.Container.DB.QueryRowContext(ctx, query, eventID, eventType).Scan(&returnedEventID)
	
	if err == sql.ErrNoRows {
		
		return true, nil
	}
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			// Handle specific database errors
			return false, errLib.New("Database error storing webhook event: "+pqErr.Message, http.StatusInternalServerError)
		}
		return false, errLib.New("Failed to store webhook event: "+err.Error(), http.StatusInternalServerError)
	}
	
	
	return false, nil
}


func (r *WebhookIdempotencyRepository) CleanupOldWebhookEvents(ctx context.Context, olderThan time.Duration) *errLib.CommonError {
	query := `
		DELETE FROM webhook_events 
		WHERE created_at < $1
	`
	
	cutoffTime := time.Now().Add(-olderThan)
	result, err := r.Container.DB.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return errLib.New("Failed to cleanup old webhook events: "+err.Error(), http.StatusInternalServerError)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {

	}
	
	return nil
}

// GetWebhookEventStatus returns the status of a webhook event
func (r *WebhookIdempotencyRepository) GetWebhookEventStatus(ctx context.Context, eventID string) (bool, time.Time, *errLib.CommonError) {
	query := `
		SELECT processed_at FROM webhook_events WHERE event_id = $1
	`
	
	var processedAt time.Time
	err := r.Container.DB.QueryRowContext(ctx, query, eventID).Scan(&processedAt)
	
	if err == sql.ErrNoRows {
		return false, time.Time{}, nil
	}
	
	if err != nil {
		return false, time.Time{}, errLib.New("Failed to get webhook event status: "+err.Error(), http.StatusInternalServerError)
	}
	
	return true, processedAt, nil
}