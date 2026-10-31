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

| Client type | Credential |
| ----------- | ---------- |
| Player / GM | `Authorization: Bearer <jwt>` header on the upgrade request |
| Device      | **Not implemented yet.** Any connection without a `Bearer` JWT is currently accepted as an unauthenticated device — see [architecture/authentication.md](../architecture/authentication.md#device-authentication). |

## Headers

| Header            | Description                                                                           |
| ----------------- | --------------------------------------------------------------------------------------- |
| `Authorization`   | `Bearer <jwt>` for player/GM connections — the only header currently read by the server  |
| `X-Connection-ID` | Planned: a UUID for tracking/telemetry per connection — not yet read by the server       |
| `X-Device-ID`     | Planned: device identifier for hardware screen clients — not yet read by the server      |
