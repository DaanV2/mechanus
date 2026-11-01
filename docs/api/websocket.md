# Websockets

The screen WebSocket is the continuous, bidirectional channel each client (GM, player, device/TV) uses to receive server-pushed screen state. For the full protocol — message framing, supported client messages, and the server-side flow — see [architecture/communication.md](../architecture/communication.md#websocket--screen-protocol).

## Path

```text
<ws|wss>://<host>[:<port>]/api/v1/screen/{screenId}/{id}
```

## screenId

The ID of a screen already registered with the server's `ScreenManager`. Routing by device ID instead of a concrete screen ID is not implemented yet — see [scenes](../architecture/scenes.md#screens--views).

## id

The connecting user ID or device ID.

## Authentication

Optional when connecting, enforced per route (`anonymous`, `device`, `user`, `role:<role>`). A credential that is present but invalid is rejected with `401`. See [architecture/authentication.md](../architecture/authentication.md#device-authentication).

| Client type      | Credential                                                         |
| ---------------- | ------------------------------------------------------------------ |
| Player / GM      | `Authorization: Bearer <jwt>`                                      |
| Paired device    | `Authorization: Device <token>`, token issued by pairing — planned |
| Anonymous device | None; only allowed on `anonymous` routes (pairing) — planned       |

Today only `Bearer` is implemented; any other connection is accepted as an unauthenticated device.

## Headers

Browsers cannot set headers on `new WebSocket(...)`, so browser clients use the query parameter fallback; the header wins when both are present.

| Header            | Query fallback | Description                                                                        |
| ----------------- | -------------- | ---------------------------------------------------------------------------------- |
| `Authorization`   | TBD            | `Bearer <jwt>` or `Device <token>` — only `Bearer` is read today                   |
| `X-Device-ID`     | `deviceId`     | Planned: stable device ID, stored in the connection's session info                 |
| `X-Device-Type`   | `deviceType`   | Planned: kind of client (`screen`, `tablet`, ...), stored in the session info      |
| `X-Connection-ID` | —              | Planned: a UUID for tracking/telemetry per connection — not yet read by the server |
