# Containerization

This repo ships Docker images for each component:

- **api-gateway-go** (Go, distroless, nonroot) → port 8080
- **service-python** (FastAPI) → port 8001
- **service-node** (Express) → port 8002

## Build locally

```bash
docker build -t hpag/api-gateway-go:dev ./api-gateway-go
docker build -t hpag/service-python:dev ./service-python
docker build -t hpag/service-node:dev ./service-node

