# Authentication

Mechanus uses JWT (JSON Web Tokens) for user authentication, signed with RSA-512 keys managed by the server. Device clients authenticate with API keys instead.

## JWT flow

```ascii
Client                            Server
  │                                 │
  │── POST /users.v1.LoginService ─►│  credentials (username + password)
  │                                 │
  │                                 │  verify password hash
  │                                 │  get or create JTI (JWT ID) for user
  │                                 │  sign JWT with RSA-512 private key
  │                                 │
  │◄──────── JWT token ─────────────│
  │                                 │
  │── ConnectRPC request ──────────►│  Authorization: Bearer <token>
  │                                 │
  │                                 │  parse JWT → extract kid from header
  │                                 │  look up public key by kid
  │                                 │  verify signature, expiry, audience, issuer
  │                                 │  check JTI is not revoked
  │◄────── response ────────────────│
```

## JWT claims

Every token contains:

| Field               | Description                                                          |
| ------------------- | -------------------------------------------------------------------- |
| `sub`               | User ID                                                              |
| `jti`               | JWT ID — links the token to a revocable `JTI` record in the database |
| `iss` / `aud`       | Both set to `mechanus`                                               |
| `iat`, `nbf`, `exp` | Issued at, not-before (−1 min leeway), expires at (+1 hour)          |
| `user.id`           | User ID (redundant with `sub`, kept for convenience)                 |
| `user.name`         | Username                                                             |
| `user.roles`        | Array of role strings                                                |
| `user.campaigns`    | Array of campaign IDs the user has access to                         |
| `scope`             | Token scope, e.g. `"password"`                                       |

## RSA key management

Keys are generated as RSA key pairs and persisted via `infrastructure/storage`. The `KeyManager` loads all keys from storage on startup (`AfterInitialize` lifecycle hook) and caches them in memory.

Each key has a `kid` (key ID) stored in the JWT header. On validation the server looks up the public key by `kid`, which allows key rotation: old tokens signed with a previous key remain valid as long as that key is still in storage.

`KeyManager.GetSigningKey` returns the first available private key, or generates a new one if none exist.

## JTI (JWT ID) revocation

Every token is tied to a `JTI` database record (`infrastructure/persistence/models.JTI`). A JTI record has:

- `ID` — UUID, used as the JWT `jti` claim
- `UserID` — the user it belongs to
- `Revoked` — boolean flag

On validation, after the signature and expiry checks pass, the server fetches the JTI record and rejects the token if `Revoked = true`. This enables instant token invalidation (e.g. on logout or password change) without waiting for natural expiry.

`JTIService.GetActiveOrCreate` reuses an existing active JTI for a user rather than creating a new one each login, so all active sessions share a JTI and a single revoke call signs out all of them.

## Device authentication

Devices (e.g. a TV running the screen client) authenticate with an API key rather than a user JWT. The device ID is passed as:

- `X-Device-ID` HTTP header, or
- `deviceid` query parameter

on the WebSocket upgrade request. The server uses this to identify and assign the device to a screen.

## Token validation summary

1. Parse JWT, extract `kid` from header
2. Look up RSA public key by `kid`
3. Verify signature with RS512
4. Check `exp`, `iss`, `aud` (5-minute leeway on expiry)
5. Fetch `JTI` record; confirm `UserID` matches and `Revoked = false`
