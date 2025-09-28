# Settings

| Name       | Type   | Description                    | Default | Env |
| ---------- | ------ | ------------------------------ | ------- | --- |
| api        | object | see: [api](#api)               |         |     |
| database   | object | see: [database](#database)     |         |     |
| initialize | object | see: [initialize](#initialize) |         |     |
| log        | object | see: [log](#log)               |         |     |
| mdns       | object | see: [mdns](#mdns)             |         |     |
| web        | object | see: [web](#web)               |         |     |

## Api

| Name | Type   | Description                                                 | Default | Env      |
| ---- | ------ | ----------------------------------------------------------- | ------- | -------- |
| host | string | What host to bind on, such as: "", "localhost" or "0.0.0.0" |         | API_HOST |
| port | uint16 | The port to server api traffic to                           | 8666    | API_PORT |
| cors | object | see: [cors](#cors)                                          |         |          |

### Cors

| Name            | Type | Description                                                                                                                         | Default | Env                      |
| --------------- | ---- | ----------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------------------ |
| allow-localhost | bool | Whenever or not as an origin, localhost are allowed                                                                                 | true    | API_CORS_ALLOW_LOCALHOST |
| allowed-origins |      | The origins that are allowed to be used by requesters, if empty will skip this header. Allowed strings are matched via prefix check | [*]     | API_CORS_ALLOWED_ORIGINS |

## Database

| Name            | Type     | Description                                                                                                                       | Default   | Env                      |
| --------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------- | --------- | ------------------------ |
| connmaxlifetime | duration | Sets the maximum amount of time a connection may be reused. If d <= 0, connections are not closed due to a connection's age.      | 1h0m0s    | DATABASE_CONNMAXLIFETIME |
| dsn             | string   | A datasource name, depends on type of database, but usually referes to file name or the connection string                         | db.sqlite | DATABASE_DSN             |
| maxidleconns    | int      | Sets the maximum number of connections in the idle connection pool. If n <= 0, no idle connections are retained.                  | 2         | DATABASE_MAXIDLECONNS    |
| maxopenconns    | int      | Sets the maximum number of open connections to the database. If n <= 0, then there is no limit on the number of open connections. | 0         | DATABASE_MAXOPENCONNS    |
| type            | string   | The type of database to connect/use: supported values: sqlite, postgres, mysql. (For testing purposes there is also inmemory)     | sqlite    | DATABASE_TYPE            |

## Initialize

| Name  | Type   | Description          | Default | Env |
| ----- | ------ | -------------------- | ------- | --- |
| admin | object | see: [admin](#admin) |         |     |

### Admin

| Name     | Type   | Description                                 | Default | Env                       |
| -------- | ------ | ------------------------------------------- | ------- | ------------------------- |
| password | string | The admin password to use when initializing |         | INITIALIZE_ADMIN_PASSWORD |
| username | string | The admin username to use when initializing |         | INITIALIZE_ADMIN_USERNAME |

## Log

| Name          | Type   | Description                                                  | Default | Env               |
| ------------- | ------ | ------------------------------------------------------------ | ------- | ----------------- |
| format        | string | The format of the logging                                    | text    | LOG_FORMAT        |
| level         | string | The debug level, levels are: debug, info, warn, error, fatal | info    | LOG_LEVEL         |
| report-caller | bool   | Whenever or not to output the file that outputs the log      | false   | LOG_REPORT_CALLER |

## Mdns

| Name        | Type   | Description                             | Default            | Env              |
| ----------- | ------ | --------------------------------------- | ------------------ | ---------------- |
| hostname    | string | The host name to broadcast on           | mechanus           | MDNS_HOSTNAME    |
| ipv6        | bool   | Whenever or not to support ipv6 as well | false              | MDNS_IPV6        |
| servicetype | string | The MDNS type to broadcast as           | \_http.\_tcp.local | MDNS_SERVICETYPE |

## Web

| Name   | Type   | Description                                                 | Default | Env      |
| ------ | ------ | ----------------------------------------------------------- | ------- | -------- |
| host   | string | What host to bind on, such as: "", "localhost" or "0.0.0.0" |         | WEB_HOST |
| port   | uint16 | The port to server web traffic to                           | 8080    | WEB_PORT |
| static | object | see: [static](#static)                                      |         |          |

### Static

| Name   | Type   | Description                | Default | Env               |
| ------ | ------ | -------------------------- | ------- | ----------------- |
| folder | string | The default files to serve | /web    | WEB_STATIC_FOLDER |

