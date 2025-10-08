[![CI](https://github.com/AparnaSarawadekar/high-performance-api-gateway/actions/workflows/ci.yml/badge.svg)](https://github.com/AparnaSarawadekar/high-performance-api-gateway/actions/workflows/ci.yml)

# 🚀 High-Performance API Gateway for AI Workloads
*Go | Python | Node.js | Docker | Kubernetes | Azure*

> A distributed API gateway prototype optimized for **AI inference traffic** — demonstrating scalable routing, caching, throttling, and observability patterns.

---

## 🎯 Performance Targets

| Metric | Baseline Goal | Optimized Goal | Description |
|:--|:--:|:--:|:--|
| **Throughput** | 150 RPS | **200 RPS (+33%)** | Load-balanced routing + caching |
| **Latency (p95)** | ≤ 200 ms | **≤ 140 ms** | Lower tail latency under burst load |
| **Error Rate (5xx)** | < 0.1 % | — | Consistent uptime under stress |
| **Cache Hit Ratio** | — | 60–80 % | For repeatable idempotent GETs |

---

## 🧱 Architecture Overview

| Component | Language | Purpose |
|:--|:--|:--|
| **api-gateway-go** | Go 1.25 | Entry point; routes & proxies traffic |
| **service-python** | FastAPI 3.12 | Simulated AI inference backend |
| **service-node** | Express 20.x | Fallback inference/test backend |

---

## 🐳 Local Development with Docker Compose

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

## 🧩 Gateway MVP — Step 7 (Routing + Health Checks)

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

## ⚙️ Continuous Integration — Step 8

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

## 📂 Repository Structure
```
.
├── api-gateway-go/        # Go gateway
│   ├── main.go
│   ├── Dockerfile
│   └── main_test.go
├── service-python/        # FastAPI service
│   ├── app.py
│   ├── Dockerfile
│   └── requirements.txt
├── service-node/          # Express service
│   ├── index.js
│   ├── Dockerfile
│   └── package.json
├── tests/
│   └── load/              # k6 scripts (baseline & perf)
├── docs/                  # Setup, containerization, metrics
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

Next → **Step 9 : Baseline Load Test (k6)**

---

## 👩‍💻 Author
**Aparna Vivek Sarawadekar**  
*M.S. Computer Science @ UCR • Software Engineer (Cloud & AI Infrastructure)*  
[LinkedIn](https://linkedin.com/in/aparna-sarawadekar) | [GitHub](https://github.com/AparnaSarawadekar)
