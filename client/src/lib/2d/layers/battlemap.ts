import { Container, Sprite } from 'pixi.js';
import { LayerKind, type GridSettings, type ScreenState } from '../../../proto/screens/v1/state_pb';
import { createEntityLayer } from './entity-layer';
import { createGridLayer } from './grid-layer';
import { createImageLayer } from './image-layer';
import { MapSpace, type ScreenSize, type ViewContent } from './layer';
import type { TextureCache } from './textures';

/**
 * Builds the layer stack of a battle map, in map pixels: the state's layers in order, with the grid
 * above the images and below the first entity layer. Layout fits that map to the screen.
 */
export async function createBattleMap(
  state: ScreenState,
  textures: TextureCache
): Promise<ViewContent> {
  const root = new Container({ label: 'battlemap' });
  const grid = state.grid;
  if (!grid || grid.cellSize <= 0) return { root, layout: () => {} };

  const space = new MapSpace(grid);
  const layers = await Promise.all(
    state.layer.map(async (layer) => {
      switch (layer.kind) {
        case LayerKind.IMAGE:
          return createImageLayer(layer, space, textures);
        case LayerKind.ENTITIES:
          return createEntityLayer(layer, space, textures);
        default:
          return undefined;
      }
    })
  );

  const size = mapSize(grid, layers);
  const firstEntityLayer = state.layer.findIndex((l) => l.kind === LayerKind.ENTITIES);
  layers.forEach((layer, i) => {
    if (i === firstEntityLayer && grid.show) root.addChild(createGridLayer(grid, size));
    if (layer) root.addChild(layer);
  });
  if (firstEntityLayer < 0 && grid.show) root.addChild(createGridLayer(grid, size));

  const camera = state.camera;
  const focus =
    camera?.center && (camera.center.x !== 0 || camera.center.y !== 0)
      ? { x: space.x(camera.center.x), y: space.y(camera.center.y) }
      : { x: size.width / 2, y: size.height / 2 };

  return {
    root,
    layout(screen) {
      if (size.width <= 0 || size.height <= 0) return;

      const fit = Math.min(screen.width / size.width, screen.height / size.height);
      const scale = fit * (camera?.zoom || 1);
      root.scale.set(scale);
      root.position.set(screen.width / 2 - focus.x * scale, screen.height / 2 - focus.y * scale);
    }
  };
}

/** The size of the map in map pixels: from its grid, or else as large as the images on it. */
function mapSize(grid: GridSettings, layers: (Container | undefined)[]): ScreenSize {
  if (grid.columns > 0 && grid.rows > 0) {
    return {
      width: grid.xOffset + grid.columns * grid.cellSize,
      height: grid.yOffset + grid.rows * grid.cellSize
    };
  }

  const size = { width: 0, height: 0 };
  for (const layer of layers) {
    if (layer instanceof Sprite) {
      size.width = Math.max(size.width, layer.x + layer.width);
      size.height = Math.max(size.height, layer.y + layer.height);
    }
  }

  return size;
}
