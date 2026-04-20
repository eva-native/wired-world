package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu          sync.Mutex
	limiters    map[string]*ipLimiter
	r           rate.Limit
	burst       int
	behindProxy bool
}

func NewRateLimiter(r rate.Limit, burst int, behindProxy bool) *RateLimiter {
	return &RateLimiter{
		limiters:    make(map[string]*ipLimiter),
		r:           r,
		burst:       burst,
		behindProxy: behindProxy,
	}
}

func (rl *RateLimiter) get(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	entry, ok := rl.limiters[ip]
	if !ok {
		entry = &ipLimiter{limiter: rate.NewLimiter(rl.r, rl.burst)}
		rl.limiters[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (rl *RateLimiter) Cleanup(ctx context.Context, ttl time.Duration) {
	ticker := time.NewTicker(ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			for ip, entry := range rl.limiters {
				if time.Since(entry.lastSeen) > ttl {
					delete(rl.limiters, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// clientIP extracts the real client IP. When behindProxy is true it reads
// X-Real-IP first, then the leftmost entry of X-Forwarded-For, falling back
// to RemoteAddr. Only enable behindProxy when a trusted proxy sets these headers.
func clientIP(r *http.Request, behindProxy bool) string {
	if behindProxy {
		if ip := r.Header.Get("X-Real-IP"); ip != "" {
			return ip
		}
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			if i := strings.IndexByte(fwd, ','); i != -1 {
				return strings.TrimSpace(fwd[:i])
			}
			return strings.TrimSpace(fwd)
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.get(clientIP(r, rl.behindProxy)).Allow() {
			w.Header().Set("Retry-After", "5")
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
