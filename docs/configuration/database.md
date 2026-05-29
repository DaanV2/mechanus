# Database

- [Database](#database)
  - [SQLite (default)](#sqlite-default)
  - [In-memory SQLite](#in-memory-sqlite)
  - [PostgreSQL](#postgresql)
  - [MySQL](#mysql)
  - [Key-value storage](#key-value-storage)

Mechanus supports four database backends, configured via CLI flags, environment variables, or the config file.

| Flag                           | Env var                      | Default        | Description                                              |
| ------------------------------ | ---------------------------- | -------------- | -------------------------------------------------------- |
| `--database.type`              | `DATABASE_TYPE`              | `sqlite`       | Backend: `sqlite`, `memory`, `postgres`, `mysql`         |
| `--database.dsn`               | `DATABASE_DSN`               | `db.sqlite`    | File path (SQLite) or connection string (Postgres/MySQL) |
| `--database.max-idle-conns`    | `DATABASE_MAX_IDLE_CONNS`    | `2`            | Max idle connections in the pool                         |
| `--database.max-open-conns`    | `DATABASE_MAX_OPEN_CONNS`    | `0` (no limit) | Max open connections                                     |
| `--database.conn-max-lifetime` | `DATABASE_CONN_MAX_LIFETIME` | `1h`           | Max time a connection may be reused                      |

## SQLite (default)

The default backend. If `dsn` is empty or the default `db.sqlite`, the path is resolved to the OS state directory (`~/.local/share/mechanus/db.sqlite` on Linux, `%APPDATA%\mechanus\db.sqlite` on Windows).

```yaml
database:
  type: sqlite
  dsn: /path/to/mechanus.db # omit to use the default OS state directory
```

Or with environment variables:

```bash
DATABASE_TYPE=sqlite
DATABASE_DSN=/path/to/mechanus.db
```

## In-memory SQLite

Intended for testing only — all data is lost on shutdown.

```yaml
database:
  type: memory
```

```bash
DATABASE_TYPE=memory
```

## PostgreSQL

Requires a [PostgreSQL DSN](https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters).

```yaml
database:
  type: postgres
  dsn: "host=localhost user=mechanus password=secret dbname=mechanus port=5432 sslmode=disable"
  max-idle-conns: 5
  max-open-conns: 20
  conn-max-lifetime: 30m
```

```bash
DATABASE_TYPE=postgres
DATABASE_DSN="host=localhost user=mechanus password=secret dbname=mechanus port=5432 sslmode=disable"
```

## MySQL

Requires a [MySQL DSN](https://github.com/go-sql-driver/mysql#dsn-data-source-name).

```yaml
database:
  type: mysql
  dsn: "mechanus:secret@tcp(localhost:3306)/mechanus?charset=utf8mb4&parseTime=True&loc=Local"
  max-idle-conns: 5
  max-open-conns: 20
  conn-max-lifetime: 30m
```

```bash
DATABASE_TYPE=mysql
DATABASE_DSN="mechanus:secret@tcp(localhost:3306)/mechanus?charset=utf8mb4&parseTime=True&loc=Local"
```

## Key-value storage

Mechanus also uses key-value storage for things like certificates. This can be configured separately — see `server/pkg/storage` for details. Backends: local file storage, or a table in the relational database.
