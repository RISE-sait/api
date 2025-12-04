package payment

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"
)

// WebhookIdempotency provides database-backed idempotency protection for webhooks
// with in-memory caching for performance
type WebhookIdempotency struct {
	db              *sql.DB
	cache           map[string]time.Time
	mutex           sync.RWMutex
	maxCacheAge     time.Duration
	maxCacheSize    int
}

// NewWebhookIdempotency creates a new webhook idempotency tracker with database backing
func NewWebhookIdempotency(maxAge time.Duration, maxSize int) *WebhookIdempotency {
	idempotency := &WebhookIdempotency{
		cache:        make(map[string]time.Time),
		maxCacheAge:  maxAge,
		maxCacheSize: maxSize,
	}

	// Start cleanup goroutine for cache
	go idempotency.cleanupLoop()

	return idempotency
}

// NewWebhookIdempotencyWithDB creates a new webhook idempotency tracker with database connection
func NewWebhookIdempotencyWithDB(db *sql.DB, maxAge time.Duration, maxSize int) *WebhookIdempotency {
	idempotency := &WebhookIdempotency{
		db:           db,
		cache:        make(map[string]time.Time),
		maxCacheAge:  maxAge,
		maxCacheSize: maxSize,
	}

	// Start cleanup goroutines
	go idempotency.cleanupLoop()
	go idempotency.dbCleanupLoop()

	return idempotency
}

// IsProcessed checks if an event has already been processed
func (w *WebhookIdempotency) IsProcessed(eventID string) bool {
	// First check in-memory cache (fast path)
	w.mutex.RLock()
	processedAt, exists := w.cache[eventID]
	w.mutex.RUnlock()

	if exists && time.Since(processedAt) <= w.maxCacheAge {
		return true
	}

	// If not in cache and we have DB, check database
	if w.db != nil {
		return w.isProcessedInDB(eventID)
	}

	return false
}

// TryClaimEvent atomically attempts to claim an event for processing.
// Returns true if this caller successfully claimed the event (should process it).
// Returns false if the event was already claimed by another process.
// This prevents race conditions where two concurrent handlers both pass IsProcessed check.
func (w *WebhookIdempotency) TryClaimEvent(eventID, eventType string) bool {
	// First check in-memory cache (fast path for recently processed)
	w.mutex.RLock()
	processedAt, exists := w.cache[eventID]
	w.mutex.RUnlock()

	if exists && time.Since(processedAt) <= w.maxCacheAge {
		log.Printf("[IDEMPOTENCY] Event %s already in cache, rejecting claim", eventID)
		return false
	}

	// Use database atomic insert to claim the event
	if w.db != nil {
		claimed := w.tryClaimInDB(eventID, eventType)
		if claimed {
			// Successfully claimed - add to cache
			w.mutex.Lock()
			if len(w.cache) >= w.maxCacheSize {
				w.removeOldestEntries(w.maxCacheSize / 4)
			}
			w.cache[eventID] = time.Now()
			w.mutex.Unlock()
		}
		return claimed
	}

	// No database - use cache-only with mutex protection
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Double-check after acquiring write lock
	if _, exists := w.cache[eventID]; exists {
		return false
	}

	// Claim it
	if len(w.cache) >= w.maxCacheSize {
		w.removeOldestEntries(w.maxCacheSize / 4)
	}
	w.cache[eventID] = time.Now()
	return true
}

// tryClaimInDB atomically attempts to insert the event and returns whether it succeeded
func (w *WebhookIdempotency) tryClaimInDB(eventID, eventType string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use INSERT with ON CONFLICT DO NOTHING and check rows affected
	// This is atomic - only one concurrent caller can successfully insert
	query := `INSERT INTO payment.webhook_events (event_id, event_type, processed_at, status)
		VALUES ($1, $2, NOW(), 'processing')
		ON CONFLICT (event_id) DO NOTHING`

	result, err := w.db.ExecContext(ctx, query, eventID, eventType)
	if err != nil {
		log.Printf("[IDEMPOTENCY] Failed to claim event %s in DB: %v", eventID, err)
		// On DB error, check if it exists to decide
		return !w.isProcessedInDB(eventID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[IDEMPOTENCY] Failed to get rows affected for event %s: %v", eventID, err)
		return false
	}

	if rowsAffected > 0 {
		log.Printf("[IDEMPOTENCY] Successfully claimed event %s", eventID)
		return true
	}

	log.Printf("[IDEMPOTENCY] Event %s already claimed by another process", eventID)
	return false
}

// MarkEventComplete updates the event status to processed after successful handling
func (w *WebhookIdempotency) MarkEventComplete(eventID string) {
	if w.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		query := `UPDATE payment.webhook_events SET status = 'processed' WHERE event_id = $1`
		_, err := w.db.ExecContext(ctx, query, eventID)
		if err != nil {
			log.Printf("[IDEMPOTENCY] Failed to mark event %s as complete: %v", eventID, err)
		}
	}
}

