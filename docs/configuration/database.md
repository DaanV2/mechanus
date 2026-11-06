# Database

- [Database](#database)
  - [SQLite (default)](#sqlite-default)
  - [In-memory SQLite](#in-memory-sqlite)
  - [PostgreSQL](#postgresql)
  - [MySQL](#mysql)
  - [What is stored where](#what-is-stored-where)

Mechanus supports four database backends, configured via CLI flags, environment variables, or the [config file](./config-file.md).

| Flag                           | Env var                      | Default        | Description                                              |
| ------------------------------ | ---------------------------- | -------------- | -------------------------------------------------------- |
| `--database.type`              | `DATABASE_TYPE`              | `sqlite`       | Backend: `sqlite`, `memory`, `postgres`, `mysql`         |
| `--database.dsn`               | `DATABASE_DSN`               | `db.sqlite`    | File path (SQLite) or connection string (Postgres/MySQL) |
| `--database.max-idle-conns`    | `DATABASE_MAX_IDLE_CONNS`    | `2`            | Max idle connections in the pool                         |
| `--database.max-open-conns`    | `DATABASE_MAX_OPEN_CONNS`    | `0` (no limit) | Max open connections                                     |
| `--database.conn-max-lifetime` | `DATABASE_CONN_MAX_LIFETIME` | `1h`           | Max time a connection may be reused                      |

The database schema is migrated automatically on every start.

## SQLite (default)

The default backend, built into the binary (pure Go, no cgo or system library needed). If `dsn` is empty or the default `db.sqlite`, the file is placed in the state folder:

| OS      | Default database file                                              |
| ------- | ------------------------------------------------------------------ |
| Linux   | `$XDG_STATE_HOME/mechanus/db.sqlite` (default `~/.local/state/mechanus/db.sqlite`) |
| macOS   | `~/Library/Application Support/mechanus/State/db.sqlite`           |
| Windows | `%LOCALAPPDATA%\mechanus\db.sqlite`                                |

If a `.local` folder exists in the working directory, it is used as the state folder instead (`./.local/db.sqlite`). The Docker image runs from `/app`, so mounting a volume at `/app/.local` keeps the database, see [Run the Docker image](../getting-started/docker.md#what-lives-where). `mechanus config paths` shows the folders on your machine.

Any other `dsn` is used as the file path; a relative path is resolved against the working directory.

```yaml
database:
  type: sqlite
  dsn: /path/to/mechanus.db # omit to use the state folder
```

Or with environment variables:

```bash
DATABASE_TYPE=sqlite
DATABASE_DSN=/path/to/mechanus.db
```

## In-memory SQLite

Intended for testing only — all data is lost on shutdown. `dsn` is ignored, and the pool is fixed to a single connection (every new connection to an in-memory SQLite database would be a separate, empty database), overriding the pool settings.

```yaml
database:
  type: memory
```

```bash
DATABASE_TYPE=memory
```

## PostgreSQL

Uses the [pgx](https://github.com/jackc/pgx) driver. The DSN is either key/value pairs or a URL, see [pgconn.ParseConfig](https://pkg.go.dev/github.com/jackc/pgx/v5/pgconn#ParseConfig) for every parameter. Create the database first; Mechanus creates the tables.

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
DATABASE_DSN="postgres://mechanus:secret@localhost:5432/mechanus?sslmode=disable"
```

## MySQL

Uses the [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql#dsn-data-source-name) DSN format. Keep `parseTime=True`, or timestamps can't be read. Create the database first; Mechanus creates the tables.

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

## What is stored where

The database holds users, campaigns, characters, invites, paired devices, issued token IDs and a key-value table. That key-value table stores the runtime state of scenes (`scenes.State`, under keys prefixed `state/scene/`).

Not everything is in the database, and this isn't configurable:

- **Scene and room configuration** (`scenes/`, `rooms/`, `shared/`) are files in the config folder, see [config file](./config-file.md) and [battle maps](./battle-maps.md).
- **Screen state** is kept as files in the `screens/` folder of the state folder, next to the default SQLite file.

So with PostgreSQL or MySQL you still want to keep (or back up) the state folder.
