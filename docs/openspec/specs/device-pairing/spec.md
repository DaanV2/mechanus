# device-pairing Specification

## Purpose

Lets an admin register a screen device (TV, tablet, table) with a short pairing code, so it gets its own revocable credentials and can be bound to a scene target.

## Requirements

### Requirement: Device pairing code
The system SHALL let an unauthenticated device request a 4-digit pairing code over a stream, keep the code valid only while the stream is open and for at most 5 minutes, and renew it automatically.

#### Scenario: Device shows a code
- **WHEN** an unpaired device opens `/devices/pair`
- **THEN** a 4-digit code is shown full screen and renewed before it expires

#### Scenario: Stream closed
- **WHEN** the device closes the pairing page
- **THEN** its code can no longer be confirmed

### Requirement: Admin confirms pairing
The system SHALL let an admin confirm a pairing code with a device name, after which the device receives a device token and is registered with a stable device ID.

#### Scenario: Correct code
- **WHEN** an admin enters the code shown on a device and a name on `/devices`
- **THEN** the device stores its device token and opens its screen view

### Requirement: Pairing attempts are rate limited
The system SHALL, after 5 wrong codes within a minute, replace every open pairing code and block that admin from confirming codes for the rest of the minute.

#### Scenario: Too many wrong codes
- **WHEN** an admin enters a wrong code for the fifth time within a minute
- **THEN** all open codes change and further confirmations by that admin are refused until the minute ends

### Requirement: Device tokens
The system SHALL issue JTI-backed device tokens, sent as `Authorization: Device <token>`, with an 8-hour access token and a 30-day refresh token; a device unused for 30 days SHALL have to pair again.

#### Scenario: Device refreshes
- **WHEN** a paired device's access token expires and its refresh token is valid
- **THEN** the device obtains new tokens without pairing again

#### Scenario: Long unused device
- **WHEN** a device has not refreshed for 30 days
- **THEN** it is sent back to the pairing page

### Requirement: Device revocation
The system SHALL let an admin revoke a paired device, which invalidates its tokens, closes its screen connections and removes its target binding.

#### Scenario: Revoke
- **WHEN** an admin revokes a device
- **THEN** the device's screen disconnects, it returns to the pairing page, and it no longer receives instructions for its former target

### Requirement: Device management page
The system SHALL list paired and connected devices to admins on `/devices`.

#### Scenario: List devices
- **WHEN** an admin opens `/devices`
- **THEN** paired devices are listed with their name, binding and whether they are connected

### Requirement: Device binding to a target
The system SHALL let an admin bind a paired device to a room target. The binding SHALL be stored with the device registration, merged into the room when it is read, and applied from the next instruction of a scene.

#### Scenario: Bind a TV
- **WHEN** an admin binds a device to the `table` target
- **THEN** from the next scene instruction the device receives what the scene sends to `table`

#### Scenario: Room file untouched
- **WHEN** a device is bound
- **THEN** the authored room file is not modified

### Requirement: Device identity metadata
The system SHALL read a device's ID and type from `X-Device-ID` / `X-Device-Type`, falling back to the `deviceId` / `deviceType` query parameters. Types SHALL be `screen`, `table`, `tablet`, `phone`, `desktop` or `other`, and SHALL NOT be used for authorization.

#### Scenario: Unknown device type
- **WHEN** a device connects with `X-Device-Type: fridge`
- **THEN** its type is recorded as `other`

#### Scenario: Query fallback
- **WHEN** a device connects without headers but with `?deviceId=tv-1`
- **THEN** its device ID is `tv-1`
