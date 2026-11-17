# device-drivers Specification

## Purpose

Defines how scene instructions reach non-screen devices through integration drivers, how the GM sees whether devices are reachable, and the MQTT integration for custom devices.

## Requirements

### Requirement: Devices are instructed concurrently with a timeout
The system SHALL instruct all devices of a scene concurrently and SHALL abandon an instruction that exceeds its per-instruction timeout without blocking the others.

#### Scenario: One slow device
- **WHEN** one light bridge does not answer while a scene activates
- **THEN** the other devices are instructed on time and the slow one is reported as failed

### Requirement: Integration configuration
The system SHALL read global integration settings (credentials, bridge or broker addresses) from a config set per integration, and device-specific settings from the device's room configuration.

#### Scenario: Broker address
- **WHEN** `integrations.mqtt` sets a broker address
- **THEN** the MQTT driver connects to that broker

### Requirement: Device status for the GM
The system SHALL track a health / connection status per device and let the GM list it; screens SHALL be `online` while the device has its screen open.

#### Scenario: TV closes its screen
- **WHEN** a bound TV closes the device view
- **THEN** its status in the GM view's Devices table is no longer `online`

### Requirement: MQTT custom devices
The system SHALL publish a scene's opaque payload for an MQTT device, retained, to the device's configured topic (default `mechanus/devices/<device-id>/set`), wrapped as `{scene, variant, target, payload}`.

#### Scenario: Scene activates with an MQTT device
- **WHEN** a scene instructs a target bound to MQTT device `fog-machine`
- **THEN** a retained message is published to `mechanus/devices/fog-machine/set` with the scene, variant, target and payload

### Requirement: MQTT scene events and availability
The system SHALL publish scene events to `mechanus/events/scene` (not retained) and its availability to `mechanus/status` as `online` / `offline` (retained, with `offline` as last will).

#### Scenario: Server goes away
- **WHEN** the server disconnects from the broker unexpectedly
- **THEN** `mechanus/status` reads `offline`
