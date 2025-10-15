package cache

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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
	// Buffer the body
	return rc.buf.Write(p)
}

func (rc *responseCapture) FlushTo(w http.ResponseWriter) {
	// Copy headers to real writer
	for k, v := range rc.header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	if rc.status == 0 {
		rc.status = http.StatusOK
	}
	w.WriteHeader(rc.status)
	_, _ = w.Write(rc.buf.Bytes())
}

type Middleware struct {
	store *Store
	// path prefixes to skip (health, metrics, etc.)
	bypass map[string]struct{}
}

func NewMiddleware(store *Store, bypassPaths ...string) *Middleware {
	b := make(map[string]struct{}, len(bypassPaths))
	for _, p := range bypassPaths {
		b[p] = struct{}{}
	}
	return &Middleware{store: store, bypass: b}
}

func isCacheableRequest(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	// Authorization on request? Be conservative; skip caching.
	if r.Header.Get("Authorization") != "" {
		return false
	}
	// Skip if client forces no-store
	cc := strings.ToLower(r.Header.Get("Cache-Control"))
	if strings.Contains(cc, "no-store") {
		return false
	}
	return true
}

func isCacheableResponse(status int, hdr http.Header) bool {
	if status != http.StatusOK {
		return false
	}
	// Respect origin-provided no-store
	cc := strings.ToLower(hdr.Get("Cache-Control"))
	if strings.Contains(cc, "no-store") {
		return false
	}
	return true
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	if m.store == nil || !m.store.enabled {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, skip := m.bypass[r.URL.Path]; skip {
			next.ServeHTTP(w, r)
			return
		}
		if !isCacheableRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		key := m.store.Key(r.Method, r.URL.Path, r.URL.RawQuery)
		if status, hdr, body, ok := m.store.Get(key); ok {
			// Serve hit
			for k, v := range hdr {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			// Update Age
			w.Header().Set("Age", strconv.Itoa(int(time.Since(time.Now().Add(-m.store.ttl)).Seconds())))
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(status)
			_, _ = w.Write(body)
			return
		}

		// MISS: capture downstream response
		rc := &responseCapture{ResponseWriter: w}
		next.ServeHTTP(rc, r)

		if isCacheableResponse(rc.status, rc.header) && rc.buf.Len() <= m.store.maxBytes {
			// store and then write through
			m.store.Set(key, rc.status, rc.header, rc.buf.Bytes())
		}

		// Mark MISS and forward
		w.Header().Set("X-Cache", "MISS")
		rc.FlushTo(w)
	})
}
