# Baseline Performance (k6)

**Run:** 2025-10-09 20:06  
**Script:** `tests/load/k6_baseline.js`  
**Command:** `k6 run --summary-export=tests/load/baseline_summary.json tests/load/k6_baseline.js`  

## Summary

| Metric | Value |
|--|--:|
| Requests (total) | 0 |
| Requests per second (RPS) | 0.0 |
| Error rate (%) | 0.000 |
| p50 latency (ms) | 0.0 |
| p90 latency (ms) | 0.0 |
| p95 latency (ms) | 0.0 |
| p99 latency (ms) | 0.0 |
| Data sent (KB) | 0.0 |
| Data received (KB) | 0.0 |

## Targets vs. Actuals

- **Throughput target:** 150 RPS → *actual* **0.0 RPS**
- **Latency target:** p95 ≤ 200 ms → *actual* **0.0 ms**
- **Error rate target:** < 0.1% → *actual* **0.000%**

> Baseline scenario: warm-up 30s @ 50 RPS, then 2m @ 150 RPS to `/infer/python`.
