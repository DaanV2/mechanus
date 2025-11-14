# Development

## Running the Server

Start the development server:

```bash
cd server
make start-server
```

## Testing

### Unit Tests

Run unit tests using Ginkgo:

```bash
cd server
make test
```

### Integration Tests

Integration tests verify the server with OpenTelemetry instrumentation and Grafana observability stack.

**Quick Start:**

```bash
cd server/tests/integration
make test-with-stack
```

This will:
1. Start the Grafana observability stack (Tempo, Loki, Prometheus, Grafana)
2. Run the integration tests
3. Display links to Grafana dashboards

**Manual Steps:**

```bash
# Start services
cd server/tests/integration
make start

# Wait for services to be ready (about 10-15 seconds)
sleep 15

# Run tests
make test

# View logs
make logs

# Stop services
make stop

# Clean up everything
make clean
```

**Services Included:**
- OpenTelemetry Collector (ports 4317/4318) - OTLP receiver
- Grafana Tempo (port 3200) - Distributed tracing
- Grafana Loki (port 3100) - Log aggregation
- Prometheus (port 9090) - Metrics collection
- Grafana (port 3000) - Unified dashboards

**Access Grafana:**
- URL: http://localhost:3000
- Username: `admin`
- Password: `admin`

For more details, see [server/tests/integration/README.md](../tests/integration/README.md)

## Building

Build the server:

```bash
cd server
make build
```

## Linting and Formatting

```bash
cd server
make lint
make format
```

## Full Checks

Run all checks (build, test, format, lint):

```bash
cd server
make checks
```