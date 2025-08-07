package ratelimit

import (
	"context"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for multiple keys
type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	limit    rate.Limit
	burst    int
	cleanup  time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps float64, burst int, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(rps),
		burst:    burst,
		cleanup:  cleanup,
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	limiter := rl.getLimiter(key)
	return limiter.Allow()
}

// Wait waits for permission to proceed
func (rl *RateLimiter) Wait(ctx context.Context, key string) error {
	limiter := rl.getLimiter(key)
	return limiter.Wait(ctx)
}

// Reserve reserves a token for future use
func (rl *RateLimiter) Reserve(key string) *rate.Reservation {
	limiter := rl.getLimiter(key)
	return limiter.Reserve()
}

// getLimiter returns the rate limiter for a key, creating it if necessary
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := rl.limiters[key]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rl.limit, rl.burst)
	rl.limiters[key] = limiter
	return limiter
}

// cleanupRoutine periodically removes unused rate limiters
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.performCleanup()
	}
}

// performCleanup removes inactive rate limiters
func (rl *RateLimiter) performCleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// In a real implementation, you'd track last access time
	// For now, we'll skip cleanup to keep it simple
}

// Middleware for HTTP rate limiting
func (rl *RateLimiter) HTTPMiddleware(keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if key == "" {
				key = r.RemoteAddr
			}

			if !rl.Allow(key) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPRateLimiter is a specialized rate limiter for IP addresses
type IPRateLimiter struct {
	*RateLimiter
}

// NewIPRateLimiter creates a rate limiter for IP addresses
func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		RateLimiter: NewRateLimiter(rps, burst, 10*time.Minute),
	}
}

// AllowIP checks if a request from the given IP is allowed
func (rl *IPRateLimiter) AllowIP(ip string) bool {
	return rl.Allow(ip)
}

// ConnectionRateLimiter limits connection attempts
type ConnectionRateLimiter struct {
	*RateLimiter
	maxConnections int
	connections    map[string]int
	connMu         sync.RWMutex
}

// NewConnectionRateLimiter creates a rate limiter for connections
func NewConnectionRateLimiter(rps float64, burst int, maxConn int) *ConnectionRateLimiter {
	return &ConnectionRateLimiter{
		RateLimiter:    NewRateLimiter(rps, burst, 10*time.Minute),
		maxConnections: maxConn,
		connections:    make(map[string]int),
	}
}

// AllowConnection checks if a new connection is allowed
func (rl *ConnectionRateLimiter) AllowConnection(key string) bool {
	// Check rate limit first
	if !rl.Allow(key) {
		return false
	}

	// Check connection count
	rl.connMu.RLock()
	count := rl.connections[key]
	rl.connMu.RUnlock()

	if count >= rl.maxConnections {
		return false
	}

	// Increment connection count
	rl.connMu.Lock()
	rl.connections[key]++
	rl.connMu.Unlock()

	return true
}

// ReleaseConnection decrements the connection count
func (rl *ConnectionRateLimiter) ReleaseConnection(key string) {
	rl.connMu.Lock()
	defer rl.connMu.Unlock()

	if count, exists := rl.connections[key]; exists && count > 0 {
		rl.connections[key]--
		if rl.connections[key] == 0 {
			delete(rl.connections, key)
		}
	}
}
