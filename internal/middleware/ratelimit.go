package middleware

import (
	"net/http"
	"sync"
	"time"

	"curanmor-ai/internal/httpx"
)

// RateLimiter adalah pembatas laju sederhana berbasis sliding window per-IP
// (in-memory). Cukup untuk melindungi endpoint sensitif (mis. /auth/login)
// dari brute force pada single-instance deployment. Untuk multi-instance,
// ganti dengan Redis.
type RateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{hits: make(map[string][]time.Time), limit: limit, window: window}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for k, times := range rl.hits {
			kept := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					kept = append(kept, t)
				}
			}
			if len(kept) == 0 {
				delete(rl.hits, k)
			} else {
				rl.hits[k] = kept
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)
	times := rl.hits[key]
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= rl.limit {
		rl.hits[key] = kept
		return false
	}
	kept = append(kept, now)
	rl.hits[key] = kept
	return true
}

func (rl *RateLimiter) Middleware() httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := httpx.ClientIP(r)
			if !rl.Allow(ip) {
				httpx.TooManyRequests(w, "Terlalu banyak percobaan, coba lagi beberapa saat lagi")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
