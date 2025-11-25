package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"sync"
	"time"

	"api/internal/libs/logger"
	"github.com/stripe/stripe-go/v81"
)

// RetryAttempt represents a failed webhook processing attempt
type RetryAttempt struct {
	EventID       string
	Event         stripe.Event
	AttemptNumber int
	NextRetryAt   time.Time
	LastError     error
	CreatedAt     time.Time
}

// WebhookRetryService manages webhook retry logic with exponential backoff
type WebhookRetryService struct {
	pendingRetries map[string]*RetryAttempt
	mutex          sync.RWMutex
	maxRetries     int
	baseDelay      time.Duration
	maxDelay       time.Duration
	multiplier     float64
	logger         *logger.StructuredLogger
	webhookService *WebhookService
	db             *sql.DB
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewWebhookRetryService creates a new webhook retry service
func NewWebhookRetryService(webhookService *WebhookService) *WebhookRetryService {
	return &WebhookRetryService{
		pendingRetries: make(map[string]*RetryAttempt),
		maxRetries:     5,                             // Maximum retry attempts
		baseDelay:      1 * time.Second,              // Initial delay
		maxDelay:       5 * time.Minute,              // Maximum delay between retries
		multiplier:     2.0,                          // Exponential backoff multiplier
		logger:         logger.WithComponent("webhook-retry"),
		webhookService: webhookService,
		db:             webhookService.db,            // Use DB from webhook service
		stopChan:       make(chan struct{}),
	}
}

// Start begins the retry processing goroutine
func (r *WebhookRetryService) Start(ctx context.Context) {
	r.wg.Add(1)
	go r.retryLoop(ctx)
	r.logger.Info("Webhook retry service started")
}

// Stop gracefully shuts down the retry service
func (r *WebhookRetryService) Stop() {
	close(r.stopChan)
	r.wg.Wait()
	r.logger.Info("Webhook retry service stopped")
}

// ScheduleRetry schedules a webhook event for retry
func (r *WebhookRetryService) ScheduleRetry(event stripe.Event, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	existing, exists := r.pendingRetries[event.ID]
	
	var attempt *RetryAttempt
	if exists {
		// Increment existing attempt
		attempt = existing
		attempt.AttemptNumber++
		attempt.LastError = err
	} else {
		// Create new retry attempt
		attempt = &RetryAttempt{
			EventID:       event.ID,
			Event:         event,
			AttemptNumber: 1,
			LastError:     err,
			CreatedAt:     time.Now(),
		}
	}

	// Check if we've exceeded max retries
	if attempt.AttemptNumber > r.maxRetries {
		r.logger.WithFields(map[string]interface{}{
			"event_id":       event.ID,
			"attempt_number": attempt.AttemptNumber,
			"max_retries":    r.maxRetries,
		}).Error("Webhook processing failed permanently after max retries", err)
		
		delete(r.pendingRetries, event.ID)
		
		// TODO: Store failed event in dead letter queue or notify administrators
		r.handlePermanentFailure(attempt)
		return
	}

	// Calculate next retry time with exponential backoff
	delay := r.calculateDelay(attempt.AttemptNumber)
	attempt.NextRetryAt = time.Now().Add(delay)

	r.pendingRetries[event.ID] = attempt

	r.logger.WithFields(map[string]interface{}{
		"event_id":       event.ID,
		"attempt_number": attempt.AttemptNumber,
		"next_retry_at":  attempt.NextRetryAt.Format(time.RFC3339),
		"delay_seconds":  delay.Seconds(),
	}).Warn("Webhook processing failed, scheduling retry")
}

// RemoveRetry removes a successfully processed event from retry queue
func (r *WebhookRetryService) RemoveRetry(eventID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if attempt, exists := r.pendingRetries[eventID]; exists {
		delete(r.pendingRetries, eventID)
		
		r.logger.WithFields(map[string]interface{}{
			"event_id":       eventID,
			"attempt_number": attempt.AttemptNumber,
		}).Info("Webhook retry successful, removed from retry queue")
	}
}

// GetRetryStats returns statistics about pending retries
func (r *WebhookRetryService) GetRetryStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := map[string]interface{}{
		"pending_retries": len(r.pendingRetries),
		"max_retries":     r.maxRetries,
		"base_delay":      r.baseDelay.String(),
		"max_delay":       r.maxDelay.String(),
	}

	// Count retries by attempt number
	attemptCounts := make(map[int]int)
	for _, attempt := range r.pendingRetries {
		attemptCounts[attempt.AttemptNumber]++
	}
	stats["attempts_distribution"] = attemptCounts

	return stats
}

// retryLoop is the main retry processing loop
func (r *WebhookRetryService) retryLoop(ctx context.Context) {
	defer r.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second) // Check for retries every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopChan:
			return
		case <-ticker.C:
			r.processRetries()
		}
	}
}

// processRetries processes all due retries
func (r *WebhookRetryService) processRetries() {
	now := time.Now()
	
	r.mutex.Lock()
	dueRetries := make([]*RetryAttempt, 0)
	
	for _, attempt := range r.pendingRetries {
		if now.After(attempt.NextRetryAt) {
			dueRetries = append(dueRetries, attempt)
		}
	}
	r.mutex.Unlock()

	if len(dueRetries) == 0 {
		return
	}

	r.logger.WithField("due_retries", len(dueRetries)).Info("Processing due webhook retries")

	// Process retries concurrently but limit concurrency
	semaphore := make(chan struct{}, 5) // Max 5 concurrent retries
	var wg sync.WaitGroup

	for _, attempt := range dueRetries {
		wg.Add(1)
		go func(a *RetryAttempt) {
			defer wg.Done()
			
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			r.processRetryAttempt(a)
		}(attempt)
	}

	wg.Wait()
}

