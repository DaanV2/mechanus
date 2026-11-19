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
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
```

### Viewing Traces

After starting the server with telemetry enabled and an OpenTelemetry Collector + Jaeger backend:

1. Start your services: `docker-compose up -d`
2. Start the Mechanus server with telemetry: `./mechanus server --otel.enabled=true`
3. Make some requests to the server
4. Open Jaeger UI at http://localhost:16686
5. Select "mechanus-server" from the service dropdown
6. Click "Find Traces" to view traces

### Local development (Grafana)

`just dev` also starts [`grafana/otel-lgtm`](../../docker/telemetry.yaml), which receives traces, logs and metrics.
Open Grafana at http://localhost:3000 (admin / admin), go to **Explore** and pick a data source:

- **Tempo**: traces, search for service `mechanus`
- **Loki**: logs, query `{service_name="mechanus"}`; a log line with a `trace_id` links to its trace
- **Prometheus**: metrics, e.g. `http_server_request_duration_seconds_count{job="mechanus"}`

#### Dashboards

Dashboards live as JSON files in [`docker/grafana/dashboards/`](../../docker/grafana/dashboards) and are provisioned
into the **Mechanus** folder by [`docker/grafana/dashboards.yaml`](../../docker/grafana/dashboards.yaml).
Grafana rescans the folder every 10 seconds, so adding or editing a file needs no restart.

To change a dashboard: edit it in Grafana, then **Export → Export as JSON** (leave "Export for sharing externally" off,
so the data source UIDs `prometheus`, `loki` and `tempo` stay as they are) and overwrite the file in the repo.
Saving in the UI alone only stores the change in Grafana's database (`docker/data/`); the next change to the file overwrites it.
Give every new dashboard a fixed `uid` so links to it stay stable.

## What is Traced, Logged and Measured?

The OpenTelemetry integration automatically captures:

### Traces

- **HTTP requests** to both the web and API servers
- **gRPC/Connect RPC calls** including login, user management, and other services
- **Database queries** through the GORM plugin
- **Request context propagation** across service boundaries

### Logs

- **Application logs** from the charm.sh logger are forwarded to OpenTelemetry, the console output stays as configured
- Logs include all structured attributes (key-value pairs), plus the logger prefix and caller
- Log severity levels are mapped to OpenTelemetry severity levels
- Lines logged through `logging.From(ctx)` while a span is active carry its trace and span id, correlating logs and traces

Under the hood the default logger writes JSON into a bridge (`infrastructure/logging/bridge.go`) that hands every line
to OpenTelemetry and re-renders it on the console. Loggers derived from `log.Default()` before telemetry is set up are not forwarded.

### Metrics

- **HTTP server**: request duration and body sizes (`http_server_*`)
- **RPC server**: call duration per procedure and status (`rpc_server_*`)
- **Database**: connection pool statistics (`go_sql_*`)
- **Go runtime**: memory, GC, goroutines (`go_*`)

Metrics are pushed every 60 seconds; set `OTEL_METRIC_EXPORT_INTERVAL` (milliseconds) to change that.

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

