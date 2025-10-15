COMPOSE_FILE=docker-compose.dev.yml

.PHONY: up down rebuild logs ps smoke clean

up:
	docker compose -f $(COMPOSE_FILE) up -d

down:
	docker compose -f $(COMPOSE_FILE) down

rebuild:
	docker compose -f $(COMPOSE_FILE) build --no-cache
	docker compose -f $(COMPOSE_FILE) up -d --force-recreate

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

ps:
	docker compose -f $(COMPOSE_FILE) ps

smoke:
	@echo "— Hitting service-python —"
	@curl -s http://localhost:$(PY_PORT)/health || true
	@echo "\n— Hitting service-node —"
	@curl -s http://localhost:$(NODE_PORT)/health || true
	@echo "\n— Hitting api-gateway —"
	@curl -s http://localhost:$(GATEWAY_PORT)/healthz || true
	@echo "\n— Sample routed calls (if your gateway maps these) —"
	@echo "GET /py/hello"
	@curl -s http://localhost:$(GATEWAY_PORT)/py/hello || true
	@echo "\nGET /node/hello"
	@curl -s http://localhost:$(GATEWAY_PORT)/node/hello || true

clean:
	docker compose -f $(COMPOSE_FILE) down -v --remove-orphans

perf:baseline:
	@echo "Running baseline (40 VUs, 20s)…"
	K6_SUMMARY_OUT=tests/load/perf_baseline_summary.json \
	k6 run --vus 40 --duration 20s tests/load/cache_test.js | tee tests/load/baseline_console.txt

perf:cached:
	@echo "Warming cache…"
	curl -sS http://localhost:8080/slow >/dev/null || true
	curl -sS http://localhost:8080/slow >/dev/null || true
	@echo "Running cached (40 VUs, 20s)…"
	K6_SUMMARY_OUT=tests/load/perf_cached_summary.json \
	k6 run --vus 40 --duration 20s tests/load/cache_test.js | tee tests/load/perf_cached_console.txt

perf:report:
	python3 scripts/perf_report.py

perf:all: perf:baseline perf:cached perf:report