import json, sys, datetime, pathlib

src = pathlib.Path("tests/load/baseline_summary.json")
dst = pathlib.Path("docs/Perf_Baseline.md")

if not src.exists():
    print(f"ERROR: {src} not found. Run k6 first.", file=sys.stderr)
    sys.exit(1)

data = json.loads(src.read_text())

m = data.get("metrics", {})
def get(mkey, vkey, default=None):
    return m.get(mkey, {}).get("values", {}).get(vkey, default)

# Core metrics
rps = get("http_reqs", "rate", 0)  # requests/sec
p50 = get("http_req_duration", "med", 0)
p90 = get("http_req_duration", "p(90)", 0)
p95 = get("http_req_duration", "p(95)", 0)
p99 = get("http_req_duration", "p(99)", 0) or 0
errors = get("http_req_failed", "rate", 0) * 100  # %

count = get("http_reqs", "count", 0)
sent_kb = get("data_sent", "sum", 0) / 1024 if m.get("data_sent") else 0
recv_kb = get("data_received", "sum", 0) / 1024 if m.get("data_received") else 0

now = datetime.datetime.now().strftime("%Y-%m-%d %H:%M")
md = f"""# Baseline Performance (k6)

**Run:** {now}  
**Script:** `tests/load/k6_baseline.js`  
**Command:** `k6 run --summary-export=tests/load/baseline_summary.json tests/load/k6_baseline.js`  

## Summary

| Metric | Value |
|--|--:|
| Requests (total) | {int(count):,} |
| Requests per second (RPS) | {rps:.1f} |
| Error rate (%) | {errors:.3f} |
| p50 latency (ms) | {p50:.1f} |
| p90 latency (ms) | {p90:.1f} |
| p95 latency (ms) | {p95:.1f} |
| p99 latency (ms) | {p99:.1f} |
| Data sent (KB) | {sent_kb:.1f} |
| Data received (KB) | {recv_kb:.1f} |

## Targets vs. Actuals

- **Throughput target:** 150 RPS → *actual* **{rps:.1f} RPS**
- **Latency target:** p95 ≤ 200 ms → *actual* **{p95:.1f} ms**
- **Error rate target:** < 0.1% → *actual* **{errors:.3f}%**

> Baseline scenario: warm-up 30s @ 50 RPS, then 2m @ 150 RPS to `/infer/python`.
"""

dst.write_text(md)
print(f"Wrote {dst}")
