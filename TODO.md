# TODO

Tracked items that don't have a natural place in the code yet.

## Server

- [ ] **Scene management**: Implement scene lifecycle (create, load, switch scenes for a campaign)
- [ ] **Battle map rendering pipeline**: Process map files (dd2vtt, dungeondraft_map) into screen state that gets pushed to clients
- [ ] **Device registration & management**: Allow devices to register themselves, assign roles, and persist device-screen associations
- [ ] **Campaign session state**: Track active campaign session, connected players, and which scene is displayed
- [ ] **File format parsers**: Implement importers for Dungeondraft (.dd2vtt, .dungeondraft_map) and other VTTRPG map formats
- [ ] **WebSocket message handling**: Handle incoming client messages beyond initial setup (pan, zoom, token moves, etc.)

## Client

- [ ] **GM view**: Dashboard for the game master to control scenes, manage tokens, switch maps
- [ ] **Player view**: Individual player perspective with limited visibility / fog of war
- [ ] **Device/screen view**: Fullscreen battle map display mode for TV/table screens
- [ ] **Campaign management UI**: Create, edit, delete campaigns; invite/manage players
- [ ] **Map upload**: UI for uploading and managing battle map files
- [ ] **Layer system** (`src/lib/2d/`): Implement composable layers (background map, grid, tokens, fog of war, UI overlay)
- [ ] **Token system**: Add, move, resize tokens on the battle map
- [ ] **Logout flow**: Complete the logout handler (`src/lib/handlers/user.ts`)

## Documentation

- [ ] Client development guide (`client/docs/development.md`)
- [ ] WebSocket API protocol documentation (`docs/api/websocket.md`)
- [ ] macOS configuration paths (`docs/configuration/config-file.md`)
- [ ] Key-value storage link (`docs/configuration/database.md`)

## Future / Nice-to-have

- [ ] Music integration (ambient sound control from GM view)
- [ ] Lighting control integration (smart lights reacting to scenes)
- [ ] mDNS auto-discovery of server by clients on local network
- [ ] About page content & licenses (`client/src/routes/about/+page.svelte`)
