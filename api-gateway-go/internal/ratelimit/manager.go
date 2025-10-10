package ratelimit

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	enabled bool
	global  *bucket

	clients      sync.Map // key (IP) -> *clientEntry
	cleanupAfter time.Duration
	clientRPS    float64
	clientBurst  int
}

type clientEntry struct {
	b   *bucket
	ttL time.Time
}

func getenvFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
func getenvBool(key string, def bool) bool {
	v := strings.ToLower(os.Getenv(key))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return def
}

func NewManagerFromEnv() *Manager {
	enabled := getenvBool("RATE_LIMIT_ENABLED", true)

	globalRPS := getenvFloat("GLOBAL_RPS", 200)
	globalBurst := getenvInt("GLOBAL_BURST", 100)

	clientRPS := getenvFloat("CLIENT_RPS", 20)
	clientBurst := getenvInt("CLIENT_BURST", 40)

	cleanupMin := getenvInt("RL_CLEANUP_MINUTES", 10)

	m := &Manager{
		enabled:      enabled,
		global:       newBucket(globalBurst, globalRPS),
		clientRPS:    clientRPS,
		clientBurst:  clientBurst,
		cleanupAfter: time.Duration(cleanupMin) * time.Minute,
	}

	// Periodic cleanup of idle client buckets
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for range t.C {
			now := time.Now()
			m.clients.Range(func(k, v any) bool {
				ce := v.(*clientEntry)
				if now.Sub(ce.ttL) > m.cleanupAfter {
					m.clients.Delete(k)
				}
				return true
			})
		}
	}()

	return m
}

func (m *Manager) clientKey(r *http.Request) string {
	// Respect X-Forwarded-For (left-most) when present
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	// Fallback to remote address
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func (m *Manager) getClientBucket(key string) *bucket {
	now := time.Now()
	if v, ok := m.clients.Load(key); ok {
		ce := v.(*clientEntry)
		ce.ttL = now
		return ce.b
	}
	nb := newBucket(m.clientBurst, m.clientRPS)
	m.clients.Store(key, &clientEntry{b: nb, ttL: now})
	return nb
}

// Middleware applies global + per-client buckets.
// Skips /healthz so orchestrators don't flap.
func (m *Manager) Middleware(next http.Handler) http.Handler {
	if !m.enabled {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		if ok, retry, remaining := m.global.allow(); !ok {
			writeLimited(w, retry, "global", remaining)
			return
		}
		key := m.clientKey(r)
		cb := m.getClientBucket(key)
		if ok, retry, remaining := cb.allow(); !ok {
			writeLimited(w, retry, "client", remaining)
			return
		}

		w.Header().Set("RateLimit-Policy", "global, client; unit=second")
		next.ServeHTTP(w, r)
	})
}

func writeLimited(w http.ResponseWriter, retry time.Duration, scope string, remaining int) {
	if retry > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
	}
	w.Header().Set("RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("RateLimit-Scope", scope)
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(`{"error":"too_many_requests"}`))
}