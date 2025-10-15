# Performance Report — CacheV1

**Scope:** Baseline vs in-memory cached run (same VUs and duration)

| Metric | Baseline | Cached | Δ % |
|---|---:|---:|---:|
| Throughput (RPS) | 708.91 | 708.98 | 0.0% |
| p50 Latency | n/a | n/a | — |
| p95 Latency | n/a | n/a | n/a |
| p99 Latency | n/a | n/a | n/a |
| Error Rate | 0.00% | 0.00% | n/a |

**Notes**
- Profile: 40 VUs, 20 s, GET /slow, `X-Forwarded-For` per VU
- Cache warm-up: two GETs before timed run
- Same environment and Docker versions as baseline

**Artifacts**
- `tests/load/perf_baseline.json`
- `tests/load/perf_cached.json`
- `tests/load/perf_cache_v1_results.json`
- `tests/load/perf_cache_v1_results.csv`
