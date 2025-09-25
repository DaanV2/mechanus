# 📝 Architecture TODOs

### ✅ Naming & Consistency

- [ ] Rename **`infrastructure/authentication/` → `infrastructure/authn/`** (authentication = authN).
- [ ] Keep **`engine/authz/`** as authorization (authZ).
- [ ] Review consistency across `authn` vs `authz` to make it explicit: _authN = who you are, authZ = what you can do_.

---

### ✅ Application Layer

- [ ] Review **`application/checks/`** — consider renaming to **`application/health/`** or **`application/bootstrap_checks/`** for clarity (denotes readiness/startup checks).
- [ ] Ensure use cases (services like `UserService`, `SessionService`) live only in `application/`, **not mixing DB calls directly**.
- [ ] Make sure startup checks (`ensure admin exists`) are written in **application/health** and delegate to engine + infra repos.

---

### ✅ Infrastructure Layer

- [ ] Confirm **repository implementations** live in `infrastructure/persistence/` and only expose **interfaces.UserRepository** contracts back up.
- [ ] Keep ORM/db “models” in `infrastructure/persistence/models/`.
- [ ] Ensure converters exist (`model → domain entity`, `entity → model`), so **engine never sees db models**.
- [ ] Verify generic **blob/key interfaces** (`Storage`) live in `infrastructure/storage/`, while _domain repositories_ stay in `persistence/`.

---

### ✅ `mechanus/` Directory

- [ ] Decide on the purpose of **`mechanus/`**. Options:
  - Merge into `components/` if it’s mainly DI/startup wiring.
  - Rename to **`runtime/`** or **`orchestrator/`** if it manages long‑running sessions, loops, or orchestration logic.
- [ ] Make its intent explicit in docs so contributors understand its scope.

---

### ✅ Documentation & Reports

- [ ] Add a short guideline to `docs/` (e.g. `docs/CONTRIBUTING.md`) explaining:
  - What belongs in `engine/`, `application/`, `infrastructure/`.
  - How repos and models should be separated.

---

### ✅ Nice-to-Haves

- [ ] Consider renaming **`application/checks` → `application/health/`**, and expose them both as:
  - CLI check (`your-project doctor`).
  - HTTP health endpoint (`/health`, `/ready`).

