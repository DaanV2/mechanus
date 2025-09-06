# Mechanus (WIP)

[![⚙️ CI](https://github.com/DaanV2/mechanus/actions/workflows/checks.yaml/badge.svg)](https://github.com/DaanV2/mechanus/actions/workflows/checks.yaml)

This project provides software for a local vttrpg / gaming setup, by managing multiple screens/clients/players/dm to help create a ttrpg experience using hardware, allowing a virtual table screen, scenery screens etc.

This project uses server driven UI, where the server written in [Go](https://go.dev/), The screen (or client) side is using [Pixi.js](https://pixijs.com/) for the rendering of the screens. While [connectRpc](https://connectrpc.com/) pins down the protocol between server and client. Each screen or client can connect and either login in as a viewer, device, player, or GM and either get their own screen rendered.

This allows the server and thus GM to control a TV screen on the table as battle map. Or TV or PC monitors to display scenery, maps or anything else. Intheory when also music and lightning is integrated, it will allow you to control the entire game room via browser.

![overview](./docs/assets/overview.svg)

## Contents

- [Developing](./docs/development.md)
- [Contributing](./docs/contributing.md)
- [Server](./server/README.md)

## TODO Features

- [x] Use https://connectrpc.com/docs/go/getting-started
- [ ] Manage the TV on the table for battle map displayment
- [x] Local mDNS
- [ ] Player views
- [ ] DM views
- [ ] Support for other vttrpg file formats

