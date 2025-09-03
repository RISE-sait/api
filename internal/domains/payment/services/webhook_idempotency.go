package payment

import (
	"sync"
	"time"
)

// WebhookIdempotency provides in-memory idempotency protection for webhooks
type WebhookIdempotency struct {
	processedEvents map[string]time.Time
	mutex           sync.RWMutex
	maxAge          time.Duration
	maxSize         int
}

// NewWebhookIdempotency creates a new webhook idempotency tracker
func NewWebhookIdempotency(maxAge time.Duration, maxSize int) *WebhookIdempotency {
	idempotency := &WebhookIdempotency{
		processedEvents: make(map[string]time.Time),
		maxAge:          maxAge,
		maxSize:         maxSize,
	}
	
	// Start cleanup goroutine
	go idempotency.cleanupLoop()
	
	return idempotency
}

// IsProcessed checks if an event has already been processed
func (w *WebhookIdempotency) IsProcessed(eventID string) bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	processedAt, exists := w.processedEvents[eventID]
	if !exists {
		return false
	}
	
	// Check if the event is too old (stale)
	if time.Since(processedAt) > w.maxAge {
		return false
	}
	
	return true
}

// MarkAsProcessed marks an event as processed
func (w *WebhookIdempotency) MarkAsProcessed(eventID string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	// If we're at max capacity, remove oldest entries
	if len(w.processedEvents) >= w.maxSize {
		w.removeOldestEntries(w.maxSize / 4) // Remove 25% of entries
	}
	
	w.processedEvents[eventID] = time.Now()
}

// removeOldestEntries removes the oldest entries (caller must hold write lock)
func (w *WebhookIdempotency) removeOldestEntries(count int) {
	if count <= 0 {
		return
	}
	
	// Find oldest entries
	type entry struct {
		id   string
		time time.Time
	}
	
	var entries []entry
	for id, t := range w.processedEvents {
		entries = append(entries, entry{id: id, time: t})
	}
	
	// Sort by time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].time.After(entries[j].time) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	
	// Remove oldest entries
	removeCount := count
	if removeCount > len(entries) {
		removeCount = len(entries)
	}
	
	for i := 0; i < removeCount; i++ {
		delete(w.processedEvents, entries[i].id)
	}
}

// cleanupLoop periodically removes old entries
func (w *WebhookIdempotency) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()
	
	for range ticker.C {
		w.cleanup()
	}
}

// cleanup removes expired entries
func (w *WebhookIdempotency) cleanup() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	cutoff := time.Now().Add(-w.maxAge)
	
	for eventID, processedAt := range w.processedEvents {
		if processedAt.Before(cutoff) {
			delete(w.processedEvents, eventID)
		}
	}
}

// GetStats returns statistics about the idempotency tracker
func (w *WebhookIdempotency) GetStats() map[string]interface{} {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	return map[string]interface{}{
		"total_events":     len(w.processedEvents),
		"max_size":         w.maxSize,
		"max_age_minutes":  w.maxAge.Minutes(),
	}
}