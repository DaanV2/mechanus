# Go Integration Tests with OpenTelemetry & Grafana

This directory contains integration tests for the Mechanus Go server with OpenTelemetry instrumentation and Grafana observability stack.

## Overview

These integration tests verify that:
1. The Go server can export telemetry data (traces, logs, metrics) via OpenTelemetry
2. Telemetry data is properly ingested by the Grafana observability stack
3. Data is queryable through Grafana dashboards

## Architecture

```
┌─────────────┐     OTLP      ┌──────────────────┐
│  Go Server  │ ──────────────▶│ OTEL Collector   │
│ (Instrumented│               │  (Port 4317/8)   │
│   with OTEL) │               └──────────────────┘
└─────────────┘                       │ │ │
                                      │ │ │
                    ┌─────────────────┘ │ └──────────────┐
                    ▼                   ▼                 ▼
              ┌──────────┐        ┌────────┐      ┌────────────┐
              │  Tempo   │        │  Loki  │      │ Prometheus │
              │ (Traces) │        │ (Logs) │      │ (Metrics)  │
              └──────────┘        └────────┘      └────────────┘
                    │                   │                 │
                    └───────────────────┴─────────────────┘
                                        │
                                        ▼
                                  ┌──────────┐
                                  │ Grafana  │
                                  │(Dashboard│
                                  └──────────┘
```

## Prerequisites

- Docker and Docker Compose
- Go 1.25.1 or later
- Make (optional, for convenience commands)

## Services

The docker-compose stack includes:

- **OpenTelemetry Collector** (port 4317/4318) - Receives and exports telemetry data
- **Grafana Tempo** (port 3200) - Distributed tracing backend
- **Grafana Loki** (port 3100) - Log aggregation system
- **Prometheus** (port 9090) - Metrics collection and storage
- **Grafana** (port 3000) - Unified visualization dashboard

## Quick Start

### Using Make (Recommended)

```bash
# Start the Grafana stack and run tests
make test-with-stack

# Or run individually:
make start      # Start services
make test       # Run tests
make logs       # View service logs
make stop       # Stop services
make clean      # Clean up everything
```

### Manual Commands

```bash
# Start services
docker compose up -d

# Wait for services to be ready (about 10-15 seconds)
sleep 15

# Run tests
go test -v ./...

# Stop services
docker compose down
```

## Environment Variables

The tests support the following environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `TEMPO_ENDPOINT` | `http://localhost:3200` | Tempo API endpoint |
| `LOKI_ENDPOINT` | `http://localhost:3100` | Loki API endpoint |
| `PROMETHEUS_ENDPOINT` | `http://localhost:9090` | Prometheus API endpoint |
| `GRAFANA_ENDPOINT` | `http://localhost:3000` | Grafana UI endpoint |

## OpenTelemetry Configuration

The Go server should be configured with the following OpenTelemetry environment variables:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_SERVICE_NAME=mechanus-server
```

## Test Coverage

The integration tests verify:

1. **Service Health**: All services are up and responding
2. **Data Source Connectivity**: Grafana can connect to Tempo, Loki, and Prometheus
3. **OTLP Endpoints**: OpenTelemetry Collector is ready to receive data
4. **API Accessibility**: All service APIs are accessible and responding correctly
5. **Server Integration**: The Go server runs with OpenTelemetry enabled and exports traces to Tempo
6. **Trace Generation**: HTTP requests generate traces that appear in Tempo
7. **Trace Context**: The server correctly handles W3C trace context headers
8. **Configuration**: OpenTelemetry can be disabled without errors

## Viewing Telemetry Data

After running tests, access the Grafana dashboard:

1. Open http://localhost:3000 in your browser
2. Login with username `admin` and password `admin`
3. Navigate to Explore to view:
   - **Traces**: Select "Tempo" as the data source
   - **Logs**: Select "Loki" as the data source
   - **Metrics**: Select "Prometheus" as the data source

### Direct URLs

- Grafana UI: http://localhost:3000
- Tempo API: http://localhost:3200
- Loki API: http://localhost:3100
- Prometheus UI: http://localhost:9090
- OTLP Receiver (gRPC): http://localhost:4317
- OTLP Receiver (HTTP): http://localhost:4318

## CI/CD Integration

See `.github/workflows/go-integration-tests.yaml` for the GitHub Actions workflow configuration.

## Troubleshooting

### Services not starting

```bash
# Check service logs
docker compose logs

# Check specific service
docker compose logs tempo
docker compose logs loki
docker compose logs prometheus
```

### Tests failing

```bash
# Verify services are healthy
curl http://localhost:3200/ready  # Tempo
curl http://localhost:3100/ready  # Loki
curl http://localhost:9090/-/ready  # Prometheus

# Check OTLP collector
curl http://localhost:8888/metrics
```

### Clean slate

```bash
# Remove all containers and volumes
make clean

# Or manually:
docker compose down -v
```

## Development

To add new integration tests:

1. Create a new `*_test.go` file in this directory
2. Use Ginkgo/Gomega test framework (consistent with the project)
3. Follow existing patterns for service health checks and API calls
4. Document any new environment variables or requirements

## Notes

- The tests assume OpenTelemetry instrumentation is already implemented in the Go server
- Actual telemetry data from the application will appear during integration test runs
- The Grafana stack is configured for development/testing, not production use
- All services use default/test credentials - do not use in production
