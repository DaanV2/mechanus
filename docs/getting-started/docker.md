# Run the Docker image

Part of [Getting started](./README.md). Prefer no Docker? See [Run a release binary](./binaries.md).

Every release publishes a multi-arch image for `linux/amd64` and `linux/arm64`:

| Tag | Points to |
| --- | --- |
| `ghcr.io/daanv2/mechanus:latest` | The newest release |
| `ghcr.io/daanv2/mechanus:v0.4.1` | That exact release (see the [releases](https://github.com/DaanV2/mechanus/releases) for the list) |

Pin a version tag for a setup you don't want to change under you; `latest` is fine to try things out. Docker picks the right architecture by itself. A Raspberry Pi needs a 64-bit OS (arm64); there is no 32-bit ARM image.

- [Install Docker](#install-docker)
- [Start the container](#start-the-container)
- [Docker Compose](#docker-compose)
- [What lives where](#what-lives-where)
- [Scenes, rooms and config](#scenes-rooms-and-config)
- [Settings](#settings)
- [LAN discovery (mDNS)](#lan-discovery-mdns)
- [Update](#update)
- [Back up](#back-up)
- [Troubleshooting](#troubleshooting)

## Install Docker

| OS | Install |
| --- | --- |
| Linux | [Docker Engine](https://docs.docker.com/engine/install/) for your distribution, or `curl -fsSL https://get.docker.com \| sh`. Add yourself to the `docker` group (`sudo usermod -aG docker $USER`, then log out and in) to run `docker` without `sudo`. |
| Windows | [Docker Desktop for Windows](https://docs.docker.com/desktop/setup/install/windows-install/) with the WSL 2 backend. |
| macOS | [Docker Desktop for Mac](https://docs.docker.com/desktop/setup/install/mac-install/), the Apple silicon or Intel build to match your Mac. |

On Windows and macOS, Docker Desktop runs Linux containers in a small VM; it has to be running before the commands below work.

## Start the container

**Linux / macOS** (bash, zsh):

```bash
docker run -d --name mechanus \
  -p 8080:8080 \
  -v mechanus-data:/app/.local \
  -e INITIALIZE_ADMIN_USERNAME=admin \
  -e INITIALIZE_ADMIN_PASSWORD='change-me' \
  --restart unless-stopped \
  ghcr.io/daanv2/mechanus:latest
```

**Windows** (PowerShell):

```powershell
docker run -d --name mechanus `
  -p 8080:8080 `
  -v mechanus-data:/app/.local `
  -e INITIALIZE_ADMIN_USERNAME=admin `
  -e INITIALIZE_ADMIN_PASSWORD='change-me' `
  --restart unless-stopped `
  ghcr.io/daanv2/mechanus:latest
```

In `cmd.exe`, put it on one line or end each line with `^` instead of `` ` ``.

What each part does:

- `-p 8080:8080` publishes the web app on port 8080 of the host. Use `-p 80:8080` to serve it on the default HTTP port instead.
- `-v mechanus-data:/app/.local` keeps the SQLite database and screen state in a named volume, so it survives updates. Without it, everything is lost when the container is removed.
- `INITIALIZE_ADMIN_*` creates the first admin account when the database has none. Leave them out and the server creates `admin` with a random password, printed once in `docker logs mechanus`. Once the account exists these variables are ignored.
- `--restart unless-stopped` starts the container again after a reboot (as long as Docker itself starts on boot, the default on Linux; enable _Start Docker Desktop when you sign in_ on Windows/macOS).

Then open `http://localhost:8080` and continue with [Getting started → Log in](./README.md#2-log-in).

Day-to-day commands:

```bash
docker logs -f mechanus   # follow the log
docker stop mechanus      # stop
docker start mechanus     # start again
```

## Docker Compose

The same setup as a `compose.yaml`, with a `config` folder next to it for your scenes (see [below](#scenes-rooms-and-config)):

```yaml
services:
  mechanus:
    image: ghcr.io/daanv2/mechanus:latest
    container_name: mechanus
    ports:
      - "8080:8080"
    environment:
      INITIALIZE_ADMIN_USERNAME: admin
      INITIALIZE_ADMIN_PASSWORD: change-me
    volumes:
      - mechanus-data:/app/.local
      - ./config:/app/.config
    restart: unless-stopped

volumes:
  mechanus-data:
```

```bash
docker compose up -d
```

Compose works the same on all three OSes; relative paths like `./config` are resolved against the folder of `compose.yaml`.

## What lives where

Inside the container the server runs from `/app`:

| Path | Holds | Mount it? |
| --- | --- | --- |
| `/app/.local` | State: the SQLite database (`db.sqlite`) and screen state | Yes, as a volume, or you lose your data on every update |
| `/app/.config` | Configuration you author: `scenes/`, `rooms/`, `shared/` and an optional `config.yaml` | Yes, as a bind mount to a folder you edit |
| `/web` | The web app, part of the image | No |

A `.local` or `.config` folder in the working directory takes precedence over the OS defaults, which is why mounting them at `/app` is enough. See [config file](../configuration/config-file.md) for the lookup order.

## Scenes, rooms and config

Make a folder on the host with this layout (all parts optional):

```text
config/
├── config.yaml     settings, instead of environment variables
├── rooms/          rooms: devices and target bindings
├── scenes/         scene configs (.dat) and their assets
└── shared/         assets shared between scenes
```

and mount it at `/app/.config`:

```bash
# Linux / macOS
-v "$(pwd)/config:/app/.config"
```

```powershell
# Windows (PowerShell)
-v "${PWD}\config:/app/.config"
```

The repository has a working example in [`docs/examples/battlemap`](../examples/battlemap); copy its `scenes`, `rooms` and `shared` folders into your `config` folder to start from it. See [battle maps](../configuration/battle-maps.md) for the file formats.

On Windows and macOS, keep the folder somewhere Docker Desktop is allowed to share (your user folder is by default).

## Settings

Every server flag has an environment variable: upper-case it and replace `.` and `-` with `_`. `--server.port` is `SERVER_PORT`, `--auth.signup-enabled` is `AUTH_SIGNUP_ENABLED`. Pass them with `-e NAME=value` or under `environment:` in Compose. `docker run --rm ghcr.io/daanv2/mechanus:latest /app/mechanus server --help` lists them all with their defaults.

The image already sets:

| Variable | Value | |
| --- | --- | --- |
| `SERVER_PORT` | `8080` | Port inside the container; change the `-p` host side instead |
| `SERVER_STATIC_FOLDER` | `/web` | Where the web app is |
| `LOG_FORMAT` | `json` | Set to `text` for easier reading in `docker logs` |

Commonly changed:

- `DATABASE_TYPE` / `DATABASE_DSN`: PostgreSQL or MySQL instead of SQLite, see [database](../configuration/database.md).
- `SERVER_TLS_CERT` / `SERVER_TLS_KEY`: serve HTTPS; mount the certificate files into the container and point these at them.
- `INTEGRATIONS_MQTT_BROKER`: enable MQTT devices, see [integrations](../configuration/integrations.md).
- `LOG_LEVEL`: `debug`, `info` (default), `warn`, `error`.

## LAN discovery (mDNS)

The server announces itself on the LAN over mDNS. With the default bridge network those announcements stay inside Docker. On a **Linux** host, use host networking to let them through; `-p` is then not needed, as the server binds directly to port 8080 of the host:

```bash
docker run -d --name mechanus --network host \
  -v mechanus-data:/app/.local \
  --restart unless-stopped \
  ghcr.io/daanv2/mechanus:latest
```

In Compose: `network_mode: host` instead of `ports:`.

On **Windows and macOS**, Docker Desktop's VM sits between the container and the LAN, so mDNS won't reach other devices even with host networking. Everything else works; devices just connect by IP. Use the [release binary](./binaries.md) if you need discovery on those OSes.

## Update

Your data is in the volume, so updating is replacing the container:

```bash
docker pull ghcr.io/daanv2/mechanus:latest
docker stop mechanus
docker rm mechanus
# then the same docker run command as before
```

With Compose:

```bash
docker compose pull
docker compose up -d
```

If you pinned a version tag, change it to the new version first. Database migrations run on start. Read the release notes before skipping several versions.

## Back up

Stop the container, then copy the volume to an archive in the current folder:

```bash
docker stop mechanus
docker run --rm -v mechanus-data:/data -v "$(pwd):/backup" busybox tar czf /backup/mechanus-data.tar.gz -C /data .
docker start mechanus
```

On Windows PowerShell, use `-v "${PWD}:/backup"`. Your `config` folder is a normal folder on the host; back it up like any other.

## Troubleshooting

- **The page doesn't load from other devices**: check the host firewall allows the published port, see [Getting started → Connect the other devices](./README.md#3-connect-the-other-devices).
- **`port is already allocated`**: something else uses 8080 on the host. Publish another port, e.g. `-p 8081:8080`, and open `http://<host>:8081`.
- **Lost the generated admin password**: it is only printed on the start that created the account. Search the log with `docker logs mechanus 2>&1 | grep -A2 "admin account"`; if the container was recreated since, the log is gone.
- **Scenes don't show up**: check the mount with `docker exec mechanus ls /app/.config/scenes`. An empty listing means the host path is wrong (or, on Docker Desktop, not shared).
- **`no matching manifest for linux/arm/v7`**: the host runs a 32-bit OS (common on older Raspberry Pi installs). Install the 64-bit version of the OS.
