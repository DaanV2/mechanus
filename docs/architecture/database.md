# Database

Mechanus uses [GORM](https://gorm.io) as its ORM. The default backend is SQLite, but PostgreSQL and MySQL are also supported via configuration.

## Supported backends

| Type             | Config value | Notes                                                                 |
| ---------------- | ------------ | --------------------------------------------------------------------- |
| SQLite           | `sqlite`     | Default; file path auto-resolved to the OS state directory if not set |
| In-memory SQLite | `memory`     | Testing only; data is lost on shutdown                                |
| PostgreSQL       | `postgres`   | Requires a DSN connection string                                      |
| MySQL            | `mysql`      | Requires a DSN connection string                                      |

## Configuration

Set via CLI flags, environment variables, or the config file:

| Flag                           | Default         | Description                                              |
| ------------------------------ | --------------- | -------------------------------------------------------- |
| `--database.type`              | `sqlite`        | Backend type                                             |
| `--database.dsn`               | `db.sqlite`     | File path (SQLite) or connection string (Postgres/MySQL) |
| `--database.max-idle-conns`    | `2`             | Max idle connections in the pool                         |
| `--database.max-open-conns`    | `0` (unlimited) | Max open connections                                     |
| `--database.conn-max-lifetime` | `1h`            | Max connection reuse duration                            |

For SQLite, if the DSN is empty or the default `db.sqlite`, the server resolves the path to the OS-appropriate state directory (e.g. `~/.local/share/mechanus/db.sqlite` on Linux, `%APPDATA%\mechanus\db.sqlite` on Windows).

See [configuration/config-file.md](../configuration/config-file.md) for config file locations.

## Schema

Migrations are applied automatically on startup via GORM's `AutoMigrate`. All tables use a UUID primary key generated before insert.

### Base model

All entities embed `Model`:

| Column       | Type                 | Description                 |
| ------------ | -------------------- | --------------------------- |
| `id`         | string (UUID)        | Primary key, auto-generated |
| `created_at` | timestamp            |                             |
| `updated_at` | timestamp            |                             |
| `deleted_at` | timestamp (nullable) | Soft-delete                 |
