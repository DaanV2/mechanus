# screen-connection Specification

## Purpose

Defines how screens connect to the server over the screen WebSocket, who may open which screen, and how the connection survives token expiry.

## Requirements

### Requirement: Screen WebSocket authentication
The system SHALL accept credentials on the screen WebSocket as an `Authorization` header, a `mechanus.bearer.<jwt>` subprotocol or the auth cookie, and SHALL reject invalid credentials instead of treating them as anonymous.

#### Scenario: Browser connects with cookie
- **WHEN** a signed-in browser opens the screen WebSocket on the same origin
- **THEN** the connection is authenticated as that user

#### Scenario: Invalid token
- **WHEN** a client connects with an expired or malformed token
- **THEN** the upgrade is rejected with `401`

### Requirement: Per-route auth policies
The system SHALL enforce an auth policy per route (`anonymous`, `device`, `user` or `role:<role>`); authentication is optional at connect and enforced by the route.

#### Scenario: Screen without credentials
- **WHEN** a client connects to the screen WebSocket without any credential
- **THEN** the connection is refused

### Requirement: Screen access check
The system SHALL let users open every screen and a paired device only its own screen, answering `403` when access is denied after the route policy passed.

#### Scenario: Device opens another screen
- **WHEN** a paired device requests the screen of a different device ID
- **THEN** the system responds `403`

### Requirement: Connections end with their credentials
The system SHALL close a screen connection when its token expires or its session or device is revoked.

#### Scenario: Token expires
- **WHEN** the token used to open a screen expires
- **THEN** the server closes the connection

### Requirement: Client reconnects
The client SHALL reconnect the screen WebSocket automatically, using a refreshed token when the server closed the connection on expiry.

#### Scenario: Reconnect after expiry
- **WHEN** the server closes a screen connection because its token expired
- **THEN** the client refreshes its token and reconnects without user action

### Requirement: Connections are routed by device ID
The system SHALL key screen connections by device ID, so instructions for a device reach every open connection of that device, and SHALL record the device ID and type in logs and in the connection's trace span.

#### Scenario: Instruction for a device
- **WHEN** the screen driver pushes state for device `tv-1`
- **THEN** every open screen connection of `tv-1` receives it

### Requirement: Initial setup and ping
The system SHALL answer `Ping` and `InitialSetupRequest` on the screen WebSocket, replying to the initial setup with the device's current screen state.

#### Scenario: Screen connects during a scene
- **WHEN** a bound device connects while a scene is active and sends `InitialSetupRequest`
- **THEN** it receives the current screen state for its target
