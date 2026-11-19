# Server tests

The Go server has three kinds of tests. All of them use Ginkgo + Gomega and run with `just server test`.

| Kind | Where | What it tests |
| --- | --- | --- |
| Unit | `*_test.go` next to the code | One package on its own, no database. |
| Component | [`component/`](component) | The behaviour of a handler, service, policy or transport through its interface. The database is only setup, an in-memory SQLite one is enough. |
| Integration | [`integration/`](integration) | What gets stored in and read back from the database: queries, revocation, soft deletes, relations, round trips. These are the tests that can behave differently per database backend. |

The browser end-to-end tests (Playwright) live in [`tests/`](../../tests) at the repository root.

## Which one to write

- The assertions are about what ends up in the database, or a query has to be right → **integration**.
- The assertions are about responses, errors, policies or messages, and the database just has to exist → **component**.
- No database or other layers needed → **unit**, next to the code.

## Database backends

Component tests always use `util.CreateDatabase`, an in-memory SQLite database.

Integration tests use `util.CreateIntegrationDatabase`, which picks the backend from `TEST_DATABASE_TYPE`:

| `TEST_DATABASE_TYPE` | Database per test |
| --- | --- |
| `memory` (default) | A new in-memory SQLite database. This is what `just server test` uses. |
| `sqlite` | A SQLite file in a temporary directory. |
| `postgres` | A schema of its own (`test_<random>`), dropped afterwards. The server comes from `TEST_DATABASE_DSN`. |

Run them locally with:

```bash
just server test-integration            # sqlite
just server test-postgres-up            # throwaway Postgres on port 55432
just server test-integration postgres
just server test-postgres-down
```

CI runs the integration tests against sqlite and postgres (the `integration` job in `.github/workflows/server-checks.yaml`).

## Shared helpers

[`util/`](util) holds helpers both suites use, such as `CreateDatabase` / `CreateIntegrationDatabase` (a fresh database per test) and `CreateUser`.
