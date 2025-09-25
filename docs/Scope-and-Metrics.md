# Scope and Success Metrics

This document defines the scope, workload model, and performance targets for the **High-Performance API Gateway for AI Workloads** project.

---

## Project Scope

- **Gateway (Go)**  
  - Reverse proxy and routing  
  - Rate limiting and throttling  
  - Caching (in-memory → Redis)  
  - Health checks  

- **Upstream Services**  
  - `service-python` (FastAPI) → `/infer` endpoint simulating AI inference latency  
  - `service-node` (Express) → `/logs`, `/metrics`, `/healthz`  

- **Infrastructure**  
  - Docker containers for each service  
  - Local orchestration with `docker-compose`  
  - Kubernetes manifests (Kind/Minikube) in later phase  

- **Observability**  
  - Structured JSON logs  
  - Prometheus metrics endpoints  
  - (Optional) OpenTelemetry tracing  

- **CI/CD**  
  - GitHub Actions for build, lint, and test  
  - Continuous deployment to Koyeb (no credit card required)

---

## Workload Model

- **Target volume**: ≥ **1,000,000 requests/day** (validated via synthetic load tests)  
- **Average load**: ~11.6 RPS sustained  
- **Peak traffic**: bursts of 200–250 RPS for 5–10 minutes, repeated hourly  
- **Request mix**:  
  - 80% `GET /infer?prompt=...` (cacheable)  
  - 20% `POST /infer` (non-cacheable)  
- **Payload size**: 0.5–2 KB JSON requests/responses

---

## Success Metrics

**Latency (end-to-end via gateway)**  
- Baseline: p95 ≤ 200 ms, p99 ≤ 400 ms @ 150 RPS  
- Optimized: p95 ≤ 140 ms, p99 ≤ 280 ms @ 200 RPS  

**Throughput**  
- Baseline: ≥ 120 RPS  
- Optimized: ≥ 160 RPS (**~33% gain**)  

**Error Rate**  
- < 0.1% 5xx errors under target load  
- Throttling (HTTP 429) < 2% during peak  

**Caching Effectiveness**  
- Target hit ratio: 60–80% on repeated GET prompts  

**Uptime (demo target)**  
- ≥ 99.9% success over 48h synthetic checks  

---

## Rate Limiting Policy

- **Global token bucket**: 150 RPS steady, 300 burst  
- **Per-IP limit**: 20 RPS steady, 40 burst  
- Exceeding limits returns HTTP 429 with `Retry-After` header

---

## Measurement Method

- **Load testing**: k6 with steady, spike, and soak scenarios  
- **Thresholds enforced in test config**:  
  - `http_req_failed < 0.001`  
  - `p(95) < 200` (baseline) → `p(95) < 140` (optimized)  

- **Test runs**:  
  - Baseline (no cache, no rate limiting)  
  - Optimized (rate limiting, caching, connection pooling)  

- **Artifacts**:  
  - `/docs/PerfReport_v1.md` (baseline results)  
  - `/docs/PerfReport_v2.md` (optimized results)  

Each report will include RPS, latency (p50/p90/p95/p99), error rate, CPU/memory usage, and cache hit ratio.

---

## Observability Metrics

Gateway exports via Prometheus:  
- `http_requests_total{route=...}`  
- `request_duration_seconds_bucket`  
- `rate_limit_dropped_total`  
- `cache_hits_total`, `cache_misses_total`  
- `upstream_errors_total`

---

## Definition of Done for Scope

- [x] Scope documented here  
- [x] Performance targets added to README  
- [x] GitHub Issue created and closed for Step 1  
- [ ] Baseline k6 test skeleton committed (`tests/load/k6_baseline.js`)