// MarkEventFailed updates the event status to failed for retry
func (w *WebhookIdempotency) MarkEventFailed(eventID string, errorMsg string) {
	if w.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		query := `UPDATE payment.webhook_events
			SET status = 'failed', error_message = $2
			WHERE event_id = $1`
		_, err := w.db.ExecContext(ctx, query, eventID, errorMsg)
		if err != nil {
			log.Printf("[IDEMPOTENCY] Failed to mark event %s as failed: %v", eventID, err)
		}
	}

	// Remove from cache so retry can be attempted
	w.mutex.Lock()
	delete(w.cache, eventID)
	w.mutex.Unlock()
}

// isProcessedInDB checks the database for processed events
func (w *WebhookIdempotency) isProcessedInDB(eventID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var isProcessed bool
	query := `SELECT EXISTS(
		SELECT 1 FROM payment.webhook_events
		WHERE event_id = $1
		AND processed_at > NOW() - INTERVAL '24 hours'
	)`

	err := w.db.QueryRowContext(ctx, query, eventID).Scan(&isProcessed)
	if err != nil {
		log.Printf("[IDEMPOTENCY] Failed to check DB for event %s: %v", eventID, err)
		return false // Fail open - better to risk duplicate than block
	}

	// If found in DB, add to cache
	if isProcessed {
		w.mutex.Lock()
		w.cache[eventID] = time.Now()
		w.mutex.Unlock()
	}

	return isProcessed
}

// MarkAsProcessed marks an event as processed in both cache and database
func (w *WebhookIdempotency) MarkAsProcessed(eventID string) {
	w.MarkAsProcessedWithType(eventID, "unknown")
}

// MarkAsProcessedWithType marks an event as processed with its event type
func (w *WebhookIdempotency) MarkAsProcessedWithType(eventID, eventType string) {
	// Update in-memory cache first (fast)
	w.mutex.Lock()
	if len(w.cache) >= w.maxCacheSize {
		w.removeOldestEntries(w.maxCacheSize / 4)
	}
	w.cache[eventID] = time.Now()
	w.mutex.Unlock()

	// Then persist to database (async for performance)
	if w.db != nil {
		go w.markProcessedInDB(eventID, eventType)
	}
}

// markProcessedInDB persists the processed event to database
func (w *WebhookIdempotency) markProcessedInDB(eventID, eventType string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO payment.webhook_events (event_id, event_type, processed_at, status)
		VALUES ($1, $2, NOW(), 'processed')
		ON CONFLICT (event_id) DO NOTHING`

	_, err := w.db.ExecContext(ctx, query, eventID, eventType)
	if err != nil {
		log.Printf("[IDEMPOTENCY] Failed to persist event %s to DB: %v", eventID, err)
	}
}

// removeOldestEntries removes the oldest entries (caller must hold write lock)
func (w *WebhookIdempotency) removeOldestEntries(count int) {
	if count <= 0 {
		return
	}

	type entry struct {
		id   string
		time time.Time
	}

	var entries []entry
	for id, t := range w.cache {
		entries = append(entries, entry{id: id, time: t})
	}

	// Sort by time (oldest first) using simple bubble sort
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].time.After(entries[j].time) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	removeCount := count
	if removeCount > len(entries) {
		removeCount = len(entries)
	}

	for i := 0; i < removeCount; i++ {
		delete(w.cache, entries[i].id)
	}
}

// cleanupLoop periodically removes old entries from cache
func (w *WebhookIdempotency) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		w.cleanup()
	}
}

// cleanup removes expired entries from cache
func (w *WebhookIdempotency) cleanup() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	cutoff := time.Now().Add(-w.maxCacheAge)

	for eventID, processedAt := range w.cache {
		if processedAt.Before(cutoff) {
			delete(w.cache, eventID)
		}
	}
}

// dbCleanupLoop periodically removes old entries from database
func (w *WebhookIdempotency) dbCleanupLoop() {
	// Run cleanup once a day
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		w.dbCleanup()
	}
}

// dbCleanup removes old entries from database (older than 7 days)
func (w *WebhookIdempotency) dbCleanup() {
	if w.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `DELETE FROM payment.webhook_events WHERE processed_at < NOW() - INTERVAL '7 days'`

	result, err := w.db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("[IDEMPOTENCY] Failed to cleanup old DB entries: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[IDEMPOTENCY] Cleaned up %d old webhook events from database", rowsAffected)
	}
}

// GetStats returns statistics about the idempotency tracker
func (w *WebhookIdempotency) GetStats() map[string]interface{} {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	stats := map[string]interface{}{
		"cache_size":       len(w.cache),
		"max_cache_size":   w.maxCacheSize,
		"max_age_minutes":  w.maxCacheAge.Minutes(),
		"database_enabled": w.db != nil,
	}

	return stats
}
