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

# Start all required services
start-services:
	@echo "🚀 Starting services..."
	-$(MAKE) stop-services
	@echo "Starting server..."
	$(MAKE) docker-server
	@echo "Waiting for services to be ready..."
	@timeout /t 5 >nul 2>&1 || sleep 5

# Start server in Docker with host networking (for Linux)
docker-server-linux:
	docker run --env-file .env -d --name mechanus-server --network host mechanus server --log.level debug --log.format text --log.report-caller

# Start server in Docker with port mapping (works on all platforms)
docker-server: server

# Start server in Docker with port mapping
server:
	-docker rm mechanus-server
	docker run --name mechanus-server --env-file .env -d -p 8080:8080 -p 8443:8443 -p 8666:8666 mechanus ./mechanus server --log.level debug --log.format text --log.report-caller

# Start local server for development
local-server:
	cd .. && go run server/main.go server --log.level debug --web.static.folder "client/build"

# Stop services
stop-services:
	@echo "Stopping services..."
	-docker stop mechanus-server >nul 2>&1 || docker stop mechanus-server >/dev/null 2>&1
	-docker rm mechanus-server >nul 2>&1 || docker rm mechanus-server >/dev/null 2>&1

# Run tests
test:
	npm run test

# Clean up everything
clean: stop-services
	@echo "🧹 Cleaning up..."

.PHONY: setup browsers server docker-server docker-server-linux local-server test clean dev-setup test-setup setup start-services stop-services