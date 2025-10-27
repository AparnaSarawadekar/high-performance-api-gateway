package cache

import "sync/atomic"

var hitCount atomic.Uint64
var missCount atomic.Uint64

func IncHit()  { hitCount.Add(1) }
func IncMiss() { missCount.Add(1) }

type Metrics struct {
	Hits     uint64  `json:"hits"`
	Misses   uint64  `json:"misses"`
	HitRatio float64 `json:"hit_ratio"`
	Backend  string  `json:"backend"`
	TTL      string  `json:"ttl"`
}

func Snapshot(backend, ttl string) Metrics {
	h := hitCount.Load()
	m := missCount.Load()
	r := 0.0
	if h+m > 0 {
		r = float64(h) / float64(h+m)
	}
	return Metrics{Hits: h, Misses: m, HitRatio: r, Backend: backend, TTL: ttl}
}
