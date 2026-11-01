# Config file

Mechanus looks for a `config.yaml` in the folders below, in this order, and uses the first one it finds. Run `mechanus config paths` (or `go run main.go config paths`) to print the exact folders on your machine.

If a `.config` folder exists in the current working directory, it replaces the OS config folder in this list. The same goes for a `.user` folder, which replaces the OS user data folder.

## Linux

Follows the XDG Base Directory spec: `$XDG_CONFIG_HOME` defaults to `~/.config`, `$XDG_DATA_HOME` to `~/.local/share`.

- `.config/config.yaml`
- `~/.config/mechanus/.config/config.yaml`
- `~/.local/share/mechanus/.config/config.yaml`
- `~/.config/mechanus/config.yaml`
- `~/.local/share/mechanus/config.yaml`

## macOS

- `.config/config.yaml`
- `~/Library/Preferences/mechanus/.config/config.yaml`
- `~/Library/Application Support/mechanus/.config/config.yaml`
- `~/Library/Preferences/mechanus/config.yaml`
- `~/Library/Application Support/mechanus/config.yaml`

## Windows

Config and user data share the same folder, `%APPDATA%\mechanus` (falls back to `%USERPROFILE%\AppData\Roaming\mechanus`).

- `.\.config\config.yaml`
- `%APPDATA%\mechanus\.config\config.yaml`
- `%APPDATA%\mechanus\config.yaml`

## Other folders

Besides config, Mechanus keeps state (such as the SQLite database and screen state) in a state folder:

| OS      | State folder                                                       |
| ------- | ------------------------------------------------------------------ |
| Linux   | `$XDG_STATE_HOME/mechanus` (default `~/.local/state/mechanus`)     |
| macOS   | `~/Library/Application Support/mechanus/State`                     |
| Windows | `%LOCALAPPDATA%\mechanus`                                          |

A `.local` folder in the current working directory replaces it.
