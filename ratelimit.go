package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/render"
)

type RateLimiter struct {
	mu       sync.Mutex
	tokens   map[string]float64
	lastSeen map[string]time.Time
	rate     float64
	burst    float64
	cleanup  time.Duration
}

func NewRateLimiter(rate, burst float64) *RateLimiter {
	rl := &RateLimiter{
		tokens:   make(map[string]float64),
		lastSeen: make(map[string]time.Time),
		rate:     rate,
		burst:    burst,
		cleanup:  5 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, t := range rl.lastSeen {
			if now.Sub(t) > rl.cleanup {
				delete(rl.tokens, k)
				delete(rl.lastSeen, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	if _, exists := rl.tokens[key]; !exists {
		rl.tokens[key] = rl.burst
		rl.lastSeen[key] = now
		return true
	}
	elapsed := now.Sub(rl.lastSeen[key]).Seconds()
	rl.tokens[key] += elapsed * rl.rate
	if rl.tokens[key] > rl.burst {
		rl.tokens[key] = rl.burst
	}
	rl.lastSeen[key] = now
	if rl.tokens[key] >= 1 {
		rl.tokens[key]--
		return true
	}
	return false
}

func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetIP(r)
			if !limiter.Allow(ip) {
				render.Status(r, http.StatusTooManyRequests)
				render.JSON(w, r, map[string]string{"error": "too many requests"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
