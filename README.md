[![CI](https://github.com/AparnaSarawadekar/high-performance-api-gateway/actions/workflows/ci.yml/badge.svg)](https://github.com/AparnaSarawadekar/high-performance-api-gateway/actions/workflows/ci.yml)

# High-Performance API Gateway for AI Workloads
*Go | Python | Node.js | Docker | Kubernetes | Azure*

> A distributed API gateway prototype optimized for **AI inference traffic** — demonstrating scalable routing, caching, throttling, and observability patterns.

---

## Performance Targets

| Metric | Baseline Goal | Optimized Goal | Description |
|:--|:--:|:--:|:--|
| **Throughput** | 150 RPS | **200 RPS (+33%)** | Load-balanced routing + caching |
| **Latency (p95)** | ≤ 200 ms | **≤ 140 ms** | Lower tail latency under burst load |
| **Error Rate (5xx)** | < 0.1 % | — | Consistent uptime under stress |
| **Cache Hit Ratio** | — | 60–80 % | For repeatable idempotent GETs |

---

## Architecture Overview

| Component | Language | Purpose |
|:--|:--|:--|
| **api-gateway-go** | Go 1.25 | Entry point; routes & proxies traffic |
| **service-python** | FastAPI 3.12 | Simulated AI inference backend |
| **service-node** | Express 20.x | Fallback inference/test backend |

---

## Local Development with Docker Compose

### Prerequisites
- Docker Desktop (or Docker Engine + Compose v2)
- Make (optional but recommended)

### Quick Start
```bash
# From repository root
docker compose up -d --build
docker compose ps
```

Check that all containers are **healthy**:
```bash
curl -s http://localhost:8080/healthz | jq .
```

Expected:
```json
{ "ok": true, "service": "api-gateway-go", "uptime_ms": 12345 }
```

---

## Gateway MVP — Step 7 (Routing + Health Checks)

The **Go API Gateway** exposes health info and proxies inference requests to both backend services.

### Routes

| Route | Method | Description | Target |
|:--|:--:|:--|:--|
| `/healthz` | GET | Returns gateway uptime and backend status | — |
| `/infer/python` | POST | Proxies JSON to `service-python /infer` | FastAPI backend |
| `/infer/node` | POST | Proxies JSON to `service-node /infer` | Node backend |

### Environment Variables

| Variable | Description | Default |
|:--|:--|:--|
| `PY_SERVICE_URL` | URL of Python backend | `http://service-python:8001` |
| `NODE_SERVICE_URL` | URL of Node backend | `http://service-node:8002` |
| `PORT` | Gateway listening port | `8080` |

### Run & Test
```bash
docker compose up -d --build
curl -fsS http://localhost:8080/healthz
curl -fsS -X POST http://localhost:8080/infer/python \
  -H "Content-Type: application/json" -d '{"prompt":"hello"}'
curl -fsS -X POST http://localhost:8080/infer/node \
  -H "Content-Type: application/json" -d '{"prompt":"hello"}'
docker compose down -v
```

---

## Continuous Integration — Step 8

Automated testing & builds run via  
[`.github/workflows/ci.yml`](.github/workflows/ci.yml).

### Pipeline Stages

| Job | Language | Key Actions |
|:--|:--|:--|
| **Go (Gateway)** | Go 1.25 | `go vet`, `go build`, `go test` |
| **Python (Service)** | Python 3.12 | `pip install`, `ruff check`, `pytest` |
| **Node (Service)** | Node 20 | `npm ci`, `eslint .` |
| **Docker Builds** | Docker 25 + Buildx | Build images for all services |
| **Compose Smoke Test** | Docker Compose V2 | Bring up stack, check `/healthz`, POST to `/infer/python` and `/infer/node`, then teardown |

### Run the Same Checks Locally
```bash
# From repo root

# --- Lint & Unit Tests ---
( cd api-gateway-go && go vet ./... && go test -v ./... )
( cd service-python && ruff check . )      # pytest optional
( cd service-node && npm ci && npm run lint )

# --- Docker Compose Smoke ---
docker compose up -d --build
curl -fsS http://localhost:8080/healthz
curl -fsS -X POST http://localhost:8080/infer/python \
  -H "Content-Type: application/json" -d '{"prompt":"hello"}'
curl -fsS -X POST http://localhost:8080/infer/node \
  -H "Content-Type: application/json" -d '{"prompt":"hello"}'
docker compose down -v
```

### ✅ CI Status Badge
[![CI](https://github.com/AparnaSarawadekar/high-performance-api-gateway/actions/workflows/ci.yml/badge.svg)](https://github.com/AparnaSarawadekar/high-performance-api-gateway/actions/workflows/ci.yml)

---

## Repository Structure
```
.
├── api-gateway-go/        # Go gateway
│   ├── main.go
│   ├── Dockerfile
│   ├── main_test.go
│   └── go.mod
├── service-python/        # FastAPI service
│   ├── app.py
│   ├── Dockerfile
│   └── requirements.txt
├── service-node/          # Express service
│   ├── index.js
│   ├── Dockerfile
│   ├── .eslintrc.json
│   ├── package-lock.json
│   └── package.json
├── tests/
│   └── load/              # k6 scripts (baseline & perf)
├── docs/                  # Setup, containerization, metrics
│   ├── Containerization.md
│   ├── Local-Tooling-Status.md
│   ├── Perf_Baseline.md
│   └── Scope-and-Metrics.md
├── docker-compose.yml
├── Makefile
└── .github/workflows/ci.yml
```

---

## 🧾 Progress Summary (Completed Steps 1 – 8)

| Step | Description | Status |
|:--:|:--|:--:|
| 1 | Define scope & success metrics | ✅ |
| 2 | Set up local tooling (Docker, Kind, k6) | ✅ |
| 3 | Create GitHub repo + project board | ✅ |
| 4 | Scaffold services (Go + Python + Node) | ✅ |
| 5 | Add Dockerfiles for all services | ✅ |
| 6 | Wire up docker-compose (local orchestration) | ✅ |
| 7 | Implement gateway MVP (routing + health) | ✅ |
| 8 | Add GitHub Actions CI (pipeline) | ✅ |
| 9 | Baseline Load Test (k6) | ✅ |

---

## ⚡ Step 9 — Baseline Load Test with k6

This step benchmarks the **current performance** of the gateway stack before adding caching, throttling, or Redis.  
We’ll measure **RPS, p95/p99 latency, and error rates** to establish a baseline.

---

### 🧪 Test Script

File: `tests/load/k6_baseline.js`

```js
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const PAYLOAD = JSON.stringify({ prompt: 'hello world' });
const HEADERS = { 'Content-Type': 'application/json' };

export const options = {
  scenarios: {
    warmup: {
      executor: 'constant-arrival-rate',
      rate: 50,               
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 50,
      maxVUs: 100,
      startTime: '0s',
    },
    baseline: {
      executor: 'constant-arrival-rate',
      rate: 150,              
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 200,
      maxVUs: 300,
      startTime: '30s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.001'],      
    http_req_duration: ['p(95)<200', 'p(99)<400'],
  },
  tags: { test_stage: 'baseline' },
};

export default function () {
  const res = http.post(`${BASE}/infer/python`, PAYLOAD, { headers: HEADERS });
  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(0.05);
}
```

---

### ▶️ Run the Baseline Test

From the repo root:

```bash
# 1️⃣ Ensure stack is up
docker compose up -d --build
curl -s http://localhost:8080/healthz

# 2️⃣ Run k6 test
k6 run --summary-export=tests/load/baseline_summary.json tests/load/k6_baseline.js
```

You’ll see live results in the terminal, and a JSON summary will be saved to  
`tests/load/baseline_summary.json`.

---

### 📊 Generate Markdown Summary

File: `tools/k6_summary_to_md.py`

```python
import json, datetime, pathlib

src = pathlib.Path("tests/load/baseline_summary.json")
dst = pathlib.Path("docs/Perf_Baseline.md")

data = json.loads(src.read_text())
m = data.get("metrics", {})
def get(mkey, vkey): return m.get(mkey, {}).get("values", {}).get(vkey, 0)

rps = get("http_reqs","rate")
p95 = get("http_req_duration","p(95)")
errors = get("http_req_failed","rate") * 100
count = get("http_reqs","count")

md = f"""# Baseline Performance (k6)

**Date:** {datetime.datetime.now():%Y-%m-%d %H:%M}  
**Script:** `tests/load/k6_baseline.js`

| Metric | Value |
|--|--:|
| Requests | {int(count):,} |
| Throughput (RPS) | {rps:.1f} |
| p95 Latency (ms) | {p95:.1f} |
| Error Rate (%) | {errors:.3f}% |

Targets:
- Throughput ≥150 RPS
- p95 ≤ 200 ms
- Error < 0.1 %

> This report was generated automatically from the k6 JSON summary.
"""

dst.write_text(md)
print(f"Wrote {dst}")
```

Run it:

```bash
python3 tools/k6_summary_to_md.py
```

A new file `docs/Perf_Baseline.md` will appear containing a Markdown report.

---

### Commit Artifacts

```bash
git add tests/load/k6_baseline.js tests/load/baseline_summary.json tools/k6_summary_to_md.py docs/Perf_Baseline.md
git commit -m "step9: add k6 baseline test + summary report"
git push origin main
```

---

### Expected Outcome

After Step 9, you’ll have:
- A reproducible **load test** (`k6_baseline.js`)
- A **JSON metrics file**
- A readable **Markdown performance report**
- Documented baseline targets before optimization

### Step 9 Baseline Snapshot

| Metric | Value |
|:--|--:|
| RPS | ~1950 |
| p95 (ms) | ~1.92 |
| p99 (ms) | ~2.60 |
| Error Rate | 0.00% |

_Source: [`tests/load/baseline_console.txt`](tests/load/baseline_console.txt). Tag: `perf-baseline-v0`._

Next → **Step 10 — Add rate limiting & throttling in gateway (token bucket)**

---

## 👩‍💻 Author
**Aparna Vivek Sarawadekar**  
*M.S. Computer Science @ UCR • Software Engineer (Cloud & AI Infrastructure)*  
[LinkedIn](https://linkedin.com/in/aparna-sarawadekar) | [GitHub](https://github.com/AparnaSarawadekar)