// processRetryAttempt processes a single retry attempt
func (r *WebhookRetryService) processRetryAttempt(attempt *RetryAttempt) {
	retryLogger := r.logger.WithFields(map[string]interface{}{
		"event_id":       attempt.EventID,
		"attempt_number": attempt.AttemptNumber,
		"event_type":     attempt.Event.Type,
	})

	retryLogger.Info("Retrying webhook processing")

	// Process the webhook based on event type
	var err error
	switch attempt.Event.Type {
	case "checkout.session.completed":
		err = r.webhookService.HandleCheckoutSessionCompleted(attempt.Event)
	case "customer.subscription.created":
		err = r.webhookService.HandleSubscriptionCreated(attempt.Event)
	case "customer.subscription.updated":
		err = r.webhookService.HandleSubscriptionUpdated(attempt.Event)
	case "customer.subscription.deleted":
		err = r.webhookService.HandleSubscriptionDeleted(attempt.Event)
	case "invoice.payment_succeeded":
		err = r.webhookService.HandleInvoicePaymentSucceeded(attempt.Event)
	case "invoice.payment_failed":
		err = r.webhookService.HandleInvoicePaymentFailed(attempt.Event)
	default:
		retryLogger.Warn("Unknown event type for retry, removing from queue")
		r.RemoveRetry(attempt.EventID)
		return
	}

	if err != nil {
		// Retry failed, schedule another retry
		retryLogger.Error("Webhook retry failed", err)
		r.ScheduleRetry(attempt.Event, err)
	} else {
		// Retry successful, remove from queue
		retryLogger.Info("Webhook retry successful")
		r.RemoveRetry(attempt.EventID)
	}
}

// calculateDelay calculates the delay for a given attempt number using exponential backoff
func (r *WebhookRetryService) calculateDelay(attemptNumber int) time.Duration {
	// Exponential backoff: baseDelay * multiplier^(attemptNumber-1)
	delay := float64(r.baseDelay) * math.Pow(r.multiplier, float64(attemptNumber-1))
	
	// Add jitter (±25% randomization) to prevent thundering herd
	jitter := 1.0 + (0.5 - math.Mod(float64(time.Now().UnixNano()), 1.0)) * 0.5
	delay *= jitter
	
	// Cap at maximum delay
	if delay > float64(r.maxDelay) {
		delay = float64(r.maxDelay)
	}
	
	return time.Duration(delay)
}

// handlePermanentFailure handles webhooks that have permanently failed
func (r *WebhookRetryService) handlePermanentFailure(attempt *RetryAttempt) {
	r.logger.WithFields(map[string]interface{}{
		"event_id":       attempt.EventID,
		"event_type":     attempt.Event.Type,
		"total_attempts": attempt.AttemptNumber,
		"duration":       time.Since(attempt.CreatedAt).String(),
	}).Error("Webhook permanently failed after all retry attempts", attempt.LastError)

	// Store in dead letter queue for manual intervention
	r.insertIntoDeadLetterQueue(attempt)

	// Log the permanent failure details
	r.logger.WithFields(map[string]interface{}{
		"event_data": string(attempt.Event.Data.Raw),
		"event_id":   attempt.EventID,
	}).Error("Permanently failed webhook event details", nil)
}

// insertIntoDeadLetterQueue stores the failed webhook in the database for manual recovery
func (r *WebhookRetryService) insertIntoDeadLetterQueue(attempt *RetryAttempt) {
	if r.db == nil {
		log.Printf("[DEAD_LETTER] No database connection, cannot store failed webhook %s", attempt.EventID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert event data to JSON
	payloadJSON, err := json.Marshal(attempt.Event)
	if err != nil {
		log.Printf("[DEAD_LETTER] Failed to marshal event payload for %s: %v", attempt.EventID, err)
		payloadJSON = attempt.Event.Data.Raw // Fall back to raw data
	}

	errorMessage := ""
	if attempt.LastError != nil {
		errorMessage = attempt.LastError.Error()
	}

	query := `INSERT INTO payment.failed_webhooks (event_id, event_type, payload, error_message, attempts, status)
		VALUES ($1, $2, $3, $4, $5, 'failed')`

	_, dbErr := r.db.ExecContext(ctx, query, attempt.EventID, string(attempt.Event.Type), payloadJSON, errorMessage, attempt.AttemptNumber)
	if dbErr != nil {
		log.Printf("[DEAD_LETTER] CRITICAL: Failed to insert webhook into dead letter queue: event_id=%s, error=%v", attempt.EventID, dbErr)
	} else {
		log.Printf("[DEAD_LETTER] Successfully stored failed webhook in dead letter queue: event_id=%s, event_type=%s, attempts=%d",
			attempt.EventID, attempt.Event.Type, attempt.AttemptNumber)
	}
}

// GetPendingRetries returns all pending retry attempts (for monitoring/debugging)
func (r *WebhookRetryService) GetPendingRetries() []*RetryAttempt {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	retries := make([]*RetryAttempt, 0, len(r.pendingRetries))
	for _, attempt := range r.pendingRetries {
		retries = append(retries, attempt)
	}

	return retries
}