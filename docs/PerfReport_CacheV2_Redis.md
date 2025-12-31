# PerfReport — Cache V2 (Redis)

## Workload
- Script: `tests/load/cache_test.js`
- VUs: 40
- Duration: 20s
- Target: `GET /slow`
- Backend: Redis (`CACHE_BACKEND=redis`)

## Artifacts
- Console: `tests/load/perf_cache_v2_redis_console.txt`
- Summary: `tests/load/perf_cache_v2_redis_summary.json`

## Key Results (Redis) — verified from summary JSON
- RPS: ~701 req/s (http_reqs.rate = 700.69/s)
- p95 latency: ~11.74 ms (http_req_duration p(95) = 11.74 ms)
- Error rate: 0% (http_req_failed.rate = 0)

## Notes
- Same k6 workload as Cache V1 to keep results comparable.
- Redis backend enables shared cache across gateway replicas and survives gateway restarts.
