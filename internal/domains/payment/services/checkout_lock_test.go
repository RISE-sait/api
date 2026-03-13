package payment

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Fix 4: Double-click checkout prevention
// ============================================================

func newTestService() *Service {
	return &Service{}
}

func TestCheckoutLock_AcquireAndRelease(t *testing.T) {
	svc := newTestService()
	customerID := uuid.New()
	itemID := uuid.New()

	// First acquire should succeed
	err := svc.tryAcquireCheckoutLock(customerID, "membership", itemID)
	assert.Nil(t, err, "first lock acquisition should succeed")

	// Release
	svc.releaseCheckoutLock(customerID, "membership", itemID)

	// Should be able to acquire again after release
	err = svc.tryAcquireCheckoutLock(customerID, "membership", itemID)
	assert.Nil(t, err, "lock should be acquirable after release")

	svc.releaseCheckoutLock(customerID, "membership", itemID)
}

func TestCheckoutLock_DuplicateBlocked(t *testing.T) {
	svc := newTestService()
	customerID := uuid.New()
	itemID := uuid.New()

	err := svc.tryAcquireCheckoutLock(customerID, "membership", itemID)
	require.Nil(t, err)

	// Second acquire same key should fail
	err = svc.tryAcquireCheckoutLock(customerID, "membership", itemID)
	require.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, err.HTTPCode)
	assert.Contains(t, err.Error(), "already in progress")

	svc.releaseCheckoutLock(customerID, "membership", itemID)
}

func TestCheckoutLock_DifferentCustomersSameItem(t *testing.T) {
	svc := newTestService()
	customer1 := uuid.New()
	customer2 := uuid.New()
	itemID := uuid.New()

	err1 := svc.tryAcquireCheckoutLock(customer1, "membership", itemID)
	assert.Nil(t, err1, "customer 1 should get lock")

	err2 := svc.tryAcquireCheckoutLock(customer2, "membership", itemID)
	assert.Nil(t, err2, "customer 2 should get lock (different customer)")

	svc.releaseCheckoutLock(customer1, "membership", itemID)
	svc.releaseCheckoutLock(customer2, "membership", itemID)
}

func TestCheckoutLock_SameCustomerDifferentItems(t *testing.T) {
	svc := newTestService()
	customerID := uuid.New()
	item1 := uuid.New()
	item2 := uuid.New()

	err1 := svc.tryAcquireCheckoutLock(customerID, "membership", item1)
	assert.Nil(t, err1)

	err2 := svc.tryAcquireCheckoutLock(customerID, "membership", item2)
	assert.Nil(t, err2, "same customer, different item should work")

	svc.releaseCheckoutLock(customerID, "membership", item1)
	svc.releaseCheckoutLock(customerID, "membership", item2)
}

func TestCheckoutLock_SameCustomerDifferentTypes(t *testing.T) {
	svc := newTestService()
	customerID := uuid.New()
	itemID := uuid.New()

	err1 := svc.tryAcquireCheckoutLock(customerID, "membership", itemID)
	assert.Nil(t, err1)

	err2 := svc.tryAcquireCheckoutLock(customerID, "event", itemID)
	assert.Nil(t, err2, "same customer+item but different type should work")

	svc.releaseCheckoutLock(customerID, "membership", itemID)
	svc.releaseCheckoutLock(customerID, "event", itemID)
}

func TestCheckoutLock_ExpiredLockOverridden(t *testing.T) {
	svc := newTestService()
	customerID := uuid.New()
	itemID := uuid.New()

	// Manually insert an old lock (simulating a stale lock from a crashed request)
	key := customerID.String() + ":membership:" + itemID.String()
	svc.activeCheckouts.Store(key, &checkoutLock{
		createdAt: time.Now().Add(-15 * time.Minute), // 15 min old — past 10 min expiry
	})

	// Should override the expired lock
	err := svc.tryAcquireCheckoutLock(customerID, "membership", itemID)
	assert.Nil(t, err, "expired lock should be overridden")

	svc.releaseCheckoutLock(customerID, "membership", itemID)
}

func TestCheckoutLock_NotExpiredLockNotOverridden(t *testing.T) {
	svc := newTestService()
	customerID := uuid.New()
	itemID := uuid.New()

	// Insert a recent lock (5 min old — within 10 min window)
	key := customerID.String() + ":membership:" + itemID.String()
	svc.activeCheckouts.Store(key, &checkoutLock{
		createdAt: time.Now().Add(-5 * time.Minute),
	})

	// Should NOT override — still valid
	err := svc.tryAcquireCheckoutLock(customerID, "membership", itemID)
	require.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, err.HTTPCode)
}

func TestCheckoutLock_ConcurrentDoubleClick(t *testing.T) {
	svc := newTestService()
	customerID := uuid.New()
	itemID := uuid.New()

	results := make(chan bool, 100)
	var wg sync.WaitGroup

	// Simulate 100 concurrent checkout attempts (aggressive double-click)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := svc.tryAcquireCheckoutLock(customerID, "membership", itemID)
			results <- (err == nil)
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	for success := range results {
		if success {
			successCount++
		}
	}

	assert.Equal(t, 1, successCount, "exactly one concurrent checkout should succeed")

	svc.releaseCheckoutLock(customerID, "membership", itemID)
}

func TestCheckoutLock_ReleaseNonexistent(t *testing.T) {
	svc := newTestService()

	// Should not panic when releasing a lock that doesn't exist
	assert.NotPanics(t, func() {
		svc.releaseCheckoutLock(uuid.New(), "membership", uuid.New())
	})
}

func TestCheckoutLock_RapidAcquireReleaseCycles(t *testing.T) {
	svc := newTestService()
	customerID := uuid.New()
	itemID := uuid.New()

	// Simulate a customer clicking, getting checkout page, completing, then clicking again
	for i := 0; i < 10; i++ {
		err := svc.tryAcquireCheckoutLock(customerID, "membership", itemID)
		assert.Nil(t, err, "cycle %d: acquire should succeed", i)

		svc.releaseCheckoutLock(customerID, "membership", itemID)
	}
}
