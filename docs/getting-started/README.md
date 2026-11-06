# Getting started

Mechanus is one server that runs on a machine in your home network: a Raspberry Pi, NAS, mini PC, desktop or laptop. Every other device at the table (the TV with the battle map, player phones, the GM's laptop) only needs a browser that can reach that machine.

This guide gets that server running from a [GitHub release](https://github.com/DaanV2/mechanus/releases), then walks you through the first login.

## 1. Pick how to run it

Every release ships the same server in two forms:

| | [Docker image](./docker.md) | [Release binary](./binaries.md) |
| --- | --- | --- |
| What you download | `ghcr.io/daanv2/mechanus:<version>` | `mechanus_<version>_<os>_<arch>` archive |
| Runs on | Linux (amd64, arm64); Windows and macOS through Docker Desktop | Linux, Windows and macOS, each on amd64 and arm64 |
| Needs | Docker | Nothing else |
| Updating | Pull the new image, recreate the container | Download the new archive, swap the files |
| Good fit for | A NAS or home server that already runs containers | A laptop or Raspberry Pi, or when you don't want Docker |

Not sure? Use the **binary**: it is one download and one command on every OS. Use **Docker** if the machine already runs other containers.

- **[Run the Docker image →](./docker.md)**
- **[Run a release binary →](./binaries.md)**

Both guides cover Linux, Windows and macOS. Come back here once the server is running.

## 2. Log in

1. On the server machine, open `http://localhost:8080`.
2. Log in with the admin account:
   - If you passed an admin username and password on the first start (`--initialize.admin.*` flags or `INITIALIZE_ADMIN_*` variables), use those.
   - Otherwise the first start created an `admin` account with a random password and printed it **once** in the log. Look for it in the terminal output, or in `docker logs mechanus`.
3. Change that password from your account page if it was generated or is a placeholder like `change-me`.

The admin account only has to be created once; the account lives in the database, so later starts don't need the admin flags.

## 3. Connect the other devices

Other devices reach the server on the machine's LAN address, for example `http://192.168.1.20:8080`. To find that address:

| OS | Command |
| --- | --- |
| Linux | `hostname -I` |
| macOS | `ipconfig getifaddr en0` (Wi-Fi) or `en1` |
| Windows | `ipconfig`, the _IPv4 Address_ of your network adapter |

If a phone or TV can't load the page:

- Check that both are on the same network (guest Wi-Fi is often isolated).
- Allow port `8080` through the firewall of the server machine. Windows asks on the first start of the binary; pick _Private networks_. On Linux with `ufw`: `sudo ufw allow 8080/tcp`.
- Give the server machine a fixed IP (a DHCP reservation in your router), so the address doesn't change between sessions.

The server also announces itself on the LAN over mDNS, so apps that browse for services can find it without the IP. For Docker that needs host networking, see [Docker → LAN discovery](./docker.md#lan-discovery-mdns).

## 4. Next steps

- [Battle maps](../configuration/battle-maps.md): put your first scene on the table
- [Integrations](../configuration/integrations.md): lights, music and other devices
- [Database](../configuration/database.md): use PostgreSQL or MySQL instead of the built-in SQLite
- [Config file](../configuration/config-file.md): keep settings in a `config.yaml` instead of flags
- [OpenTelemetry](../configuration/opentelemetry.md): traces and metrics
