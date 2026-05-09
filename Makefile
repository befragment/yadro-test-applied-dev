SERVICE_CMD := ./cmd/service
DOCKER_COMPOSE := docker compose

.PHONY: help
.PHONY: build build-service run run-service
.PHONY: test test-nocache tcoverage
.PHONY: migrate-create up migrate-up migrate-down migrate-status seed
.PHONY: dc-up dc-down dc-psql
.PHONY: fmt lint swag test-e2e

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------

help:
	@echo "Available targets:"
	@echo ""
	@echo "  build           - build HTTP service binary (alias: build-service)"
	@echo "  run             - run HTTP service with go run (alias: run-service)"
	@echo "  test            - run go test ./... -race"
	@echo "  test-nocache    - run go test ./... -count=1"
	@echo "  tcoverage       - generate and open html with test coverage report"
	@echo "  test-e2e        - run e2e tests in docker (compose down -v, then up --build)"
	@echo ""
	@echo "  fmt             - run golangci-lint fmt ./..."
	@echo "  lint            - run golangci-lint run"
	@echo "  swag            - generate swagger docs"
	@echo ""
	@echo "  up              - start postgres, apply migrations, then start service"
	@echo "  seed            - apply seed data to DB"
	@echo "  migrate-create  - create new migration (usage: make migrate-create NAME=migration_name)"
	@echo "  migrate-up      - apply DB migrations in container"
	@echo "  migrate-down    - rollback last DB migration in container"
	@echo "  migrate-status  - show DB migrations status in container"
	@echo ""
	@echo "  dc-up           - run docker compose up -d"
	@echo "  dc-down         - run docker compose down"
	@echo "  dc-psql         - run psql in postgres container"

# ------------------------------------------------------------------------------
# Build & run
# ------------------------------------------------------------------------------

up:
	$(DOCKER_COMPOSE) up -d postgres
	$(MAKE) migrate-up
	$(DOCKER_COMPOSE) up -d --build service

build: build-service
build-service:
	go build -o bin/service $(SERVICE_CMD)

run: run-service
run-service:
	go run $(SERVICE_CMD)

test:
	go list ./... | sed '/\/tests\/e2e/d' | xargs go test -race

test-nocache:
	go list ./... | sed '/\/tests\/e2e/d' | xargs go test -count=1

tcoverage:
	go list ./... | sed '/\/tests\/e2e/d' | xargs go test -coverprofile=coverage.out -coverpkg=./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

test-e2e:
	$(DOCKER_COMPOSE) -f docker-compose.e2e.yaml down -v
	$(DOCKER_COMPOSE) -f docker-compose.e2e.yaml up --build --abort-on-container-exit --exit-code-from e2e-tests

fmt:
	golangci-lint fmt ./...

lint:
	golangci-lint run

swag:
	swag init -g cmd/service/main.go -o docs

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	goose -dir ./migrations create $(NAME) sql

migrate-up:
	$(DOCKER_COMPOSE) run --rm service sh ./scripts/migrate.sh up

migrate-down:
	$(DOCKER_COMPOSE) run --rm service sh ./scripts/migrate.sh down

migrate-status:
	$(DOCKER_COMPOSE) run --rm service sh ./scripts/migrate.sh status

seed:
	sh ./scripts/seed.sh

# Docker shortcuts
dc-up:
	$(DOCKER_COMPOSE) up -d

dc-down:
	$(DOCKER_COMPOSE) down

dc-psql:
	sh ./scripts/psql-container.sh
