# Config file

Mechanus will look for a couple of places to look for a config.yaml:

## Linux

-  `/home/<user>/.config/mechanus/config.yaml`
-  `/home/<user>/.local/share/mechanus/config.yaml`
-  `/home/<user>/.config/mechanus/.config/config.yaml`
-  `/home/<user>/.local/share/mechanus/.config/config.yaml`
-  `.config/config.yaml`

## Macos

TODO: need someone to fill this in on a MAC: `go run main.go config paths` will give all the hints

## Windows

- `%APPDATA%\mechanus\config.yaml`
- `%APPDATA%\mechanus\.config`
- `.\.config\config.yaml`