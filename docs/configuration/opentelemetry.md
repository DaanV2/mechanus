# OpenTelemetry Configuration

Mechanus server supports OpenTelemetry for distributed telemetry. This allows you to observe and monitor the behavior of your server across different components and services.

## Configuration

OpenTelemetry telemetry is **disabled by default** and must be explicitly enabled through command-line flags or environment variables.

### Available Options

| Flag                  | Environment Variable | Default           | Description                                  |
| --------------------- | -------------------- | ----------------- | -------------------------------------------- |
| `--otel.enabled`      | `OTEL_ENABLED`       | `false`           | Enable OpenTelemetry telemetry               |
| `--otel.endpoint`     | `OTEL_ENDPOINT`      | `localhost:4318`  | OpenTelemetry collector endpoint (OTLP HTTP) |
| `--otel.service-name` | `OTEL_SERVICE_NAME`  | `mechanus-server` | Service name for traces                      |
| `--otel.insecure`     | `OTEL_INSECURE`      | `true`            | Use insecure connection to OTLP collector    |

## Usage

### Starting the server with telemetry enabled

```bash
./mechanus server --otel.enabled=true --otel.endpoint=localhost:4318
```

Or using environment variables:

```bash
export OTEL_ENABLED=true
export OTEL_ENDPOINT=localhost:4318
./mechanus server
```

### Using with OpenTelemetry Collector

To use OpenTelemetry telemetry, you'll need an OpenTelemetry Collector running. Here's a simple docker-compose example:

```yaml
version: "3"
services:
  otel-collector:
    image: otel/opentelemetry-collector:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4318:4318" # OTLP HTTP receiver
      - "55679:55679" # zpages extension

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686" # Jaeger UI
      - "14250:14250" # Jaeger gRPC
```

Example `otel-collector-config.yaml`:

```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:

exporters:
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true
  logging:
    loglevel: debug

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [jaeger, logging]
```

### Viewing Traces

After starting the server with telemetry enabled and an OpenTelemetry Collector + Jaeger backend:

1. Start your services: `docker-compose up -d`
2. Start the Mechanus server with telemetry: `./mechanus server --otel.enabled=true`
3. Make some requests to the server
4. Open Jaeger UI at http://localhost:16686
5. Select "mechanus-server" from the service dropdown
6. Click "Find Traces" to view traces

## What is Traced?

The OpenTelemetry integration automatically traces:

- **HTTP requests** to both the web and API servers
- **gRPC/Connect RPC calls** including login, user management, and other services
- **Request context propagation** across service boundaries

## Production Considerations

For production deployments:

1. Set `--otel.insecure=false` when using TLS with your collector
2. Configure appropriate sampling rates in your OpenTelemetry Collector
3. Ensure your collector can handle the trace volume
4. Consider using a managed telemetry backend (e.g., Jaeger, Zipkin, or cloud providers)

## Troubleshooting

### Traces not appearing

1. Verify the collector endpoint is correct and reachable
2. Check the collector logs for errors
3. Ensure `--otel.enabled=true` is set
4. Check Mechanus server logs for OpenTelemetry initialization messages

### High overhead

1. Configure sampling in your OpenTelemetry Collector
2. Adjust batch processor settings in the collector configuration
3. Consider head-based sampling strategies

