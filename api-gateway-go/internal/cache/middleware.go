package cache

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
)

// responseCapture buffers the downstream response so we can decide
// whether to cache it, then write it out to the real ResponseWriter.
type responseCapture struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
	header http.Header
	wroteH bool
}

func (rc *responseCapture) Header() http.Header {
	if rc.header == nil {
		rc.header = make(http.Header)
	}
	return rc.header
}

func (rc *responseCapture) WriteHeader(status int) {
	if rc.wroteH {
		return
	}
	rc.status = status
	rc.wroteH = true
}

func (rc *responseCapture) Write(p []byte) (int, error) {
	// Buffer only; we'll write to the real writer in FlushTo().
	return rc.buf.Write(p)
}

func (rc *responseCapture) FlushTo(w http.ResponseWriter, method string) {
	// Copy headers to the real writer
	for k, v := range rc.header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	if rc.status == 0 {
		rc.status = http.StatusOK
	}
	w.WriteHeader(rc.status)

	// HEAD responses must not include a body.
	if method != http.MethodHead {
		_, _ = w.Write(rc.buf.Bytes())
	}
}

// Middleware provides transparent HTTP response caching for idempotent GET/HEAD.
type Middleware struct {
	store  StoreAPI
	bypass map[string]struct{} // exact paths to skip (health, metrics, etc.)
}

// NewMiddleware constructs a caching middleware with optional bypass paths.
// Any request whose URL.Path has an exact match in bypass is not cached.
func NewMiddleware(store StoreAPI, bypassPaths ...string) *Middleware {
	b := make(map[string]struct{}, len(bypassPaths))
	for _, p := range bypassPaths {
		b[p] = struct{}{}
	}
	return &Middleware{store: store, bypass: b}
}

func isCacheableRequest(r *http.Request) bool {
	// Only cache idempotent requests
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	// Authorization present? Be conservative.
	if r.Header.Get("Authorization") != "" {
		return false
	}
	// Client forced no-store?
	cc := strings.ToLower(r.Header.Get("Cache-Control"))
	if strings.Contains(cc, "no-store") {
		return false
	}
	return true
}

func isCacheableResponse(status int, hdr http.Header) bool {
	// Cache successful responses only
	if status < 200 || status >= 300 {
		return false
	}
	// Respect origin no-store
	cc := strings.ToLower(hdr.Get("Cache-Control"))
	if strings.Contains(cc, "no-store") {
		return false
	}
	return true
}

// Handler wraps the next handler with caching.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	// If caching disabled or no store, pass-through.
	if m == nil || m.store == nil || !m.store.Enabled() {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass specific paths (health, metrics, etc.)
		if _, skip := m.bypass[r.URL.Path]; skip {
			next.ServeHTTP(w, r)
			return
		}
		// Non-cacheable requests pass-through
		if !isCacheableRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Try cache
		key := m.store.Key(r.Method, r.URL.Path, r.URL.RawQuery)
		if status, hdr, body, ok := m.store.Get(key); ok {
			// Serve cache hit
			for k, v := range hdr {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.Header().Set("X-Cache", "HIT")
			IncHit()

			w.WriteHeader(status)
			if r.Method != http.MethodHead {
				_, _ = w.Write(body)
			}
			return
		}

		// MISS: capture the downstream response
		rc := &responseCapture{ResponseWriter: w}
		next.ServeHTTP(rc, r)

		// Decide to store
		if isCacheableResponse(rc.status, rc.header) && rc.buf.Len() <= m.store.MaxBytes() {
			// Optional freshness hint header for clients
			rc.Header().Set("Cache-Control", "max-age="+strconv.Itoa(int(m.store.TTL().Seconds())))
			m.store.Set(key, rc.status, rc.header, rc.buf.Bytes())
		}

		// Mark MISS and forward the response
		w.Header().Set("X-Cache", "MISS")
		IncMiss()
		rc.FlushTo(w, r.Method)
	})
}
