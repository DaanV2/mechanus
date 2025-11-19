
# Settings

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| database | object | see: [database](#database) |  |  |
| initialize | object | see: [initialize](#initialize) |  |  |
| log | object | see: [log](#log) |  |  |
| mdns | object | see: [mdns](#mdns) |  |  |
| otel | object | see: [otel](#otel) |  |  |
| server | object | see: [server](#server) |  |  |

## Database

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| conn-max-lifetime | duration | Sets the maximum amount of time a connection may be reused. If d <= 0, connections are not closed due to a connection's age. | 1h0m0s | DATABASE_CONN_MAX_LIFETIME |
| dsn | string | A datasource name, depends on type of database, but usually referes to file name or the connection string | db.sqlite | DATABASE_DSN |
| max-idle-conss | int | Sets the maximum number of connections in the idle connection pool. If n <= 0, no idle connections are retained. | 2 | DATABASE_MAX_IDLE_CONSS |
| max-open-conns | int | Sets the maximum number of open connections to the database. If n <= 0, then there is no limit on the number of open connections. | 0 | DATABASE_MAX_OPEN_CONNS |
| type | string | The type of database to connect/use: supported values: sqlite, postgres, mysql. (For testing purposes there is also inmemory) | sqlite | DATABASE_TYPE |

## Initialize

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| admin | object | see: [admin](#admin) |  |  |

### Admin

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| password | string | The admin password to use when initializing |  | INITIALIZE_ADMIN_PASSWORD |
| username | string | The admin username to use when initializing |  | INITIALIZE_ADMIN_USERNAME |

## Log

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| format | string | The format of the logging | text | LOG_FORMAT |
| level | string | The debug level, levels are: debug, info, warn, error, fatal | info | LOG_LEVEL |
| report-caller | bool | Whenever or not to output the file that outputs the log | false | LOG_REPORT_CALLER |

## Mdns

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| hostname | string | The host name to broadcast on | mechanus | MDNS_HOSTNAME |
| ipv6 | bool | Whenever or not to support ipv6 as well | false | MDNS_IPV6 |
| servicetype | string | The MDNS type to broadcast as | _http._tcp.local | MDNS_SERVICETYPE |

## Otel

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| enabled | bool | Enable OpenTelemetry tracing | false | OTEL_ENABLED |
| endpoint | string | OpenTelemetry collector endpoint (OTLP HTTP) | localhost:4318 | OTEL_ENDPOINT |
| insecure | bool | Use insecure connection to OTLP collector | true | OTEL_INSECURE |
| service-name | string | Service name for OpenTelemetry traces | mechanus-server | OTEL_SERVICE_NAME |

## Server

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| host | string | What host to bind on, such as: "", "localhost" or "0.0.0.0" |  | SERVER_HOST |
| port | uint16 | The port to server web traffic to | 8080 | SERVER_PORT |
| cors | object | see: [cors](#cors) |  |  |
| static | object | see: [static](#static) |  |  |

### Cors

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| allow-localhost | bool | Whenever or not as an origin, localhost are allowed | true | SERVER_CORS_ALLOW_LOCALHOST |
| allowed-origins |  | The origins that are allowed to be used by requesters, if empty will skip this header. Allowed strings are matched via prefix check | [*] | SERVER_CORS_ALLOWED_ORIGINS |

### Static

| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
| folder | string | The default files to serve | /web | SERVER_STATIC_FOLDER |
