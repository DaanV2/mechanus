set windows-shell := ["powershell.exe", "-c"]

set dotenv-load := true

default:
  just --list


# Build all packages
[group('build')]
build:
  go build ./...

  
# Run code generation
[group('build')]
generate:
  go generate ./...

# Run all tests with coverage
[group('test')]
test:
  go tool ginkgo run --vv --output-dir reports --cover --covermode atomic -coverprofile=coverage.out ./...

# Run the integration tests against a database backend: memory, sqlite or postgres.
# Postgres needs a running server, see `test-postgres-up`.
[group('test')]
test-integration $TEST_DATABASE_TYPE="sqlite" $TEST_DATABASE_DSN="host=localhost port=55432 user=mechanus password=mechanus_test dbname=mechanus_test sslmode=disable":
  go tool ginkgo run ./tests/integration/...

# Start a throwaway Postgres for `just server test-integration postgres` (port 55432, test credentials only)
[group('test')]
test-postgres-up:
  docker run -d --rm --name mechanus-test-postgres -p 55432:5432 -e POSTGRES_USER=mechanus -e POSTGRES_PASSWORD=mechanus_test -e POSTGRES_DB=mechanus_test postgres:18-alpine

# Stop the Postgres started by `test-postgres-up`
[group('test')]
test-postgres-down:
  docker stop mechanus-test-postgres

# Generate HTML coverage report and open in browser
[group('test')]
coverage-report:
  go test ./... --cover -coverprofile=reports/coverage.out --covermode atomic --coverpkg=./...
  go tool cover -html=reports/coverage.out

# Build with assembly output for inspection
[group('util')]
assembly:
  go build -gcflags="-S" ./mechanus/screens > cmd.S 2>&1

[group('util')]
fix:
  go fix ./...

alias fmt := format

# Format Go source files
[group('util')]
format:
  go fmt ./...

# Lint and auto-fix code
[group('util')]
lint:
  go tool golangci-lint run -v --fix

# Report suspicious constructs
[group('util')]
vet:
  go vet ./...

# Scan dependencies and the standard library for known vulnerabilities
[group('util')]
vulncheck:
  go tool govulncheck ./...

# Run all checks: build, vet, test, format, lint, vulnerabilities, and documentation
[group('test')]
checks: fix format build vet test format lint vulncheck documentation

# Generate documentation
[group('script')]
documentation:
  go run ./scripts/documentation

# Start the server
[group('dev')]
start-server:
  go run main.go server
