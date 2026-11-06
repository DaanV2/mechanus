# Installation

Mechanus is one server that runs on a machine in your home network (a Raspberry Pi, NAS, mini PC or laptop). Every other device — the table TV, player phones, the GM's laptop — only needs a browser.

Each [GitHub release](https://github.com/DaanV2/mechanus/releases) comes as:

- a binary archive for Linux, Windows and macOS, each for amd64 (x86-64) and arm64
- a Docker image for linux/amd64 and linux/arm64: `ghcr.io/daanv2/mechanus:<version>` (or `:latest`)

Pick one of the two below.

## Release binary

1. Download the archive for your system from the release page:

   | System                                    | Archive                                  |
   | ----------------------------------------- | ---------------------------------------- |
   | Linux, x86-64 (most PCs, NAS)             | `mechanus_<version>_linux_amd64.tar.gz`  |
   | Linux, arm64 (Raspberry Pi 4/5 on 64-bit) | `mechanus_<version>_linux_arm64.tar.gz`  |
   | Windows, x86-64                           | `mechanus_<version>_windows_amd64.zip`   |
   | Windows, arm64                            | `mechanus_<version>_windows_arm64.zip`   |
   | macOS, Apple silicon                      | `mechanus_<version>_darwin_arm64.tar.gz` |
   | macOS, Intel                              | `mechanus_<version>_darwin_amd64.tar.gz` |

   `checksums.txt` on the same page holds the SHA-256 of every archive (`sha256sum -c checksums.txt --ignore-missing`).

2. Extract it. The archive holds the `mechanus` binary (`mechanus.exe` on Windows), the web app in `web/`, `LICENSE` and `README.md`.

3. From the extracted folder, start the server and point it at the web app:

   ```bash
   # Linux / macOS
   ./mechanus server --server.static.folder web \
     --initialize.admin.username admin --initialize.admin.password 'change-me'
   ```

   ```powershell
   # Windows (PowerShell)
   .\mechanus.exe server --server.static.folder web `
     --initialize.admin.username admin --initialize.admin.password 'change-me'
   ```

   The `--initialize.admin.*` flags create the first admin account when the database has none; leave them off on later starts. Without them, the first start creates an `admin` account with a random password and prints it once in the log. `--server.static.folder` is resolved against the current folder, so either start from the extracted folder or pass the absolute path to `web/`. Without it the server looks for `/web` (the Docker image's location) and serves no pages.

4. Open `http://<machine-ip>:8080` in a browser and log in.

On macOS, a downloaded binary is quarantined by Gatekeeper. Allow it with `xattr -d com.apple.quarantine ./mechanus`, or via _System Settings → Privacy & Security_.

Settings can also come from environment variables or a config file instead of flags (`SERVER_STATIC_FOLDER=web`, `server.static.folder: web`), see [config file](./configuration/config-file.md). `mechanus config paths` prints where Mechanus looks for its config and keeps its state (including the SQLite database by default).

## Docker

```bash
docker run -d --name mechanus \
  -p 8080:8080 \
  -v mechanus-data:/data \
  -e DATABASE_DSN=/data/mechanus.db \
  -e INITIALIZE_ADMIN_USERNAME=admin \
  -e INITIALIZE_ADMIN_PASSWORD='change-me' \
  --restart unless-stopped \
  ghcr.io/daanv2/mechanus:latest
```

The image serves the web app from `/web` already. The `mechanus-data` volume keeps the SQLite database across container updates; without it the database lives inside the container and is lost when the container is removed. Scenes and rooms are read from `/app/.config` — mount your own folders there, as `docker-compose.yaml` does with `docs/examples/battlemap`.

Or with Docker Compose:

```yaml
services:
  mechanus:
    image: ghcr.io/daanv2/mechanus:latest
    ports:
      - "8080:8080"
    environment:
      DATABASE_DSN: /data/mechanus.db
      INITIALIZE_ADMIN_USERNAME: admin
      INITIALIZE_ADMIN_PASSWORD: change-me
    volumes:
      - mechanus-data:/data
      - ./scenes:/app/.config/scenes:ro
      - ./rooms:/app/.config/rooms:ro
    restart: unless-stopped

volumes:
  mechanus-data:
```

The server announces itself on the LAN over mDNS. Inside Docker's default bridge network those announcements don't reach the LAN; on Linux use `network_mode: host` (or `--network host`) if you want discovery.

## Next steps

- [Database](./configuration/database.md): use Postgres or MySQL instead of SQLite
- [Battle maps](./configuration/battle-maps.md): put your first scene on the table
- [Integrations](./configuration/integrations.md): lights, music and other devices
