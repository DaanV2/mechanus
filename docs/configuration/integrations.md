# Integrations

- [Integrations](#integrations)
  - [Shared settings](#shared-settings)
  - [MQTT](#mqtt)
    - [Example](#example)
    - [What a device receives](#what-a-device-receives)

Integrations connect a scene to devices other than screens. They are configured via CLI flags, environment variables, or the [config file](./config-file.md), under `integrations`. How they work is described in [architecture/scenes.md](../architecture/scenes.md#driver-framework).

An integration without its settings is disabled. A room that uses a device of a disabled integration still works, the device is reported as failed each time a scene instructs it.

## Shared settings

| Flag                         | Env var                      | Default | Description                                                               |
| ---------------------------- | ---------------------------- | ------- | ------------------------------------------------------------------------- |
| `--integrations.apply-timeout` | `INTEGRATIONS_APPLY_TIMEOUT` | `5s`    | How long a device gets to process an instruction, `0` means no limit |

## MQTT

Drives custom devices, such as fog machines, rumble packs and Arduino props, through an MQTT broker you run yourself.

| Flag                           | Env var                      | Default    | Description                                                                                                                 |
| ------------------------------ | ---------------------------- | ---------- | --------------------------------------------------------------------------------------------------------------------------- |
| `--integrations.mqtt.broker`   | `INTEGRATIONS_MQTT_BROKER`   |            | Broker address (`tcp://`, `mqtt://`, `tls://`, `ssl://`, `mqtts://`, `ws://`, `wss://`). Empty disables the integration |
| `--integrations.mqtt.client-id` | `INTEGRATIONS_MQTT_CLIENT_ID` | `mechanus` | Client ID on the broker, has to be unique per broker                                                                        |
| `--integrations.mqtt.username` | `INTEGRATIONS_MQTT_USERNAME` |            | Username for the broker                                                                                                     |
| `--integrations.mqtt.password` | `INTEGRATIONS_MQTT_PASSWORD` |            | Password for the broker. Prefer the environment variable over the config file                                               |

The server starts even if the broker is down, and keeps trying to connect. An invalid broker address stops the server from starting.

### Example

`config.yaml`:

```yaml
integrations:
  apply-timeout: 5s
  mqtt:
    broker: tcp://192.168.1.10:1883
    username: mechanus
```

```bash
INTEGRATIONS_MQTT_PASSWORD=change-me
```

`rooms/room.dat` binds a target to the device. `topic` is optional and defaults to `mechanus/devices/<device-id>/set`:

```json
{
  "id": "room",
  "name": "Living room",
  "devices": {
    "fog-machine": {
      "id": "fog-machine",
      "name": "Fog machine",
      "kind": "custom",
      "integration": "mqtt",
      "config": { "topic": "props/fog-machine/set" }
    }
  },
  "bindings": { "fog": ["fog-machine"] }
}
```

`scenes/crypt.dat` holds what that target should do. The payload can be any JSON:

```json
{
  "id": "crypt",
  "instructions": {
    "custom": {
      "fog": { "payload": { "burst_seconds": 3, "fan": "low" } }
    }
  }
}
```

### What a device receives

Activating the `crypt` scene publishes this, retained, to `props/fog-machine/set`:

```json
{
  "scene": "crypt",
  "variant": "",
  "target": "fog",
  "payload": { "burst_seconds": 3, "fan": "low" }
}
```

`variant` is empty for the scene's base instructions. Because the message is retained, a device receives the current instruction again when it reconnects, so make sure acting on it twice is harmless.

Devices and other software can also subscribe to:

| Topic                   | Retained | Content                                                                                                                 |
| ----------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------- |
| `mechanus/events/scene` | no       | `{"scene": "crypt", "variant": "", "from": {"scene": "road", "variant": ""}}` on a scene or variant switch |
| `mechanus/status`       | yes      | `online` or `offline`. The broker publishes `offline` itself if the server disappears                                   |
