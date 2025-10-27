package cache

import (
	"net/http"
	"time"
)

type StoreAPI interface {
	Key(method, path, rawQuery string) string
	Get(key string) (status int, hdr http.Header, body []byte, ok bool)
	Set(key string, status int, hdr http.Header, body []byte)
	Enabled() bool
	TTL() time.Duration
	MaxBytes() int
}
