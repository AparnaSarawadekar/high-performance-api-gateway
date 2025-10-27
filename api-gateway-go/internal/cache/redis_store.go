package cache

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	enabled  bool
	ttl      time.Duration
	maxBytes int
	client   *redis.Client
}

type cacheEntry struct {
	Status int
	Header http.Header
	Body   []byte
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func atoiDefault(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

func NewRedisStoreFromEnv() *RedisStore {
	en := os.Getenv("CACHE_ENABLED") == "true"
	ttlSec := atoiDefault(getenv("CACHE_TTL_SECONDS", "30"), 30)
	maxB := atoiDefault(getenv("CACHE_MAX_BODY_BYTES", "1048576"), 1<<20)

	c := redis.NewClient(&redis.Options{
		Addr:     getenv("REDIS_ADDR", "redis:6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       atoiDefault(getenv("REDIS_DB", "0"), 0),
	})
	_ = c.Ping(context.Background()).Err()

	return &RedisStore{
		enabled:  en,
		ttl:      time.Duration(ttlSec) * time.Second,
		maxBytes: maxB,
		client:   c,
	}
}

func (s *RedisStore) Key(method, path, rawQuery string) string {
	h := sha1.Sum([]byte(method + " " + path + "?" + rawQuery))
	return "gw:v1:" + hex.EncodeToString(h[:])
}

func (s *RedisStore) Get(key string) (int, http.Header, []byte, bool) {
	if !s.enabled {
		return 0, nil, nil, false
	}
	b, err := s.client.Get(context.Background(), key).Bytes()
	if err != nil {
		return 0, nil, nil, false
	}
	var e cacheEntry
	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&e); err != nil {
		return 0, nil, nil, false
	}
	return e.Status, e.Header, e.Body, true
}

func (s *RedisStore) Set(key string, status int, hdr http.Header, body []byte) {
	if !s.enabled || len(body) > s.maxBytes {
		return
	}
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(cacheEntry{Status: status, Header: hdr, Body: body})
	_ = s.client.Set(context.Background(), key, buf.Bytes(), s.ttl).Err()
}

func (s *RedisStore) Enabled() bool         { return s.enabled }
func (s *RedisStore) TTL() time.Duration    { return s.ttl }
func (s *RedisStore) MaxBytes() int         { return s.maxBytes }
