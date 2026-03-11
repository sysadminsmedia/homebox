package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func TestSimpleRateLimiter(t *testing.T) {
	type testCase struct {
		name       string
		trustProxy bool
		setupReq   func(*http.Request, string)
	}

	tests := []testCase{
		{
			name:       "DirectConnection",
			trustProxy: false,
			setupReq:   func(r *http.Request, ip string) { r.RemoteAddr = ip + ":1234" },
		},
		{
			name:       "ProxyXRealIP",
			trustProxy: true,
			setupReq: func(r *http.Request, ip string) {
				r.RemoteAddr = "10.0.0.1:1234" // Proxy IP
				r.Header.Set("X-Real-IP", ip)
			},
		},
		{
			name:       "ProxyXForwardedFor",
			trustProxy: true,
			setupReq: func(r *http.Request, ip string) {
				r.RemoteAddr = "10.0.0.1:1234" // Proxy IP
				r.Header.Set("X-Forwarded-For", ip+", 10.0.0.2")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a rate limiter that allows 3 requests per 10 seconds
			limiter := newSimpleRateLimiter(3, 10*time.Second, tc.trustProxy)
			clientIP := "192.168.1.1"

			// Helper to get IP
			getIP := func(ip string) string {
				req := httptest.NewRequest("GET", "/", nil)
				tc.setupReq(req, ip)
				return limiter.getClientIP(req, limiter.trustProxy)
			}

			// First 3 requests should succeed
			for i := 0; i < 3; i++ {
				ip := getIP(clientIP)
				if !limiter.allow(ip) {
					t.Errorf("Request %d should have been allowed", i+1)
				}
			}

			// 4th request should be blocked
			ip := getIP(clientIP)
			if limiter.allow(ip) {
				t.Error("4th request should have been blocked")
			}

			// Different client should not be affected
			otherClient := "192.168.1.2"
			otherIP := getIP(otherClient)
			// Check if we are really testing a different IP
			if otherIP == ip {
				// This might happen if setupReq implementation is wrong or trustProxy logic is broken
				t.Fatalf("Test setup error: otherClient IP resolved to same as clientIP: %s", otherIP)
			}

			if !limiter.allow(otherIP) {
				t.Error("Different client should be allowed")
			}
		})
	}
}

func TestSimpleRateLimiterRefill(t *testing.T) {
	type testCase struct {
		name       string
		trustProxy bool
		setupReq   func(*http.Request, string)
	}

	tests := []testCase{
		{
			name:       "DirectConnection",
			trustProxy: false,
			setupReq:   func(r *http.Request, ip string) { r.RemoteAddr = ip + ":1234" },
		},
		{
			name:       "ProxyXRealIP",
			trustProxy: true,
			setupReq: func(r *http.Request, ip string) {
				r.RemoteAddr = "10.0.0.1:1234" // Proxy IP
				r.Header.Set("X-Real-IP", ip)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a rate limiter that allows 2 requests per 100ms
			limiter := newSimpleRateLimiter(2, 100*time.Millisecond, tc.trustProxy)
			clientIP := "192.168.1.1"

			// Helper to get IP
			getIP := func(ip string) string {
				req := httptest.NewRequest("GET", "/", nil)
				tc.setupReq(req, ip)
				return limiter.getClientIP(req, limiter.trustProxy)
			}

			ip := getIP(clientIP)

			// Use up the tokens
			if !limiter.allow(ip) {
				t.Error("First request should be allowed")
			}
			if !limiter.allow(ip) {
				t.Error("Second request should be allowed")
			}
			if limiter.allow(ip) {
				t.Error("Third request should be blocked")
			}

			// Wait for refill
			time.Sleep(150 * time.Millisecond)

			// Should be allowed again after refill
			if !limiter.allow(ip) {
				t.Error("Request after refill should be allowed")
			}
		})
	}
}

