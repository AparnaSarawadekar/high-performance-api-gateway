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

