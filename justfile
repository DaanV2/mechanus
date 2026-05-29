# Generate Go code from proto files
proto:
    buf generate

# Format proto files in place
proto-format:
    buf format --write

# Lint proto files
proto-lint:
    buf lint

# Run all checks: proto generation, formatting, linting, and image build
checks: proto proto-format proto-lint image

# Build the Docker image
image:
    docker build . -t mechanus

# Build the server binary
server:
    go build -o mechanus ./server/main.go
