# Battle maps

A battle map is what a screen shows for the `battlemap` view: grid settings plus an ordered list of typed layers. It is a hand-written JSON file, `<name>.map.json`, in the scene's folder. See [scenes.md](../architecture/scenes.md#screens--views) for how this fits in the scene model.

A scene points to a map from a screen instruction:

```json
{ "view": "battlemap", "asset": "maps/tavern.map.json" }
```

The file extension makes it a map. Any other asset (`"asset": "maps/cave.webp"`) is shown as one image under a default grid of 70 pixel cells, sized by the image.

Map files are configuration: the server reads them when it instructs a screen and only sends what the screen may see. They are not served under `/assets/`, as they name where hidden entities are. A file is read every time a screen is instructed, so an edit shows the next time the scene is activated.

## Folder layout

```text
<app config dir>/
├── rooms/room.dat                    The room: devices and target bindings
├── scenes/
│   ├── tavern.dat                    The scene config
│   └── tavern/                       The scene's assets
│       ├── maps/tavern.map.json
│       ├── images/floor.webp
│       └── tokens/barkeep.png
└── shared/                           Assets shared between scenes
    └── props/barrel.png
```

Asset references are relative to the scene's folder (`images/floor.webp`) or point to the shared folder (`shared:/props/barrel.png`), including the references inside a map file.

## Format

```json
{
  "version": 0,
  "name": "Tavern",
  "background_color": "#101010",
  "grid": { "cell_size": 70, "columns": 16, "rows": 9, "line_color": "#000000" },
  "layers": [
    { "id": "floor", "kind": "image", "asset": "images/floor.svg" },
    {
      "id": "rug",
      "kind": "image",
      "asset": "images/rug.svg",
      "position": { "x": 6, "y": 3 },
      "size": { "x": 4, "y": 3 }
    },
    {
      "id": "tokens",
      "kind": "entities",
      "instance_groups": [
        {
          "id": "barrels",
          "asset": "shared:/props/barrel.svg",
          "instances": [{ "x": 1, "y": 1 }, { "x": 2, "y": 1, "rotation": 45 }]
        }
      ],
      "entities": [
        { "entity": "barkeep", "asset": "tokens/barkeep.svg", "position": { "x": 11, "y": 2 } },
        { "entity": "hero", "position": { "x": 4, "y": 4 } }
      ]
    }
  ]
}
```

Unknown fields and invalid values are an error, so a typo doesn't silently do nothing: activating the scene reports it as a device error, and the screen keeps what it was showing.

### Map

| Field              | Required | Description                                                    |
| ------------------ | -------- | -------------------------------------------------------------- |
| `version`          |          | Format version, only `0` exists                                |
| `name`             |          | Display name                                                   |
| `background_color` |          | `#rrggbb` shown behind the layers and around the map           |
| `grid`             | yes      | See [Grid](#grid)                                              |
| `layers`           |          | Drawn in order, the first layer at the bottom                  |

### Grid

Positions and sizes in a map are in grid units: 1 is one cell. One map pixel is one pixel of an image at its natural size. A screen scales the whole map to fit.

| Field                  | Required | Description                                                             |
| ---------------------- | -------- | ----------------------------------------------------------------------- |
| `cell_size`            | yes      | Size of one cell in map pixels                                          |
| `columns`, `rows`      | yes      | Size of the map in cells                                                |
| `x_offset`, `y_offset` |          | Moves the grid, entities and instances, in map pixels. Use it to line them up with a grid painted on an image. Image layers stay where they are |
| `hidden`               |          | `true` to not draw the grid lines                                       |
| `line_color`           |          | `#rrggbb` of the grid lines, black by default                           |

The grid lines are drawn above the image layers and below the first entities layer.

### Image layer (`"kind": "image"`)

| Field      | Required | Description                                                           |
| ---------- | -------- | --------------------------------------------------------------------- |
| `id`       | yes      | Unique within the map                                                 |
| `asset`    | yes      | The image, or a video                                                 |
| `position` |          | Top left corner in grid units, from the map's top left. Default `0,0` |
| `size`     |          | Size in grid units. Default the image's natural size                  |
| `opacity`  |          | Between `0` and `1`. Default `1`                                      |

### Entities layer (`"kind": "entities"`)

Holds `instance_groups` and, on top of them, `entities`. An `id` is required, unique within the map.

**`entities`** places the scene's entities (`entities` in the scene config) on the map. A placement draws nothing while its entity is hidden, removed or not in the scene: hidden entities are never sent to a screen. An entity can be placed once per map. Within a layer, tokens are drawn sorted by entity ID.

| Field      | Required | Description                                                  |
| ---------- | -------- | ------------------------------------------------------------ |
| `entity`   | yes      | ID of the entity in the scene                                |
| `asset`    |          | Token image. Default a coloured disc with the name's initial |
| `position` | yes      | Top left corner in grid units                                |
| `size`     |          | Size in grid units. Default one cell                         |
| `rotation` |          | Degrees around the center                                    |

The position is where the entity starts. Moving tokens during play is not built yet; those positions will be stored in the scene's state, not in this file.

**`instance_groups`** draws one asset at many positions, for scenery such as trees, barrels and rubble. Instances have no identity in the scene. A group's instances are drawn together in one go, so a map can hold thousands of them.

| Field       | Required | Description                                                         |
| ----------- | -------- | ------------------------------------------------------------------- |
| `id`        | yes      | Unique within the layer                                             |
| `asset`     | yes      | Image of every instance                                             |
| `size`      |          | Size of every instance in grid units. Default one cell              |
| `instances` |          | Each with `x`, `y` (top left, in grid units) and optional `rotation` |

## Example scene

[`docs/examples/battlemap`](../examples/battlemap) is a small scene to develop against: a tavern with image layers, barrels as an instance group, an NPC with a token, a player with a placeholder token, a hidden enemy, and a `night` variant with another map. It comes with a room binding the `table` target to the device `example-tv`.

`docker compose watch` mounts it into the server. Running the server yourself, copy the folders into a `.config` folder in the directory you start the server from, which then becomes the app config folder (or into your [app config folder](./config-file.md)):

```bash
cp -r docs/examples/battlemap/. server/.config/
```

```bash
just server start-server
```

Then:

1. Log in, and open the TV view at `/views/devices/?device=example-tv`. Looking at a device's screen takes a logged in user. A screen without credentials goes to the pairing page instead.
2. In another tab, open the GM view at `/views/game-master/`.
3. Activate **Example tavern**, and switch to **Night**.

To use a real screen, pair it at `/devices` and bind it to the `table` target.
