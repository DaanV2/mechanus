set dotenv-load := true

# Build with assembly output for inspection
assembly:
  go build -gcflags="-S" ./mechanus/screens > cmd.S 2>&1

# Build all packages
build:
  go build ./...

# Run all tests with coverage
test:
  go tool ginkgo run --vv --output-dir reports --cover --covermode atomic -coverprofile=coverage.out ./...

# Generate HTML coverage report and open in browser
coverage-report:
  go test ./... --cover -coverprofile=reports/coverage.out --covermode atomic --coverpkg=./...
  go tool cover -html=reports/coverage.out

# Run code generation
generate:
  go generate ./...

# Lint and auto-fix code
lint:
  go tool golangci-lint run -v --fix

# Format Go source files
format:
  go fmt ./...

# Run all checks: build, test, format, lint, and documentation
checks: build test format lint documentation

# Generate documentation
documentation:
  go run ./scripts/documentation

# Start the server
start-server:
  go run main.go server
