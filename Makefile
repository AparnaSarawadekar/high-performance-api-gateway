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

