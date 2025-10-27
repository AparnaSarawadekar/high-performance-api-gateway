package cache

import (
	"net/http"
	"time"
)

// Wrap the in-memory *Store so it satisfies StoreAPI
type memoryAdapter struct{ *Store }

func (m memoryAdapter) Key(method, path, rawQuery string) string {
	return m.Store.Key(method, path, rawQuery)
}

func (m memoryAdapter) Get(key string) (int, http.Header, []byte, bool) {
	status, hdrMap, body, ok := m.Store.Get(key)
	if !ok {
		return 0, nil, nil, false
	}
	return status, http.Header(hdrMap), body, true
}

func (m memoryAdapter) Set(key string, status int, hdr http.Header, body []byte) {
	m.Store.Set(key, status, map[string][]string(hdr), body)
}

func (m memoryAdapter) Enabled() bool      { return m.Store.Enabled() }
func (m memoryAdapter) TTL() time.Duration { return m.Store.TTL() }
func (m memoryAdapter) MaxBytes() int      { return m.Store.MaxBytes() }

// Select backend via env; default to memory
func NewFromEnv() StoreAPI {
	switch getenv("CACHE_BACKEND", "memory") {
	case "redis":
		return NewRedisStoreFromEnv()
	default:
		return memoryAdapter{NewMemoryStoreFromEnv()}
	}
}
