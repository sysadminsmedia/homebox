package main

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func TestSimpleRateLimiter(t *testing.T) {
	// Create a rate limiter that allows 3 requests per 10 seconds
	limiter := newSimpleRateLimiter(3, 10*time.Second)
	clientIP := "192.168.1.1"

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		if !limiter.allow(clientIP) {
			t.Errorf("Request %d should have been allowed", i+1)
		}
	}

	// 4th request should be blocked
	if limiter.allow(clientIP) {
		t.Error("4th request should have been blocked")
	}

	// Different client should not be affected
	otherClient := "192.168.1.2"
	if !limiter.allow(otherClient) {
		t.Error("Different client should be allowed")
	}
}

func TestSimpleRateLimiterRefill(t *testing.T) {
	// Create a rate limiter that allows 2 requests per 100ms
	limiter := newSimpleRateLimiter(2, 100*time.Millisecond)
	clientIP := "192.168.1.1"

	// Use up the tokens
	if !limiter.allow(clientIP) {
		t.Error("First request should be allowed")
	}
	if !limiter.allow(clientIP) {
		t.Error("Second request should be allowed")
	}
	if limiter.allow(clientIP) {
		t.Error("Third request should be blocked")
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again after refill
	if !limiter.allow(clientIP) {
		t.Error("Request after refill should be allowed")
	}
}

func TestSimpleRateLimiterConcurrent(t *testing.T) {
	limiter := newSimpleRateLimiter(10, time.Second)
	clientIP := "192.168.1.1"

	var allowed int32
	done := make(chan bool)

	// Spawn multiple goroutines trying to access the limiter
	for i := 0; i < 20; i++ {
		go func() {
			if limiter.allow(clientIP) {
				atomic.AddInt32(&allowed, 1)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should allow exactly 10 requests
	allowedCount := atomic.LoadInt32(&allowed)
	if allowedCount != 10 {
		t.Errorf("Expected 10 allowed requests, got %d", allowedCount)
	}
}

func TestSimpleRateLimiterCleanup(t *testing.T) {
	// Create a rate limiter with a short window
	limiter := newSimpleRateLimiter(5, 100*time.Millisecond)
	defer limiter.Stop()

	// Add entries for multiple IPs
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5"}
	for _, ip := range ips {
		limiter.allow(ip)
	}

	// Verify entries exist
	limiter.mu.Lock()
	initialCount := len(limiter.limiters)
	limiter.mu.Unlock()

	if initialCount != len(ips) {
		t.Errorf("Expected %d entries, got %d", len(ips), initialCount)
	}

	// Wait for stale threshold (2x window = 200ms + buffer)
	time.Sleep(250 * time.Millisecond)

	// Trigger cleanup manually
	limiter.cleanup()

	// Verify stale entries are removed
	limiter.mu.Lock()
	finalCount := len(limiter.limiters)
	limiter.mu.Unlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", finalCount)
	}
}

func TestSimpleRateLimiterCleanupPreservesActive(t *testing.T) {
	limiter := newSimpleRateLimiter(5, 100*time.Millisecond)
	defer limiter.Stop()

	activeIP := "192.168.1.1"
	staleIP := "192.168.1.2"

	// Create a stale entry
	limiter.allow(staleIP)

	// Wait for it to become stale
	time.Sleep(250 * time.Millisecond)

	// Create an active entry
	limiter.allow(activeIP)

	// Trigger cleanup
	limiter.cleanup()

	// Verify active entry is preserved and stale is removed
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if _, exists := limiter.limiters[activeIP]; !exists {
		t.Error("Active entry should be preserved")
	}

	if _, exists := limiter.limiters[staleIP]; exists {
		t.Error("Stale entry should be removed")
	}
}

func TestSimpleRateLimiterStop(t *testing.T) {
	limiter := newSimpleRateLimiter(5, time.Second)

	// Stop the limiter
	limiter.Stop()

	// Verify the cleanup goroutine exits (this test passes if no panic occurs)
	time.Sleep(10 * time.Millisecond)
}

func TestAuthRateLimiterCleanup(t *testing.T) {
	cfg := config.AuthRateLimit{
		Enabled:     true,
		MaxAttempts: 3,
		BaseBackoff: 10 * time.Millisecond,
		MaxBackoff:  100 * time.Millisecond,
		Window:      50 * time.Millisecond,
	}

	limiter := newAuthRateLimiter(cfg)
	defer limiter.Stop()

	// Add multiple failed attempts for different keys
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	now := time.Now()

	for _, key := range keys {
		limiter.record(key, now, false)
	}

	// Verify entries exist
	limiter.mu.Lock()
	initialCount := len(limiter.state)
	limiter.mu.Unlock()

	if initialCount != len(keys) {
		t.Errorf("Expected %d entries, got %d", len(keys), initialCount)
	}

	// Wait for entries to become stale (window = 50ms + buffer)
	time.Sleep(100 * time.Millisecond)

	// Trigger cleanup
	limiter.cleanup()

	// Verify stale entries are removed
	limiter.mu.Lock()
	finalCount := len(limiter.state)
	limiter.mu.Unlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", finalCount)
	}
}

func TestAuthRateLimiterCleanupPreservesLocked(t *testing.T) {
	cfg := config.AuthRateLimit{
		Enabled:     true,
		MaxAttempts: 2,
		BaseBackoff: 200 * time.Millisecond,
		MaxBackoff:  1 * time.Second,
		Window:      50 * time.Millisecond,
	}

	limiter := newAuthRateLimiter(cfg)
	defer limiter.Stop()

	lockedKey := "locked"
	staleKey := "stale"
	now := limiter.nowFn()

	// Create a locked entry (exceed max attempts)
	for i := 0; i < cfg.MaxAttempts+1; i++ {
		limiter.record(lockedKey, now, false)
	}

	// Create a stale entry
	limiter.record(staleKey, now, false)

	// Wait for entries to be outside the window but locked entry still locked
	time.Sleep(100 * time.Millisecond)

	// Trigger cleanup
	limiter.cleanup()

	// Verify locked entry is preserved and stale is removed
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if _, exists := limiter.state[lockedKey]; !exists {
		t.Error("Locked entry should be preserved during lockout period")
	}

	if _, exists := limiter.state[staleKey]; exists {
		t.Error("Stale entry should be removed")
	}
}

func TestAuthRateLimiterStop(t *testing.T) {
	cfg := config.AuthRateLimit{
		Enabled:     true,
		MaxAttempts: 5,
		BaseBackoff: 10 * time.Millisecond,
		MaxBackoff:  100 * time.Millisecond,
		Window:      time.Second,
	}

	limiter := newAuthRateLimiter(cfg)

	// Stop the limiter
	limiter.Stop()

	// Verify the cleanup goroutine exits (this test passes if no panic occurs)
	time.Sleep(10 * time.Millisecond)
}
