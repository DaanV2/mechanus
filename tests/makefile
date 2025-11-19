# Main setup command for both development and testing
dev-setup: setup start-services
	@echo "✅ Development environment ready!"
	@echo "Run 'make test' to execute tests"
	@echo "Run 'make local-server' to start the local server"

# Test setup command that ensures everything is ready for testing
test-setup: setup start-services
	@echo "✅ Test environment ready!"
	npm run test

# Install all dependencies
setup:
	@echo "📦 Installing dependencies..."
	npm install
	$(MAKE) browsers

# Install browser dependencies
browsers:
	@echo "🌐 Installing browsers..."
	npx playwright install

setup:
	-mkdir playwright-report

# Start all required services
start-services: setup
	@echo "🚀 Starting services..."
	-$(MAKE) stop-services
	@echo "Starting server..."
	$(MAKE) docker-server
	@echo "Waiting for services to be ready..."
	@timeout /t 5 >nul 2>&1 || sleep 5

# Start server in Docker with port mapping (works on all platforms)
docker-server: server

# Start server using Docker Compose
default-compose-file := docker-compose.yaml
compose-up:
	docker compose -f $(default-compose-file) up -d mechanus-server

compose-down:
	docker compose -f $(default-compose-file) down

# Start server in Docker Compose (replaces old server target)
server:
	docker compose -f $(default-compose-file) rm -sf mechanus-server || true
	docker compose -f $(default-compose-file) up -d mechanus-server

server-logs:
	-mkdir ./playwright-report/server
	docker-compose logs > ./playwright-report/server/logs.txt


# Start local server for development
local-server:
	cd .. && go run server/main.go server --log.level debug --web.static.folder "client/build"

# Stop services
stop-services:
	@echo "Stopping services..."
	docker compose -f $(default-compose-file) stop mechanus-server || true
	docker compose -f $(default-compose-file) rm -sf mechanus-server || true

# Run tests
test:
	npm run test

# Clean up everything
clean: stop-services
	@echo "🧹 Cleaning up..."

.PHONY: setup browsers server docker-server local-server test clean dev-setup test-setup setup start-services stop-services