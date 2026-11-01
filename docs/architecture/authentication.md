# Authentication

Mechanus uses JWT (JSON Web Tokens) for user authentication. Tokens are signed with server-managed RSA-2048 keys using the RS512 algorithm (RSA + SHA-512). Devices (TVs, tablets, other screen clients) authenticate with a device token obtained by pairing — see "Device authentication" below; that path is designed but not implemented yet.

## JWT flow

```ascii
Client                                   Server
  │                                        │
  │── LoginService.Login ─────────────────►│  credentials (username + password)
  │                                        │  rate limit per username, verify password hash
  │                                        │  create a new session (JTI) for the user
  │                                        │  sign an access token (scope "*") and a
  │                                        │  refresh token (scope "refresh"), same JTI
  │◄──────── access + refresh token ───────│
  │                                        │
  │── ConnectRPC request ─────────────────►│  Authorization: Bearer <access token>
  │                                        │  parse JWT → kid → public key
  │                                        │  verify signature, exp / nbf (5 min leeway), aud, iss
  │                                        │  check the JTI belongs to the user and isn't revoked
  │◄────── response ───────────────────────│
  │                                        │
  │── LoginService.Refresh ───────────────►│  refresh token (only scope "refresh" is accepted)
  │◄──────── new access + refresh token ───│  same JTI, user's current roles
  │                                        │
  │── LoginService.Logout ────────────────►│  revoke the session's JTI
```

### Client behaviour

The browser client (`client/src/lib/authentication/session.ts`) stores the access token in the `access-token` cookie (`Bearer <jwt>`) and the refresh token in the `refresh-token` cookie. Both cookies get a `Max-Age` matching their token's expiry, and `Secure` when the page was served over HTTPS. Before every ConnectRPC request it refreshes the access token when it expires within a minute, and on an `Unauthenticated` answer it refreshes once and retries. A token whose `exp` passed doesn't count as logged in.

### Error codes

Handlers answer `Unauthenticated` when a request needs a user but has no valid access token (none, malformed, expired, revoked or refresh-only) — the client should refresh or log in again — and `PermissionDenied` when the user is known but lacks the role (`grpc.requireRole`). The `Authenticator` interceptor itself never rejects a request: it only puts the claims, and whether they are valid, on the context.

## JWT claims

Every token contains:

| Field               | Description                                                                                   |
| ------------------- | --------------------------------------------------------------------------------------------- |
| `sub`               | User ID                                                                                       |
| `jti`               | JWT ID — links the token to a revocable `JTI` (session) record in the database                 |
| `iss` / `aud`       | Both set to `mechanus`                                                                        |
| `iat`, `nbf`, `exp` | Issued at, not-before (−1 min), expires at (`auth.access-token-duration` / `auth.refresh-token-duration`) |
| `user.id`           | User ID (redundant with `sub`, kept for convenience)                                          |
| `user.name`         | Username                                                                                      |
| `user.roles`        | Array of role strings — role changes reach the token on the next refresh                      |
| `user.campaigns`    | Array of campaign IDs the user has access to                                                  |
| `grant`             | How the token was obtained: `"password"` (login) or `"refresh"`                               |
| `scope`             | What the token may be used for: `"*"` (everything) or `"refresh"` (only `Refresh` / `Logout`) |

## Configuration

| Flag / env                                                     | Default | Description                                                              |
| -------------------------------------------------------------- | ------- | ------------------------------------------------------------------------ |
| `auth.access-token-duration` / `AUTH_ACCESS_TOKEN_DURATION`    | `1h`    | Lifetime of an access token                                              |
| `auth.refresh-token-duration` / `AUTH_REFRESH_TOKEN_DURATION`  | `168h`  | Lifetime of a refresh token; a session not refreshed in this time ends   |
| `auth.key-rotation` / `AUTH_KEY_ROTATION`                      | `720h`  | How often a new signing key is generated, `0` disables rotation          |
| `auth.signup-enabled` / `AUTH_SIGNUP_ENABLED`                  | `false` | Let anyone call `UserService.Create`; otherwise only admins can          |
| `auth.invite-duration` / `AUTH_INVITE_DURATION`                  | `168h`  | How long an invite link stays valid when the admin picks no duration, at most 30 days |
| `server.tls.cert` + `server.tls.key` / `SERVER_TLS_CERT` + `SERVER_TLS_KEY` | — | Serve HTTPS with these PEM files                              |
| `server.trust-proxy-headers` / `SERVER_TRUST_PROXY_HEADERS`   | `false` | Apply `X-Forwarded-Proto` / `-Host` / `-For` of a reverse proxy in front of the server, see "Reverse proxy" below |

## Accounts

