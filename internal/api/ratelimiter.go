package api

import (
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nongpal/Palge-Backend/internal/apperr"
)

type ClientInfo struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

type RateLimiter struct {
	clientMap sync.Map
	rate      float64
	capacity  float64
	stop      chan struct{}
	stopOnce  sync.Once
}

func NewRateLimiter(rate float64, capacity float64) *RateLimiter {
	return &RateLimiter{
		rate:     rate,
		capacity: capacity,
		stop:     make(chan struct{}),
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	now := time.Now()

	actual, _ := rl.clientMap.LoadOrStore(clientID, &ClientInfo{
		tokens:     rl.capacity,
		lastRefill: now,
	})

	client := actual.(*ClientInfo)

	client.mu.Lock()
	defer client.mu.Unlock()

	elapsed := now.Sub(client.lastRefill).Seconds()
	client.tokens += elapsed * rl.rate

	if client.tokens > rl.capacity {
		client.tokens = rl.capacity
	}

	client.lastRefill = now
	client.lastSeen = now

	if client.tokens >= 1 {
		client.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) StartJanitor(interval, expiry time.Duration) {

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				rl.clientMap.Range(func(key, value any) bool {
					client := value.(*ClientInfo)

					client.mu.Lock()
					defer client.mu.Unlock()

					if time.Since(client.lastSeen) > expiry {
						rl.clientMap.Delete(key)
					}
					return true
				})
			case <-rl.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (rl *RateLimiter) StopJanitor() {
	rl.stopOnce.Do(func() {
		close(rl.stop)
	})
}

func (app *Application) middlewareRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.cfg.Limiter.Enabled {
			ip := app.getRealIP(r)
			if !app.rateLimiter.Allow(ip) {
				app.writeError(w, r, apperr.New(apperr.CodeRateLimited))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Application) getRealIP(r *http.Request) string {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		remoteIP = r.RemoteAddr
	}

	if !app.isTrustedProxies(remoteIP) {
		return remoteIP
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for _, part := range slices.Backward(parts) {
			ip := strings.TrimSpace(part)
			if parsed, ok := netip.ParseAddr(ip); ok == nil && !app.isTrustedProxies(parsed.String()) {
				return parsed.String()
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	return remoteIP
}

func (app *Application) isTrustedProxies(ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}

	for _, cidr := range app.cfg.Limiter.TrustedProxies {
		prefix, err := netip.ParsePrefix(cidr)
		if err == nil && prefix.Contains(addr) {
			return true
		}
	}

	return false
}
