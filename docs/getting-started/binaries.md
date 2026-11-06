# Run a release binary

Part of [Getting started](./README.md). Prefer containers? See [Run the Docker image](./docker.md).

The binary is self-contained: no runtime, database server or Docker needed. SQLite is built in.

- [Pick your archive](#pick-your-archive)
- [Download and verify](#download-and-verify)
- [Extract](#extract)
- [Start the server](#start-the-server)
- [Where your data lives](#where-your-data-lives)
- [Settings](#settings)
- [Run it in the background](#run-it-in-the-background)
- [Update](#update)
- [Uninstall](#uninstall)
- [Troubleshooting](#troubleshooting)

## Pick your archive

Each [release](https://github.com/DaanV2/mechanus/releases) has six archives, named `mechanus_<version>_<os>_<arch>`. The version has no `v` prefix: release `v0.4.1` has `mechanus_0.4.1_linux_amd64.tar.gz`.

| System | Archive |
| --- | --- |
| Linux, x86-64 (most PCs, NAS, mini PCs) | `mechanus_<version>_linux_amd64.tar.gz` |
| Linux, arm64 (Raspberry Pi 3/4/5 with a 64-bit OS) | `mechanus_<version>_linux_arm64.tar.gz` |
| Windows, x86-64 (Intel / AMD) | `mechanus_<version>_windows_amd64.zip` |
| Windows, arm64 (Snapdragon / Copilot+ PCs) | `mechanus_<version>_windows_arm64.zip` |
| macOS, Apple silicon (M1 and later) | `mechanus_<version>_darwin_arm64.tar.gz` |
| macOS, Intel | `mechanus_<version>_darwin_amd64.tar.gz` |

Not sure which architecture you have?

| OS | Command | Means |
| --- | --- | --- |
| Linux / macOS | `uname -m` | `x86_64` → `amd64`; `aarch64` or `arm64` → `arm64` |
| Windows (PowerShell) | `$env:PROCESSOR_ARCHITECTURE` | `AMD64` → `amd64`; `ARM64` → `arm64` |

A Raspberry Pi with a 32-bit OS (`uname -m` prints `armv7l`) isn't supported; install the 64-bit version of Raspberry Pi OS.

## Download and verify

Download the archive and `checksums.txt` from the release page in your browser, or from a terminal. The examples use `v0.4.1`; replace both version numbers with the release you want.

**Linux**

```bash
curl -LO https://github.com/DaanV2/mechanus/releases/download/v0.4.1/mechanus_0.4.1_linux_amd64.tar.gz
curl -LO https://github.com/DaanV2/mechanus/releases/download/v0.4.1/checksums.txt
sha256sum -c checksums.txt --ignore-missing
```

**macOS**

```bash
curl -LO https://github.com/DaanV2/mechanus/releases/download/v0.4.1/mechanus_0.4.1_darwin_arm64.tar.gz
curl -LO https://github.com/DaanV2/mechanus/releases/download/v0.4.1/checksums.txt
shasum -a 256 -c checksums.txt --ignore-missing
```

**Windows** (PowerShell)

```powershell
$v = "0.4.1"
Invoke-WebRequest "https://github.com/DaanV2/mechanus/releases/download/v$v/mechanus_${v}_windows_amd64.zip" -OutFile "mechanus_${v}_windows_amd64.zip"
Invoke-WebRequest "https://github.com/DaanV2/mechanus/releases/download/v$v/checksums.txt" -OutFile checksums.txt

# Compare the two hashes, they must match
(Get-FileHash "mechanus_${v}_windows_amd64.zip" -Algorithm SHA256).Hash.ToLower()
Select-String "windows_amd64" checksums.txt
```

The Linux and macOS check prints `<archive>: OK`. Anything else means the download is broken or changed; download it again and don't run it until it checks out.

## Extract

The archive has no top-level folder, so extract it into a folder of its own. You get:

```text
mechanus/
├── mechanus        the server (mechanus.exe on Windows)
├── web/            the web app the server hands to browsers
├── LICENSE
└── README.md
```

**Linux / macOS**

```bash
mkdir mechanus
tar -xzf mechanus_0.4.1_linux_amd64.tar.gz -C mechanus
cd mechanus
```

**Windows** (PowerShell)

```powershell
Unblock-File "mechanus_0.4.1_windows_amd64.zip"   # drop the "downloaded from the internet" mark
Expand-Archive "mechanus_0.4.1_windows_amd64.zip" -DestinationPath mechanus
cd mechanus
```

Or right-click the zip in Explorer → _Properties_ → tick _Unblock_ → _OK_, then _Extract All…_.

Where to put the folder is up to you, for example `~/mechanus` or `/opt/mechanus` on Linux, `~/Applications/mechanus` on macOS, or `C:\mechanus` on Windows.

## Start the server

From the extracted folder:

**Linux / macOS**

```bash
./mechanus server --server.static.folder web \
  --initialize.admin.username admin --initialize.admin.password 'change-me'
```

**Windows** (PowerShell)

```powershell
.\mechanus.exe server --server.static.folder web `
  --initialize.admin.username admin --initialize.admin.password 'change-me'
```

- `--server.static.folder web` points the server at the web app. The path is relative to the folder you start from; from anywhere else, pass the full path to `web`. Without it the server looks for `/web` (where the Docker image keeps it) and serves no pages.
- `--initialize.admin.*` creates the first admin account when the database has none. Leave them off on later starts. Without them, the first start creates `admin` with a random password and prints it **once** in the log.

The server runs in the foreground and logs to the terminal; stop it with `Ctrl+C`. Open `http://localhost:8080` and continue with [Getting started → Log in](./README.md#2-log-in).

### First start on macOS

A binary downloaded with a browser is quarantined by Gatekeeper, and macOS refuses to open it ("cannot be opened because the developer cannot be verified"). Clear the quarantine flag from the extracted folder:

```bash
xattr -dr com.apple.quarantine .
```

Or try to start it once, then allow it under _System Settings → Privacy & Security_ → _Open Anyway_. Archives downloaded with `curl` aren't quarantined. macOS may then ask whether `mechanus` may accept incoming network connections; allow it, or other devices can't connect.

### First start on Windows

Windows Defender Firewall asks whether `mechanus.exe` may communicate on networks. Allow _Private networks_ (your home network); leave _Public_ unticked. If you dismissed the prompt, allow it under _Windows Security → Firewall & network protection → Allow an app through firewall_.

If SmartScreen shows "Windows protected your PC", the zip wasn't unblocked before extracting: choose _More info → Run anyway_, or run `Unblock-File .\mechanus.exe`.

### First start on Linux

Nothing extra is needed. If the machine runs a firewall, open the port, e.g. `sudo ufw allow 8080/tcp` or `sudo firewall-cmd --add-port=8080/tcp --permanent && sudo firewall-cmd --reload`.

## Where your data lives

Mechanus keeps two kinds of files outside the extracted folder:

- **Config**: what you author: `scenes/`, `rooms/`, `shared/` and an optional `config.yaml`.
- **State**: what changes while you play: the SQLite database (`db.sqlite`) and screen state.

| OS | Config folder | State folder |
| --- | --- | --- |
| Linux | `~/.config/mechanus` (`$XDG_CONFIG_HOME`) | `~/.local/state/mechanus` (`$XDG_STATE_HOME`) |
| macOS | `~/Library/Preferences/mechanus` | `~/Library/Application Support/mechanus/State` |
| Windows | `%APPDATA%\mechanus` | `%LOCALAPPDATA%\mechanus` |

`./mechanus config paths` prints the folders it uses on your machine.

### Keep everything in one folder

If a `.config` or `.local` folder exists in the folder you start the server from, it is used instead of the OS folder. Create both next to the binary to keep the whole install (program, config and data) in one place, which is easy to back up, move or run as a service:

```bash
# Linux / macOS
mkdir .config .local
```

```powershell
# Windows
mkdir .config, .local
```

```text
mechanus/
├── mechanus
├── web/
├── .config/        scenes/, rooms/, shared/, config.yaml
└── .local/         db.sqlite, screens/
```

Always start the server from this folder, or the OS folders are used instead. The guides below assume this layout.

### Your first scene

Copy the `scenes`, `rooms` and `shared` folders of the example in [`docs/examples/battlemap`](../examples/battlemap) into your config folder to start from a working scene. See [battle maps](../configuration/battle-maps.md) for the file formats.

## Settings

Every setting can be a flag, an environment variable or a key in `config.yaml`:

| Flag | Environment variable | `config.yaml` |
| --- | --- | --- |
| `--server.port 9000` | `SERVER_PORT=9000` | `server: { port: 9000 }` |
| `--server.static.folder web` | `SERVER_STATIC_FOLDER=web` | `server: { static: { folder: web } }` |

`./mechanus server --help` lists every flag with its variable and default. To avoid typing `--server.static.folder web` on every start, put it in `.config/config.yaml`:

```yaml
server:
  static:
    folder: web
```

See [config file](../configuration/config-file.md) for where `config.yaml` is looked for, [database](../configuration/database.md) for PostgreSQL / MySQL, and [integrations](../configuration/integrations.md) for MQTT devices. For HTTPS, pass a certificate with `--server.tls.cert` and `--server.tls.key`.

## Run it in the background

To keep Mechanus running without a terminal open and start it on boot, run it as a service. Each example uses the [one-folder layout](#keep-everything-in-one-folder) and assumes the admin account already exists (start it once by hand first, or read the generated password from the service log).

### Linux (systemd)

Install into `/opt/mechanus` and run it as its own user:

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin mechanus
sudo mkdir -p /opt/mechanus/.config /opt/mechanus/.local
sudo tar -xzf mechanus_0.4.1_linux_amd64.tar.gz -C /opt/mechanus
sudo chown -R mechanus:mechanus /opt/mechanus/.config /opt/mechanus/.local
```

Create `/etc/systemd/system/mechanus.service`:

```ini
[Unit]
Description=Mechanus
After=network-online.target
Wants=network-online.target

[Service]
User=mechanus
WorkingDirectory=/opt/mechanus
ExecStart=/opt/mechanus/mechanus server --server.static.folder web
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Enable and start it:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now mechanus
journalctl -u mechanus -f   # follow the log
```

### macOS (launchd)

Create `~/Library/LaunchAgents/com.github.daanv2.mechanus.plist`, with `/Users/you/mechanus` replaced by the folder you extracted to:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.github.daanv2.mechanus</string>
  <key>ProgramArguments</key>
  <array>
    <string>/Users/you/mechanus/mechanus</string>
    <string>server</string>
    <string>--server.static.folder</string>
    <string>web</string>
  </array>
  <key>WorkingDirectory</key>
  <string>/Users/you/mechanus</string>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>/Users/you/mechanus/mechanus.log</string>
  <key>StandardErrorPath</key>
  <string>/Users/you/mechanus/mechanus.log</string>
</dict>
</plist>
```

Load it; it then starts whenever you log in:

```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.github.daanv2.mechanus.plist
```

Stop and unload it with `launchctl bootout gui/$(id -u)/com.github.daanv2.mechanus`. Clear the [quarantine flag](#first-start-on-macos) first, or launchd can't start the binary.

### Windows (Task Scheduler)

Extract to `C:\mechanus` with `.config` and `.local` folders in it, then in a PowerShell **run as administrator**:

```powershell
$action   = New-ScheduledTaskAction -Execute "C:\mechanus\mechanus.exe" `
              -Argument "server --server.static.folder web" -WorkingDirectory "C:\mechanus"
$trigger  = New-ScheduledTaskTrigger -AtStartup
$settings = New-ScheduledTaskSettingsSet -ExecutionTimeLimit ([TimeSpan]::Zero) `
              -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
Register-ScheduledTask -TaskName "Mechanus" -Action $action -Trigger $trigger `
  -Settings $settings -User "SYSTEM" -RunLevel Highest

# A task has no window to show the firewall prompt in, so allow it up front
New-NetFirewallRule -DisplayName "Mechanus" -Direction Inbound -Action Allow `
  -Program "C:\mechanus\mechanus.exe" -Profile Private

Start-ScheduledTask -TaskName "Mechanus"
```

It now starts on every boot, before anyone logs in. Stop it with `Stop-ScheduledTask -TaskName Mechanus`, remove it with `Unregister-ScheduledTask -TaskName Mechanus`. The task has no console, so its log isn't kept; to read the log (for example the generated admin password), start it by hand once from `C:\mechanus` instead.

## Update

Your config and data aren't in the archive, so updating only replaces the program:

1. Stop the server (`Ctrl+C`, or stop the service).
2. Download and verify the new archive.
3. Delete the old `mechanus` binary and `web/` folder; leave `.config/` and `.local/` alone.
4. Extract the new archive into the same folder.
5. Start the server again. Database migrations run on start.

Read the release notes before skipping several versions, and back up the state folder (`.local/`, or the OS state folder) first.

## Uninstall

Stop and remove the service if you set one up, then delete the extracted folder. If you didn't use the one-folder layout, also delete the config and state folders from [the table above](#where-your-data-lives).

## Troubleshooting

- **The page is blank or 404s**: the server can't find the web app. Start it from the extracted folder, or pass the full path: `--server.static.folder /opt/mechanus/web`.
- **`address already in use` / `Only one usage of each socket address`**: something else uses port 8080. Pick another with `--server.port 8081`.
- **Other devices can't connect**: it's almost always the firewall, see the first-start notes for [Windows](#first-start-on-windows), [macOS](#first-start-on-macos) and [Linux](#first-start-on-linux), and [Getting started → Connect the other devices](./README.md#3-connect-the-other-devices).
- **`permission denied` on Linux / macOS**: the executable bit got lost (for example after copying through a Windows share). Run `chmod +x mechanus`.
- **`exec format error` / `cannot execute binary file`**: wrong archive for this machine. Check the [architecture](#pick-your-archive).
- **Lost the generated admin password**: it is only printed on the start that created the account. If nobody else has an admin account, stop the server, move the database (`db.sqlite` in the state folder) away and start again with `--initialize.admin.*`. That starts with an empty database.
