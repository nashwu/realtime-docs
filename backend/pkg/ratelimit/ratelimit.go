package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Limiter is a simple token bucket keyed by client IP
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket // per-IP buckets
	max     int                // tokens per window
	per     time.Duration      // window size
}

type bucket struct {
	ts     time.Time // window start
	tokens int       // remaining tokens
}

// New creates a new IP-based limiter allowing max requests per window
func New(max int, per time.Duration) *Limiter {
	return &Limiter{buckets: map[string]*bucket{}, max: max, per: per}
}

// Middleware enforces the rate limit before calling the next handler
func (r *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ip, _, _ := net.SplitHostPort(req.RemoteAddr)

		r.mu.Lock()
		b := r.buckets[ip]
		if b == nil || time.Since(b.ts) > r.per {
			// Start a new window
			b = &bucket{ts: time.Now(), tokens: r.max}
			r.buckets[ip] = b
		}

		if b.tokens <= 0 {
			r.mu.Unlock()
			http.Error(w, "rate limit", http.StatusTooManyRequests)
			return
		}

		b.tokens--
		r.mu.Unlock()

		next.ServeHTTP(w, req)
	})
}
