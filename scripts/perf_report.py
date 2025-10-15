#!/usr/bin/env python3
import json, pathlib, datetime

repo_root = pathlib.Path(".")
baseline_path = repo_root / "tests" / "load" / "perf_baseline_summary.json"
cached_path   = repo_root / "tests" / "load" / "perf_cached_summary.json"

output_report_path = repo_root / "PerfReport_CacheV1.md"
output_json_path   = repo_root / "tests" / "load" / "perf_cache_v1_results.json"
output_csv_path    = repo_root / "tests" / "load" / "perf_cache_v1_results.csv"

def get(d, dotted, default=None):
    cur = d
    for k in dotted.split("."):
        if isinstance(cur, dict) and k in cur:
            cur = cur[k]
        else:
            return default
    return cur

def extract_metrics(summary):
    total = get(summary, "metrics.http_reqs.values.count", 0) or 0
    p50 = (get(summary, "metrics.http_req_duration.values['p(50)']") or
           get(summary, "metrics.http_req_duration.percentiles['50']"))
    p95 = (get(summary, "metrics.http_req_duration.values['p(95)']") or
           get(summary, "metrics.http_req_duration.percentiles['95']"))
    p99 = (get(summary, "metrics.http_req_duration.values['p(99)']") or
           get(summary, "metrics.http_req_duration.percentiles['99']"))
    fail_rate = get(summary, "metrics.http_req_failed.values.rate", 0.0) or 0.0

    dur_ms = get(summary, "state.testRunDurationMs")
    if not dur_ms:
        secs = get(summary, "state.testRunDuration")
        dur_ms = secs * 1000.0 if secs else None

    rps = total / (dur_ms / 1000.0) if (total and dur_ms) else get(summary, "metrics.http_reqs.values.rate")
    return {
        "rps": rps,
        "p50_ms": p50,
        "p95_ms": p95,
        "p99_ms": p99,
        "error_rate_pct": fail_rate * 100.0,
        "total_requests": total,
    }

def pct_change(new, old):
    if new is None or old in (None, 0): return None
    return (new - old) / old * 100.0

def fmt_num(x):
    if x is None: return "n/a"
    return f"{x:.2f}" if isinstance(x, float) else str(x)

def fmt_ms(x):  return "n/a" if x is None else f"{x:.1f} ms"
def fmt_pct(x): return "n/a" if x is None else f"{x:.1f}%"

baseline = json.loads(baseline_path.read_text())
cached   = json.loads(cached_path.read_text())

b = extract_metrics(baseline)
c = extract_metrics(cached)

delta = {
    "rps": pct_change(c["rps"], b["rps"]),
    "p95_ms": pct_change(c["p95_ms"], b["p95_ms"]),
    "p99_ms": pct_change(c["p99_ms"], b["p99_ms"]),
    "error_rate_pct": pct_change(c["error_rate_pct"], b["error_rate_pct"]),
}

output_json_path.parent.mkdir(parents=True, exist_ok=True)
output_json_path.write_text(json.dumps({
    "generated_at": datetime.datetime.utcnow().isoformat() + "Z",
    "baseline": b,
    "cached": c,
    "delta_pct": delta
}, indent=2))

output_csv_path.write_text(
    "metric,baseline,cached,delta_pct\n"
    f"rps,{fmt_num(b['rps'])},{fmt_num(c['rps'])},{fmt_pct(delta['rps'])}\n"
    f"p95_ms,{fmt_num(b['p95_ms'])},{fmt_num(c['p95_ms'])},{fmt_pct(delta['p95_ms'])}\n"
    f"p99_ms,{fmt_num(b['p99_ms'])},{fmt_num(c['p99_ms'])},{fmt_pct(delta['p99_ms'])}\n"
    f"error_rate_pct,{fmt_num(b['error_rate_pct'])},{fmt_num(c['error_rate_pct'])},{fmt_pct(delta['error_rate_pct'])}\n"
)

md = f"""# Performance Report — CacheV1

**Scope:** Baseline vs in-memory cached run (same VUs and duration)

| Metric | Baseline | Cached | Δ % |
|---|---:|---:|---:|
| Throughput (RPS) | {fmt_num(b['rps'])} | {fmt_num(c['rps'])} | {fmt_pct(delta['rps'])} |
| p50 Latency | {fmt_ms(b['p50_ms'])} | {fmt_ms(c['p50_ms'])} | — |
| p95 Latency | {fmt_ms(b['p95_ms'])} | {fmt_ms(c['p95_ms'])} | {fmt_pct(delta['p95_ms'])} |
| p99 Latency | {fmt_ms(b['p99_ms'])} | {fmt_ms(c['p99_ms'])} | {fmt_pct(delta['p99_ms'])} |
| Error Rate | {fmt_num(b['error_rate_pct'])}% | {fmt_num(c['error_rate_pct'])}% | {fmt_pct(delta['error_rate_pct'])} |

**Notes**
- Profile: 40 VUs, 20 s, GET /slow, `X-Forwarded-For` per VU
- Cache warm-up: two GETs before timed run
- Same environment and Docker versions as baseline

**Artifacts**
- `tests/load/perf_baseline.json`
- `tests/load/perf_cached.json`
- `tests/load/perf_cache_v1_results.json`
- `tests/load/perf_cache_v1_results.csv`
"""
output_report_path.write_text(md)
print(f"Wrote {output_report_path}\n - {output_json_path}\n - {output_csv_path}")
