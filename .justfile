set windows-shell := ["powershell.exe", "-c"]

mod client
mod server

default:
  just --list

[group('dev')]
dev:
  docker compose watch

[group('dev')]
down:
  just dev-stop

[group('dev')]
dev-stop:
  docker compose -f docker-compose.yaml down

# Build the Docker image
[group('build')]
image:
  docker build . -t mechanus

# Build the server binary
[group('build')]
binary:
  go build -o mechanus ./server/main.go

# Generate Go code from proto files
[group('proto')]
proto:
  buf generate

# Format proto files in place
[group('proto')]
proto-format:
  buf format --write

# Lint proto files
[group('proto')]
proto-lint:
  buf lint

# Run all checks: proto generation, formatting, linting, and image build
checks: proto proto-format proto-lint image

