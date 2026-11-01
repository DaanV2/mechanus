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

# Run all checks: build, test, format, lint, and documentation
[group('test')]
checks: fix format build test format lint documentation

# Generate documentation
[group('script')]
documentation:
  go run ./scripts/documentation

# Start the server
[group('dev')]
start-server:
  go run main.go server
