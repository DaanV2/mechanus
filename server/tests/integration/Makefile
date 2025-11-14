.PHONY: help start stop test clean logs

help: ## Display this help message
	@echo "Integration Tests - Makefile Commands"
	@echo "======================================"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

start: ## Start Grafana stack services
	@echo "Starting Grafana observability stack..."
	docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Services started. Access Grafana at http://localhost:3000"
	@echo "  Username: admin"
	@echo "  Password: admin"

stop: ## Stop Grafana stack services
	@echo "Stopping services..."
	docker compose down

test: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -timeout 5m ./...

test-with-stack: start ## Start services and run tests
	@echo "Running tests with Grafana stack..."
	@sleep 10
	$(MAKE) test || true
	@echo "\n=== Test Summary ==="
	@echo "Grafana Dashboard: http://localhost:3000"
	@echo "Tempo: http://localhost:3200"
	@echo "Loki: http://localhost:3100"
	@echo "Prometheus: http://localhost:9090"

clean: stop ## Clean up volumes and containers
	@echo "Cleaning up..."
	docker compose down -v
	@echo "Cleanup complete"

logs: ## View service logs
	docker compose logs -f
