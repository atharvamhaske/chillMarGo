package ratelimiter

import (
	"sync"
	"testing"
	"time"
)

func TestLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewLimiter(100, 10)
	key := "concurrent-key"

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Launch 200 concurrent requests (should only allow 100)
	numRequests := 200
	numWorkers := 20

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numRequests/numWorkers; j++ {
				if limiter.Allow(key) {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Should have allowed exactly capacity (100)
	if successCount != 100 {
		t.Errorf("expected 100 successful requests, got %d", successCount)
	}

	// Next request should fail
	if limiter.Allow(key) {
		t.Error("expected request to fail after capacity exhausted")
	}
}

func TestLimiter_TimeBasedRefill(t *testing.T) {
	// Create a limiter with fast refill for testing
	capacity := 5
	refillRate := 2 // 2 tokens per second

	limiter := NewLimiter(capacity, refillRate)
	key := "refill-test-key"

	// Exhaust all tokens
	for i := 0; i < capacity; i++ {
		if !limiter.Allow(key) {
			t.Errorf("expected Allow to succeed for request %d", i+1)
		}
	}

	// Should be exhausted
	if limiter.Allow(key) {
		t.Error("expected Allow to fail when exhausted")
	}

	// Wait for refill (1 second should add 2 tokens)
	time.Sleep(1100 * time.Millisecond)

	// Should now allow 2 more requests
	allowed := 0
	for i := 0; i < capacity; i++ {
		if limiter.Allow(key) {
			allowed++
		}
	}

	if allowed != refillRate {
		t.Errorf("expected %d tokens after 1 second, got %d", refillRate, allowed)
	}
}

func TestLimiter_MultipleKeysIsolation(t *testing.T) {
	limiter := NewLimiter(5, 1)

	keys := []string{"key1", "key2", "key3"}

	// Each key should independently allow 5 requests
	for _, key := range keys {
		for i := 0; i < 5; i++ {
			if !limiter.Allow(key) {
				t.Errorf("expected Allow to succeed for %s, request %d", key, i+1)
			}
		}

		// Each should be exhausted independently
		if limiter.Allow(key) {
			t.Errorf("expected Allow to fail for exhausted key %s", key)
		}
	}
}

func TestLimiter_Snapshot_AfterRefill(t *testing.T) {
	limiter := NewLimiter(10, 2)
	key := "snapshot-key"

	// Exhaust tokens
	for i := 0; i < 10; i++ {
		limiter.Allow(key)
	}

	remaining, limit, retryAfter := limiter.Snapshot(key)
	if remaining != 0 {
		t.Errorf("expected remaining 0, got %d", remaining)
	}
	if retryAfter == 0 {
		t.Error("expected retryAfter > 0 when exhausted")
	}

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)

	remaining, limit, retryAfter = limiter.Snapshot(key)
	if remaining < 2 {
		t.Errorf("expected at least 2 tokens after refill, got %d", remaining)
	}
	if limit != 10 {
		t.Errorf("expected limit 10, got %d", limit)
	}
	if retryAfter != 0 {
		t.Errorf("expected retryAfter 0 when tokens available, got %d", retryAfter)
	}
}

func TestTokenBucketAlgo_Integration(t *testing.T) {
	limiter := NewLimiter(3, 1)
	algo := TokenBucketAlgo(limiter)

	key := "algo-test-key"

	// First 3 should succeed
	successCount := 0
	for i := 0; i < 5; i++ {
		if err := algo(key); err == nil {
			successCount++
		}
	}

	if successCount != 3 {
		t.Errorf("expected 3 successful requests, got %d", successCount)
	}

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)

	// Should allow 1 more
	if err := algo(key); err != nil {
		t.Errorf("expected success after refill, got %v", err)
	}
}

func TestLimiter_HighFrequencyRequests(t *testing.T) {
	limiter := NewLimiter(100, 10)
	key := "high-freq-key"

	// Rapid fire requests
	successCount := 0
	start := time.Now()
	for i := 0; i < 200; i++ {
		if limiter.Allow(key) {
			successCount++
		}
		// Small delay to simulate real requests
		time.Sleep(1 * time.Millisecond)
	}
	duration := time.Since(start)

	// Should have allowed exactly capacity
	if successCount != 100 {
		t.Errorf("expected 100 successful requests, got %d", successCount)
	}

	// Should have taken some time (not instant)
	if duration < 50*time.Millisecond {
		t.Errorf("expected some time to pass, got %v", duration)
	}

	t.Logf("Processed 200 requests in %v, allowed %d", duration, successCount)
}

func TestLimiter_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	limiter := NewLimiter(1000, 100)
	key := "stress-key"

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Launch many goroutines making many requests
	numGoroutines := 50
	requestsPerGoroutine := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				if limiter.Allow(key) {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Should have allowed exactly capacity
	if successCount != 1000 {
		t.Errorf("expected 1000 successful requests, got %d", successCount)
	}

	// Verify snapshot
	remaining, limit, _ := limiter.Snapshot(key)
	if remaining != 0 {
		t.Errorf("expected 0 remaining after stress test, got %d", remaining)
	}
	if limit != 1000 {
		t.Errorf("expected limit 1000, got %d", limit)
	}
}
