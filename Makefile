GO ?= go
GO_ENV := CGO_ENABLED=0 GOTOOLCHAIN=local
DOCKER_COMPOSE ?= $(shell if command -v docker-compose >/dev/null 2>&1; then echo docker-compose; else echo docker compose; fi)
COMPOSE_PROJECT_NAME ?= task_service

.PHONY: test coverage generate-mocks integration-test integration-test-required run run-worker seed build docker-up docker-down

test:
	$(GO_ENV) $(GO) test ./...

coverage:
	mkdir -p .cache
	$(GO_ENV) $(GO) test -coverprofile=.cache/usecase.coverprofile -coverpkg=./internal/usecase/... ./internal/usecase/...
	$(GO_ENV) $(GO) tool cover -func=.cache/usecase.coverprofile | tail -n 40

generate-mocks:
	$(GO_ENV) $(GO) generate ./internal/usecase/... ./internal/adapter/http/handlers/...

integration-test:
	$(GO_ENV) $(GO) test -tags=integration ./...

integration-test-required:
	INTEGRATION_REQUIRE_DOCKER=1 $(GO_ENV) $(GO) test -tags=integration ./...

run:
	$(GO_ENV) $(GO) run ./cmd/api

run-worker:
	$(GO_ENV) $(GO) run ./cmd/worker

seed:
	$(GO_ENV) $(GO) run ./cmd/seed

build:
	$(GO_ENV) $(GO) build ./cmd/api ./cmd/worker ./cmd/seed

docker-up:
	$(DOCKER_COMPOSE) -p $(COMPOSE_PROJECT_NAME) up --build

docker-down:
	$(DOCKER_COMPOSE) -p $(COMPOSE_PROJECT_NAME) down -v
