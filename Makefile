.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev
dev: ## Run development server with hot reload
	@echo "Starting development server..."
	docker compose up --build

.PHONY: dev-detached
dev-detached: ## Run development server in background
	docker compose up -d --build

.PHONY: logs
logs: ## Show logs from all services
	docker compose logs -f

.PHONY: logs-analytics
logs-analytics: ## Show logs from analytics service only
	docker compose logs -f analytics

.PHONY: down
down: ## Stop all services
	docker compose down

.PHONY: clean
clean: down ## Stop services and remove volumes
	docker compose down -v

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: build
build: ## Build production Docker image
	docker build -t mcp-analytics:latest .

.PHONY: build-local
build-local: ## Build binary locally
	CGO_ENABLED=0 go build -o bin/analytics cmd/analytics/main.go

.PHONY: run-local
run-local: build-local ## Build and run locally (requires local services)
	./bin/analytics

.PHONY: migrate
migrate: ## Run database migrations
	@echo "Running PostgreSQL migrations..."
	docker compose exec postgres psql -U analytics -d mcp_analytics -f /docker-entrypoint-initdb.d/init.sql
	@echo "Creating Elasticsearch indices..."
	docker compose exec analytics go run scripts/create-indices.go

.PHONY: seed
seed: ## Seed databases with test data
	@echo "Seeding databases with test data..."
	docker compose exec analytics go run scripts/seed-data.go

.PHONY: elastic-health
elastic-health: ## Check Elasticsearch health
	curl -s http://localhost:9200/_cluster/health?pretty

.PHONY: elastic-indices
elastic-indices: ## List Elasticsearch indices
	curl -s http://localhost:9200/_cat/indices?v

.PHONY: kibana
kibana: ## Open Kibana in browser
	@echo "Opening Kibana..."
	open http://localhost:5601 || xdg-open http://localhost:5601

.PHONY: adminer
adminer: ## Open Adminer (PostgreSQL UI) in browser
	@echo "Opening Adminer..."
	open http://localhost:8082 || xdg-open http://localhost:8082

.PHONY: redis-cli
redis-cli: ## Connect to Redis CLI
	docker compose exec redis redis-cli

.PHONY: psql
psql: ## Connect to PostgreSQL CLI
	docker compose exec postgres psql -U analytics -d mcp_analytics

.PHONY: mongo-shell
mongo-shell: ## Connect to MongoDB shell
	docker compose exec mongodb mongosh mcp_analytics

.PHONY: deps
deps: ## Download Go dependencies
	go mod download
	go mod tidy

.PHONY: update-deps
update-deps: ## Update Go dependencies
	go get -u ./...
	go mod tidy

.PHONY: gen-docs
gen-docs: ## Generate API documentation
	swag init -g cmd/analytics/main.go -o docs/api

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: prod-build
prod-build: ## Build production image with version tag
	@VERSION=$$(git describe --tags --always --dirty); \
	docker build -t mcp-analytics:$$VERSION -t mcp-analytics:latest .
	@echo "Built mcp-analytics:$$VERSION"

.PHONY: prod-push
prod-push: prod-build ## Push production image to registry
	@VERSION=$$(git describe --tags --always --dirty); \
	docker tag mcp-analytics:$$VERSION ${DOCKER_REGISTRY}/mcp-analytics:$$VERSION; \
	docker tag mcp-analytics:latest ${DOCKER_REGISTRY}/mcp-analytics:latest; \
	docker push ${DOCKER_REGISTRY}/mcp-analytics:$$VERSION; \
	docker push ${DOCKER_REGISTRY}/mcp-analytics:latest

.PHONY: health-check
health-check: ## Check health of all services
	@echo "Checking service health..."
	@echo -n "Analytics: " && curl -s http://localhost:8081/health | jq -r .status || echo "DOWN"
	@echo -n "Elasticsearch: " && curl -s http://localhost:9200/_cluster/health | jq -r .status || echo "DOWN"
	@echo -n "PostgreSQL: " && docker compose exec -T postgres pg_isready -q && echo "UP" || echo "DOWN"
	@echo -n "MongoDB: " && docker compose exec -T mongodb mongosh --eval "db.adminCommand('ping')" --quiet > /dev/null && echo "UP" || echo "DOWN"
	@echo -n "Redis: " && docker compose exec -T redis redis-cli ping | grep -q PONG && echo "UP" || echo "DOWN"

.PHONY: load-test
load-test: ## Run load tests with k6
	@echo "Running load tests..."
	k6 run scripts/load-test.js

.PHONY: backup
backup: ## Backup all databases
	@mkdir -p backups
	@echo "Backing up PostgreSQL..."
	docker compose exec -T postgres pg_dump -U analytics mcp_analytics | gzip > backups/postgres-$$(date +%Y%m%d-%H%M%S).sql.gz
	@echo "Backing up MongoDB..."
	docker compose exec -T mongodb mongodump --db mcp_analytics --archive | gzip > backups/mongodb-$$(date +%Y%m%d-%H%M%S).gz
	@echo "Backing up Elasticsearch..."
	curl -XPUT "http://localhost:9200/_snapshot/backup_repo" -H 'Content-Type: application/json' -d'{"type": "fs","settings": {"location": "/backups"}}'
	@echo "Backups completed in ./backups/"

.PHONY: restore
restore: ## Restore from latest backup
	@echo "Restoring from latest backups..."
	@echo "Please specify which backup to restore"
	@ls -la backups/