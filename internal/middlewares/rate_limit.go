package middlewares

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter   *rate.Limiter
	lastSeen  time.Time
	failCount int
}

var (
	visitors   = make(map[string]*visitor)
	blockedIPs = make(map[string]time.Time)
	mu         sync.Mutex
)

// getRealIP returns the client's real IP address, considering reverse proxies.
func getRealIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	return r.RemoteAddr
}

// RateLimitMiddleware enforces rate limiting and temporarily blocks abusive IPs.
func RateLimitMiddleware(rps float64, burst int, cleanupInterval time.Duration) func(http.Handler) http.Handler {
	// Background cleanup routine
	go func() {
		for {
			time.Sleep(cleanupInterval)
			mu.Lock()
			now := time.Now()

			for ip, v := range visitors {
				if now.Sub(v.lastSeen) > 5*time.Minute {
					delete(visitors, ip)
				}
			}

			for ip, unblockAt := range blockedIPs {
				if now.After(unblockAt) {
					delete(blockedIPs, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getRealIP(r)

			mu.Lock()
			// Check if the IP is blocked
			if unblockTime, blocked := blockedIPs[ip]; blocked {
				if time.Now().Before(unblockTime) {
					mu.Unlock()
					http.Error(w, "IP temporarily blocked due to excessive requests", http.StatusTooManyRequests)
					log.Printf("❌ BLOCKED IP %s tried to access %s", ip, r.URL.Path)
					return
				}
				delete(blockedIPs, ip) // Unblock if time has passed
			}

			v, exists := visitors[ip]
			if !exists {
				limiter := rate.NewLimiter(rate.Limit(rps), burst)
				v = &visitor{limiter: limiter, lastSeen: time.Now()}
				visitors[ip] = v
			}
			v.lastSeen = time.Now()

			if !v.limiter.Allow() {
				v.failCount++
				if v.failCount > 20 {
					blockedIPs[ip] = time.Now().Add(15 * time.Minute)
					delete(visitors, ip)
					log.Printf("🚫 IP %s blocked for 15 minutes after %d rate limit violations", ip, v.failCount)
				} else {
					log.Printf("⚠️ Rate limit hit: IP=%s, Count=%d, Path=%s", ip, v.failCount, r.URL.Path)
				}
				mu.Unlock()
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			v.failCount = 0 // Reset on successful access
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
