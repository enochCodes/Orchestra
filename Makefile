# Orchestra — Makefile

.PHONY: help up down logs build clean test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo
	@echo 'Targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Start the full application stack
	docker compose up -d

dev-deps: ## Start only DB and Redis (for local core dev)
	docker compose up -d db redis

down: ## Stop the application stack
	docker compose down

logs: ## View logs from all services
	docker compose logs -f

build: ## Rebuild all containers
	docker compose build

clean: ## Remove containers and volumes
	docker compose down -v

test-core: ## Run core (Go) tests
	cd core && go test ./...

lint-ui: ## Run UI (Next.js) linting
	cd ui && npm run lint

vet: ## Run go vet on core
	cd core && go vet ./...
