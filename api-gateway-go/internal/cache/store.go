package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type entry struct {
	status int
	header map[string][]string
	body   []byte
	exp    time.Time
	size   int
}

type Store struct {
	mu       sync.RWMutex
	items    map[string]*entry
	enabled  bool
	ttl      time.Duration
	maxN     int
	maxBytes int
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getenvBool(key string, def bool) bool {
	v := strings.ToLower(os.Getenv(key))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func NewFromEnv() *Store {
	return New(
		getenvBool("CACHE_ENABLED", true),
		time.Duration(getenvInt("CACHE_TTL_SECONDS", 30))*time.Second,
		getenvInt("CACHE_MAX_ENTRIES", 10000),
		getenvInt("CACHE_MAX_BODY_BYTES", 1<<20), // 1MiB
	)
}

func New(enabled bool, ttl time.Duration, maxN, maxBytes int) *Store {
	s := &Store{
		items:    make(map[string]*entry),
		enabled:  enabled,
		ttl:      ttl,
		maxN:     maxN,
		maxBytes: maxBytes,
	}
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for range t.C {
			s.gc()
		}
	}()
	return s
}

func (s *Store) Key(method, path, rawQuery string) string {
	// Only GET/HEAD use cache; caller enforces this.
	// Key = METHOD + " " + normalized URL (path + ?sortedQuery if needed)
	// Here we just use path + rawQuery; upstream can sort if needed.
	h := sha1.Sum([]byte(method + " " + path + "?" + rawQuery))
	return hex.EncodeToString(h[:])
}

func (s *Store) Get(key string) (status int, header map[string][]string, body []byte, ok bool) {
	if !s.enabled {
		return 0, nil, nil, false
	}
	s.mu.RLock()
	e, ok := s.items[key]
	if ok && time.Now().Before(e.exp) {
		// Shallow copy headers to avoid callers mutating
		hcopy := make(map[string][]string, len(e.header))
		for k, v := range e.header {
			cp := make([]string, len(v))
			copy(cp, v)
			hcopy[k] = cp
		}
		bodyCopy := make([]byte, len(e.body))
		copy(bodyCopy, e.body)
		s.mu.RUnlock()
		return e.status, hcopy, bodyCopy, true
	}
	s.mu.RUnlock()
	return 0, nil, nil, false
}

func (s *Store) Set(key string, status int, header map[string][]string, body []byte) {
	if !s.enabled {
		return
	}
	if s.maxBytes > 0 && len(body) > s.maxBytes {
		return
	}
	now := time.Now()
	e := &entry{
		status: status,
		header: make(map[string][]string, len(header)),
		body:   append([]byte(nil), body...),
		exp:    now.Add(s.ttl),
		size:   len(body),
	}
	// Copy headers except hop-by-hop/caching negatives
	for k, v := range header {
		kl := strings.ToLower(k)
		if kl == "connection" || kl == "transfer-encoding" || kl == "keep-alive" {
			continue
		}
		e.header[k] = append([]string(nil), v...)
	}
	// Write Age=0; middleware will overwrite on serve
	e.header["Age"] = []string{"0"}

	s.mu.Lock()
	if s.items == nil {
		s.items = make(map[string]*entry)
	}
	// crude soft cap: if over capacity, run GC inline
	if s.maxN > 0 && len(s.items) >= s.maxN {
		s.gc()
	}
	s.items[key] = e
	s.mu.Unlock()
}

func (s *Store) gc() {
	now := time.Now()
	s.mu.Lock()
	for k, v := range s.items {
		if now.After(v.exp) {
			delete(s.items, k)
		}
	}
	s.mu.Unlock()
}
