# high-performance-api-gateway
## Performance Targets
- Daily workload: 1M+ requests (simulated with load tests)
- Baseline goal: p95 ≤ 200ms @ 150 RPS
- Optimized goal: p95 ≤ 140ms, 33% throughput gain
- Error rate: <0.1% 5xx
- Cache hit ratio: 60–80%


## Local Dev with Docker Compose
Spin up the full stack (Go gateway + Python + Node) locally.
### Prereqs
- Docker Desktop (or Docker Engine + Compose v2)
### Quick start
```bash
# from repo root
docker compose up -d --build
docker compose ps

---

## Step 7 — Gateway MVP (Routing + Health Checks)

This step turns the Go-based API Gateway into a working entry point for the stack.  
It now exposes health information and proxies inference requests to both backend services.

### 🚀 Overview
| Component | Language | Purpose |
|------------|-----------|----------|
| **api-gateway-go** | Go | Entry point; routes requests and exposes `/healthz` |
| **service-python** | FastAPI (Python) | Simulated AI inference endpoint |
| **service-node** | Express (Node.js) | Alternate backend for inference |

### 🧩 Available Routes

| Route | Method | Description | Target |
|-------|---------|--------------|--------|
| `/healthz` | GET | Returns gateway uptime and backend targets | — |
| `/infer/python` | POST | Proxies JSON payloads to `service-python /infer` | FastAPI backend |
| `/infer/node` | POST | Proxies JSON payloads to `service-node /infer` | Node backend |

### ⚙️ Environment Variables

| Variable | Description | Default |
|-----------|--------------|----------|
| `PY_SERVICE_URL` | Internal URL of Python service | `http://service-python:8001` |
| `NODE_SERVICE_URL` | Internal URL of Node service | `http://service-node:8002` |
| `PORT` | Gateway listening port | `8080` |

### ▶️ Run & Test

Start the entire stack:
```bash
docker compose up -d --build
docker compose ps

