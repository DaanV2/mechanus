# Authentication

Mechanus uses JWT (JSON Web Tokens) for user authentication. Tokens are signed with server-managed RSA-2048 keys using the RS512 algorithm (RSA + SHA-512). Devices (TVs, tablets, other screen clients) authenticate with a device token obtained by pairing — see "Device authentication" below; that path is designed but not implemented yet.

## JWT flow

```ascii
Client                            Server
  │                                 │
  │── POST /users.v1.LoginService ─►│  credentials (username + password)
  │                                 │
  │                                 │  verify password hash
  │                                 │  get or create JTI (JWT ID) for user
  │                                 │  sign JWT with RSA-2048 private key (RS512)
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
| `grant`             | How the token was obtained: `"password"` (login) or `"refresh"`      |
| `scope`             | What the token may be used for: `"*"` (everything) or `"refresh"`    |

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

> **Status**: design only. Today `infrastructure/transport/websocket.WebsocketHandler.authenticate` accepts any connection without a user JWT (from the `Authorization` header, the `mechanus.bearer.<jwt>` subprotocol or the `access-token` cookie) as an unauthenticated device with the `Device` role, without reading a device ID, device type or token (see the `// TODO: check token for device and device id.` in that function). Everything below describes the intended behaviour; [roadmap M1](../roadmap.md#m1--first-scene-on-the-table) tracks the work.

### Optional authentication, enforced per route

Authentication on the WebSocket (and on device-facing ConnectRPC routes) is **optional at connect time and enforced per route**. Every connection is first resolved to an identity, then the route's policy decides whether that identity may proceed:

| Identity         | Resolved from                                              | Roles            |
| ---------------- | ---------------------------------------------------------- | ---------------- |
| User             | `Authorization: Bearer <jwt>` (user JWT, `scope: "*"`); browsers use the `mechanus.bearer.<jwt>` subprotocol or the `access-token` cookie, see [communication.md](./communication.md#browser-credentials) | the user's roles |
| Paired device    | `Authorization: Device <device token>` (issued by pairing) | `device`         |
| Anonymous device | no credential, only the device identity headers            | none             |

A credential that is present but invalid (bad signature, expired, revoked) is **rejected** with `401`, never downgraded to anonymous — otherwise the client cannot tell "log in / pair again" from "not allowed".

Each route registers a policy alongside its handler:

| Policy        | Accepts                                   | Example                                    |
| ------------- | ----------------------------------------- | ------------------------------------------ |
| `anonymous`   | any identity, including anonymous devices | device pairing                             |
| `device`      | paired devices and users                  | screen WebSocket for room devices          |
| `user`        | users only                                | GM / player screens                        |
| `role:<role>` | users holding that role                   | admin actions such as confirming a pairing |

No identity on a route that needs one → `401 Unauthorized`; identity present but not allowed → `403 Forbidden`. The per-screen access check (`// TODO: check that screenHandler is allowed to be access by this requester`) runs after the route policy.

### Device identity headers

Every device connection identifies itself, authenticated or not:

| Header          | Query fallback | Description                                                                                   |
| --------------- | -------------- | --------------------------------------------------------------------------------------------- |
| `X-Device-ID`   | `deviceId`     | Stable device identifier, generated by the device on first start and persisted locally (UUID) |
| `X-Device-Type` | `deviceType`   | Kind of client: `screen`, `table`, `tablet`, `phone`, `desktop`, `other`; unknown values are stored as `other`          |

Browsers cannot set custom headers on `new WebSocket(...)`, so browser clients send the query parameters instead; the server reads the header first and falls back to the query parameter. For a paired device the device ID inside the token is authoritative: a mismatching `X-Device-ID` is rejected.

The device ID and type are kept in the connection's session info (`screens.ConnectionInfo`, next to the user ID, roles and connection type), added to the connection's logger / trace attributes, and used to list connected devices on the devices page. `X-Device-Type` is client-reported metadata, separate from `devices.DeviceType` (which only says whether a connection is a user or a device).

### Pairing

A new device has no credential. Pairing turns it into a paired device with a short code the admin types in:

```ascii
Device (/devices/pair)              Server                          Admin (/devices)
  │                                   │                                    │
  │── start pairing (anonymous) ─────►│  X-Device-ID, X-Device-Type        │
  │                                   │  generate 4-digit code             │
  │◄──────── code "4821" ─────────────│  (single use, short expiry)        │
  │                                   │                                    │
  │  shows  4 8 2 1  full screen      │                                    │
  │                                   │◄── confirm "4821" + name ──────────│  admin role
  │                                   │  look up device ID for the code    │
  │                                   │  register device, issue token      │
  │◄──────── device token ────────────│──────────── paired ───────────────►│
  │                                   │                                    │
  │  store token, reconnect with      │                                    │
  │  Authorization: Device <token>    │                                    │
```

- The code is 4 digits, single use, bound to the device ID that requested it, and only valid while the device's pairing stream is open, for at most 5 minutes; the server pushes a new code over the stream when it expires.
- With only 10 000 possible codes, confirm attempts are rate limited and a code is invalidated after a few wrong guesses.
- Confirming requires an admin (later possibly a GM); the admin names the device, and that name is shown on the devices page and when binding the device to a room target.
- Transport: the device side is a ConnectRPC server stream (`DeviceService.Pair`, `anonymous` route) that streams codes and ends with the device tokens.
- Tokens: pairing issues a device access token (JWT with `sub` = device ID, backed by a JTI, 8 h) and a device refresh token (30 days); the device refreshes on its own. Revoking the JTI on the devices page un-pairs the device, which drops back to the pairing page; a device unused for more than 30 days has to pair again.

## Token validation summary

1. Parse JWT, extract `kid` from header
2. Look up RSA public key by `kid`
3. Verify signature with RS512
4. Check `exp`, `iss`, `aud` (5-minute leeway on expiry)
5. Fetch `JTI` record; confirm `UserID` matches and `Revoked = false`
