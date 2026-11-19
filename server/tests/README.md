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

## Shared helpers

[`util/`](util) holds helpers both suites use, such as `CreateDatabase` (a fresh database per test) and `CreateUser`.
