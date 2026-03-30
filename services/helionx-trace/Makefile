

COMPOSE=docker compose
DB_CONTAINER=helionx-postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=helionx
DB_USER=helionx
DB_PASSWORD=helionx
DB_SSLMODE=disable
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
INIT_SQL=./db/init.sql

.PHONY: help up down restart logs ps psql db-init db-reset

help:
	@echo "Available targets:"
	@echo "  make up        - start postgres in docker compose"
	@echo "  make down      - stop docker compose services"
	@echo "  make restart   - restart docker compose services"
	@echo "  make logs      - tail postgres logs"
	@echo "  make ps        - show docker compose service status"
	@echo "  make psql      - open psql shell"
	@echo "  make db-init   - apply SQL from $(INIT_SQL) manually"
	@echo "  make db-reset  - destroy containers and postgres volume"

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) down
	$(COMPOSE) up -d

logs:
	$(COMPOSE) logs -f postgres

ps:
	$(COMPOSE) ps

psql:
	PGPASSWORD=$(DB_PASSWORD) psql "host=$(DB_HOST) port=$(DB_PORT) dbname=$(DB_NAME) user=$(DB_USER) sslmode=$(DB_SSLMODE)"

db-init:
	PGPASSWORD=$(DB_PASSWORD) psql "host=$(DB_HOST) port=$(DB_PORT) dbname=$(DB_NAME) user=$(DB_USER) sslmode=$(DB_SSLMODE)" -f $(INIT_SQL)

db-reset:
	$(COMPOSE) down -v

test:
	\t./scripts/test.sh