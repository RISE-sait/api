package payment

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ============================================================
// Fix 10: safeGo — goroutine panic recovery
// ============================================================

func TestSafeGo_NormalExecution(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	executed := false
	safeGo("test-normal", func() {
		defer wg.Done()
		executed = true
	})

	wg.Wait()
	assert.True(t, executed, "function should have executed")
}

func TestSafeGo_PanicRecovery(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	// This should NOT crash the test — safeGo should recover the panic
	safeGo("test-panic", func() {
		defer wg.Done()
		panic("deliberate test panic")
	})

	wg.Wait()
	// If we get here, the panic was recovered. Test passes.
}

func TestSafeGo_PanicDoesNotAffectOtherGoroutines(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	result := make(chan string, 2)

	// First goroutine panics
	safeGo("panicker", func() {
		defer wg.Done()
		result <- "panicker-started"
		panic("boom")
	})

	// Second goroutine should still complete
	safeGo("worker", func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // small delay to ensure panicker runs first
		result <- "worker-done"
	})

	wg.Wait()
	close(result)

	messages := make([]string, 0)
	for msg := range result {
		messages = append(messages, msg)
	}

	assert.Contains(t, messages, "worker-done", "worker goroutine should complete despite panicker")
}

func TestSafeGo_NilPanicValue(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	// panic(nil) in Go 1.21+ causes a *runtime.PanicNilError
	safeGo("nil-panic", func() {
		defer wg.Done()
		panic(nil)
	})

	wg.Wait()
	// Should not crash
}

func TestSafeGo_MultipleConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	count := 0
	var mu sync.Mutex

	for i := range 20 {
		wg.Add(1)
		safeGo("concurrent-"+string(rune('a'+i)), func() {
			defer wg.Done()
			mu.Lock()
			count++
			mu.Unlock()
		})
	}

	wg.Wait()
	assert.Equal(t, 20, count, "all 20 goroutines should complete")
}