func TestSimpleRateLimiterConcurrent(t *testing.T) {
	type testCase struct {
		name       string
		trustProxy bool
		setupReq   func(*http.Request, string)
	}

	tests := []testCase{
		{
			name:       "DirectConnection",
			trustProxy: false,
			setupReq:   func(r *http.Request, ip string) { r.RemoteAddr = ip + ":1234" },
		},
		{
			name:       "ProxyXRealIP",
			trustProxy: true,
			setupReq: func(r *http.Request, ip string) {
				r.RemoteAddr = "10.0.0.1:1234" // Proxy IP
				r.Header.Set("X-Real-IP", ip)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			limiter := newSimpleRateLimiter(10, time.Second, tc.trustProxy)
			clientIP := "192.168.1.1"

			// Helper to get IP
			getIP := func(ip string) string {
				req := httptest.NewRequest("GET", "/", nil)
				tc.setupReq(req, ip)
				return limiter.getClientIP(req, limiter.trustProxy)
			}

			ip := getIP(clientIP)

			var allowed int32
			done := make(chan bool)

			// Spawn multiple goroutines trying to access the limiter
			for i := 0; i < 20; i++ {
				go func() {
					if limiter.allow(ip) {
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
		})
	}
}

func TestSimpleRateLimiterCleanup(t *testing.T) {
	type testCase struct {
		name       string
		trustProxy bool
		setupReq   func(*http.Request, string)
	}

	tests := []testCase{
		{
			name:       "DirectConnection",
			trustProxy: false,
			setupReq:   func(r *http.Request, ip string) { r.RemoteAddr = ip + ":1234" },
		},
		{
			name:       "ProxyXRealIP",
			trustProxy: true,
			setupReq: func(r *http.Request, ip string) {
				r.RemoteAddr = "10.0.0.1:1234" // Proxy IP
				r.Header.Set("X-Real-IP", ip)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a rate limiter with a short window
			limiter := newSimpleRateLimiter(5, 100*time.Millisecond, tc.trustProxy)
			defer limiter.Stop()

			// Helper to get IP
			getIP := func(ip string) string {
				req := httptest.NewRequest("GET", "/", nil)
				tc.setupReq(req, ip)
				return limiter.getClientIP(req, limiter.trustProxy)
			}

			// Add entries for multiple IPs
			ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5"}
			for _, ip := range ips {
				limiter.allow(getIP(ip))
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
		})
	}
}

func TestSimpleRateLimiterCleanupPreservesActive(t *testing.T) {
	type testCase struct {
		name       string
		trustProxy bool
		setupReq   func(*http.Request, string)
	}

	tests := []testCase{
		{
			name:       "DirectConnection",
			trustProxy: false,
			setupReq:   func(r *http.Request, ip string) { r.RemoteAddr = ip + ":1234" },
		},
		{
			name:       "ProxyXRealIP",
			trustProxy: true,
			setupReq: func(r *http.Request, ip string) {
				r.RemoteAddr = "10.0.0.1:1234" // Proxy IP
				r.Header.Set("X-Real-IP", ip)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			limiter := newSimpleRateLimiter(5, 100*time.Millisecond, tc.trustProxy)
			defer limiter.Stop()

			// Helper to get IP
			getIP := func(ip string) string {
				req := httptest.NewRequest("GET", "/", nil)
				tc.setupReq(req, ip)
				return limiter.getClientIP(req, limiter.trustProxy)
			}

			activeIP := "192.168.1.1"
			staleIP := "192.168.1.2"

			// Create a stale entry
			limiter.allow(getIP(staleIP))

			// Wait for it to become stale
			time.Sleep(250 * time.Millisecond)

			// Create an active entry
			limiter.allow(getIP(activeIP))

			// Trigger cleanup
			limiter.cleanup()

			// Verify active entry is preserved and stale is removed
			limiter.mu.Lock()
			defer limiter.mu.Unlock()

			// Check using the resolved IP keys
			activeKey := getIP(activeIP)
			staleKey := getIP(staleIP)

			if _, exists := limiter.limiters[activeKey]; !exists {
				t.Error("Active entry should be preserved")
			}

			if _, exists := limiter.limiters[staleKey]; exists {
				t.Error("Stale entry should be removed")
			}
		})
	}
}

func TestSimpleRateLimiterStop(t *testing.T) {
	limiter := newSimpleRateLimiter(5, time.Second, false)

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
