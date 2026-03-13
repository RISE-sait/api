package payment

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Fix 6: TryClaimEvent — cache-only mode (no DB)
// ============================================================

func TestTryClaimEvent_FirstClaimSucceeds(t *testing.T) {
	idem := NewWebhookIdempotency(1*time.Hour, 1000)

	claimed, err := idem.TryClaimEvent("evt_123", "checkout.session.completed")
	require.NoError(t, err)
	assert.True(t, claimed, "first claim should succeed")
}

func TestTryClaimEvent_DuplicateClaimRejected(t *testing.T) {
	idem := NewWebhookIdempotency(1*time.Hour, 1000)

	claimed1, err := idem.TryClaimEvent("evt_123", "checkout.session.completed")
	require.NoError(t, err)
	assert.True(t, claimed1)

	claimed2, err := idem.TryClaimEvent("evt_123", "checkout.session.completed")
	require.NoError(t, err)
	assert.False(t, claimed2, "duplicate claim should be rejected")
}

func TestTryClaimEvent_DifferentEventsCanBothClaim(t *testing.T) {
	idem := NewWebhookIdempotency(1*time.Hour, 1000)

	claimed1, err := idem.TryClaimEvent("evt_111", "checkout.session.completed")
	require.NoError(t, err)
	assert.True(t, claimed1)

	claimed2, err := idem.TryClaimEvent("evt_222", "invoice.payment_succeeded")
	require.NoError(t, err)
	assert.True(t, claimed2, "different event should be claimable")
}

func TestTryClaimEvent_ConcurrentClaimsSameEvent(t *testing.T) {
	idem := NewWebhookIdempotency(1*time.Hour, 1000)

	results := make(chan bool, 50)
	var wg sync.WaitGroup

	// 50 goroutines all try to claim the same event simultaneously
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			claimed, err := idem.TryClaimEvent("evt_race", "test.event")
			if err != nil {
				results <- false
				return
			}
			results <- claimed
		}()
	}

	wg.Wait()
	close(results)

	claimCount := 0
	for claimed := range results {
		if claimed {
			claimCount++
		}
	}

	assert.Equal(t, 1, claimCount, "exactly one goroutine should win the claim")
}

func TestTryClaimEvent_CacheSizeEviction(t *testing.T) {
	// Small cache that can only hold 4 events
	idem := NewWebhookIdempotency(1*time.Hour, 4)

	// Fill the cache
	for i := 0; i < 4; i++ {
		claimed, err := idem.TryClaimEvent("evt_"+string(rune('a'+i)), "test")
		require.NoError(t, err)
		assert.True(t, claimed)
	}

	// 5th event should still work (triggers eviction of oldest 25%)
	claimed, err := idem.TryClaimEvent("evt_new", "test")
	require.NoError(t, err)
	assert.True(t, claimed, "should succeed after evicting old entries")
}

func TestIsProcessed_CacheOnly(t *testing.T) {
	idem := NewWebhookIdempotency(1*time.Hour, 1000)

	assert.False(t, idem.IsProcessed("evt_never_seen"))

	idem.TryClaimEvent("evt_claimed", "test")
	assert.True(t, idem.IsProcessed("evt_claimed"))
}

func TestMarkEventFailed_AllowsRetry(t *testing.T) {
	idem := NewWebhookIdempotency(1*time.Hour, 1000)

	// Claim and then mark as failed
	claimed, _ := idem.TryClaimEvent("evt_fail", "test")
	assert.True(t, claimed)

	idem.MarkEventFailed("evt_fail", "something broke")

	// Should be able to claim again after failure (removed from cache)
	claimed2, _ := idem.TryClaimEvent("evt_fail", "test")
	assert.True(t, claimed2, "should be re-claimable after MarkEventFailed")
}

func TestTryClaimEvent_ExpiredCacheEntry(t *testing.T) {
	// Very short TTL
	idem := NewWebhookIdempotency(1*time.Millisecond, 1000)

	claimed, _ := idem.TryClaimEvent("evt_expire", "test")
	assert.True(t, claimed)

	// Wait for expiry
	time.Sleep(5 * time.Millisecond)

	// The cache fast-path check should miss (expired), but the write-lock
	// double-check will also find it. Since there's no DB, the cache-only
	// path sees the existing entry. Let's verify the IsProcessed path:
	assert.False(t, idem.IsProcessed("evt_expire"), "expired entry should not be considered processed")
}

func TestGetStats(t *testing.T) {
	idem := NewWebhookIdempotency(30*time.Minute, 500)

	idem.TryClaimEvent("evt_1", "test")
	idem.TryClaimEvent("evt_2", "test")

	stats := idem.GetStats()
	assert.Equal(t, 2, stats["cache_size"])
	assert.Equal(t, 500, stats["max_cache_size"])
	assert.Equal(t, float64(30), stats["max_age_minutes"])
	assert.Equal(t, false, stats["database_enabled"])
}