- `UserService.Create` requires the admin role unless `auth.signup-enabled` is on (the dev / test compose files turn it on).
- Admins create accounts and change roles on the users page (`/users`).
- `InviteService` (admin) hands out single use invite links (`/users/signup?invite=<id>`, also shown as a QR code on `/users/invites`). The invite ID is a random UUID and the secret part of the link; an invite expires after the chosen duration (default `auth.invite-duration`, capped at 30 days) and can be revoked. `InviteService.Check` and `InviteService.Accept` need no login: accepting creates an account with the `user` role and uses up the invite. The invite is claimed before the account is created, so concurrent accepts can't both succeed; when the account can't be created (username taken) the invite stays usable. Unknown, expired, used and revoked invites all give the same answer.
- `UserService.ResetPassword` (admin) revokes every session of the user.
- `UserService.ChangePassword` (the user, needs the current password) revokes every other session of the user, keeping the one that made the call.
- `UserService.SetRoles` (admin) replaces a user's roles; an admin can't remove their own admin role.
- `LoginService.Login` locks a username out for 15 minutes after 5 failed attempts (in memory, per server), and compares against a dummy hash when the username doesn't exist, so an unknown username takes as long as a wrong password.

## RSA key management

Keys are generated as RSA key pairs and persisted via `infrastructure/storage`, with their creation time in the PEM headers. The `KeyManager` loads all keys from storage on startup (`AfterInitialize` lifecycle hook) and caches them in memory.

Each key has a `kid` (key ID) stored in the JWT header. On validation the server looks up the public key by `kid`, so tokens signed with an older key stay valid as long as that key is still in storage.

`KeyManager.GetSigningKey` signs with the newest key (ties broken on the `kid`), generating one when there is none or when the newest is older than `auth.key-rotation`. A key is deleted once it stopped signing longer ago than the longest token lifetime, so every token it signed has expired. Keys stored before creation times were tracked start their rotation period at the next startup.

## JTI (session) revocation

Every token is tied to a `JTI` database record (`infrastructure/persistence/models.JTI`), one per login session. A JTI record has:

- `ID` — UUID, used as the JWT `jti` claim
- `UserID` — the user it belongs to
- `Revoked` — boolean flag
- `UpdatedAt` — touched on every refresh

On validation, after the signature and time checks pass, the server fetches the JTI record and rejects the token if it belongs to another user or `Revoked = true`. Revoking a JTI signs out that one session (both its access and refresh tokens) right away; `JTIService.RevokeUser` signs out all sessions of a user.

Long-lived screen WebSocket connections of a user are closed (status 1008) when their access token expires, and when a periodic check (`server.websocket.session-check-interval`, default 1 minute) finds the session revoked. The browser client refreshes its access token before every (re)connect (`WebsocketHandler.open`), so a screen reconnects as the same user after such a close. A refresh that fails because of the network (or because the page was left) keeps the tokens; only a refresh the server refuses logs the user out.

On startup `JWTService.Cleanup` deletes revoked JTIs and JTIs not refreshed within the longest token lifetime.

## Device authentication

> **Status**: design only. Today `infrastructure/transport/websocket.WebsocketHandler.authenticate` accepts any connection without a user JWT (from the `Authorization` header, the `mechanus.bearer.<jwt>` subprotocol or the `access-token` cookie) as an unauthenticated device with the `Device` role, without checking a device token: it may only open the screen named after the device ID it reports (see the `// TODO: check token for device and device id.` in that function). Everything below describes the intended behaviour; [roadmap M1](../roadmap.md#m1--first-scene-on-the-table) tracks the work.

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

No identity on a route that needs one → `401 Unauthorized`; identity present but not allowed → `403 Forbidden`. The per-screen access check (`screens.ConnectionInfo.CanAccess`) runs after the route policy. Until screens have owners it lets users open every screen and a device only its own screen (screen ID = device ID); anything else gets `403 Forbidden`.

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
3. Verify signature, only RS512 is accepted
4. Check `exp` (required), `nbf`, `iat`, `iss`, `aud`, with a 5-minute leeway on the time claims (`JWT_LEEWAY`)
5. Fetch `JTI` record; confirm `UserID` matches and `Revoked = false`
6. The caller checks the `scope`: the API and WebSocket only accept `"*"`, `Refresh` only `"refresh"`

## Reverse proxy

When Mechanus runs behind a reverse proxy that terminates TLS, enable `server.trust-proxy-headers`. The server then takes the scheme from `X-Forwarded-Proto`, the host from `X-Forwarded-Host` (the WebSocket upgrade compares the `Origin` against it) and the client address from the first `X-Forwarded-For` entry, for logs and traces. Only enable it when every request goes through the proxy, otherwise clients can spoof these headers. Cookies get `Secure` from the browser's own `location.protocol`, so they are already right behind a proxy.
