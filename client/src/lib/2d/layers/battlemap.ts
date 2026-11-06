import { Container, Point, Sprite } from 'pixi.js';
import {
  LayerKind,
  type Camera,
  type Entity,
  type GridSettings,
  type ScreenState
} from '../../../proto/screens/v1/state_pb';
import { createEntityLayer } from './entity-layer';
import { createGridLayer } from './grid-layer';
import { createImageLayer } from './image-layer';
import { MapSpace, type ScreenSize, type ViewContent } from './layer';
import type { TextureCache } from './textures';

/** A point in grid units. */
export interface GridPoint {
  x: number;
  y: number;
}

/**
 * The layer stack of a battle map, in map pixels: the state's layers in order, with the grid above the
 * images and below the first entity layer. Layout fits that map to the screen, then applies the camera.
 */
export class BattleMapView implements ViewContent {
  readonly root = new Container({ label: 'battlemap' });
  private screen: ScreenSize = { width: 0, height: 0 };
  // Center of the screen in map pixels, and the zoom on top of fitting the map to the screen
  private focus = { x: 0, y: 0 };
  private zoom = 1;

  constructor(
    readonly space: MapSpace,
    readonly size: ScreenSize
  ) {
    this.setCamera(undefined);
  }

  layout(screen: ScreenSize) {
    this.screen = { width: screen.width, height: screen.height };
    if (this.size.width <= 0 || this.size.height <= 0) return;

    const scale = this.scale();
    this.root.scale.set(scale);
    this.root.position.set(
      screen.width / 2 - this.focus.x * scale,
      screen.height / 2 - this.focus.y * scale
    );
  }

  /** Applies a camera, unset (or centred on the origin) shows the whole map. */
  setCamera(camera: Camera | undefined) {
    const center = camera?.center;
    this.focus =
      center && (center.x !== 0 || center.y !== 0)
        ? { x: this.space.x(center.x), y: this.space.y(center.y) }
        : { x: this.size.width / 2, y: this.size.height / 2 };
    this.zoom = camera?.zoom || 1;
    this.layout(this.screen);
  }

  /** The current camera: center in grid units and zoom. */
  camera(): { center: GridPoint; zoom: number } {
    return {
      center: {
        x: (this.focus.x - this.space.xOffset) / this.space.cell,
        y: (this.focus.y - this.space.yOffset) / this.space.cell
      },
      zoom: this.zoom
    };
  }

  /** Moves the camera by a distance in screen pixels. */
  pan(dx: number, dy: number) {
    const scale = this.scale();
    this.focus = { x: this.focus.x - dx / scale, y: this.focus.y - dy / scale };
    this.layout(this.screen);
  }

  /** Multiplies the zoom, keeping it between the server's limits. */
  zoomBy(factor: number) {
    this.zoom = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, this.zoom * factor));
    this.layout(this.screen);
  }

  /** The token of an entity, if this map draws it. */
  token(layerId: string, entityId: string): Container | undefined {
    const layer = this.root.getChildByLabel(layerId);

    return layer?.getChildByLabel(entityId) ?? undefined;
  }

  /** Every token, with the layer it is on. */
  tokens(state: ScreenState): { layerId: string; entity: Entity; token: Container }[] {
    return state.layer.flatMap((layer) =>
      layer.entity.flatMap((entity) => {
        const token = this.token(layer.id, entity.id);

        return token ? [{ layerId: layer.id, entity, token }] : [];
      })
    );
  }

  /** Moves a token to the entity's position. False when this map doesn't draw it. */
  moveToken(layerId: string, entity: Entity): boolean {
    const token = this.token(layerId, entity.id);
    if (!token) return false;

    const { width, height } = this.tokenSize(entity);
    token.position.set(
      this.space.x(entity.position?.x ?? 0) + width / 2,
      this.space.y(entity.position?.y ?? 0) + height / 2
    );

    return true;
  }

  /** The grid position of a token's top left corner, from where its center is. */
  tokenPosition(entity: Entity, token: Container): GridPoint {
    const { width, height } = this.tokenSize(entity);

    return {
      x: (token.x - width / 2 - this.space.xOffset) / this.space.cell,
      y: (token.y - height / 2 - this.space.yOffset) / this.space.cell
    };
  }

  /** Converts a point on the screen to map pixels. */
  toMap(x: number, y: number): Point {
    return this.root.toLocal(new Point(x, y));
  }

  private tokenSize(entity: Entity) {
    return {
      width: (entity.size?.x ?? 1) * this.space.cell,
      height: (entity.size?.y ?? 1) * this.space.cell
    };
  }

  private scale() {
    const fit = Math.min(
      this.screen.width / this.size.width,
      this.screen.height / this.size.height
    );

    return fit * this.zoom;
  }
}

/** Zoom limits, must match the server's (scenes.MinZoom / MaxZoom). */
export const MIN_ZOOM = 0.1;
export const MAX_ZOOM = 10;

/** Builds the battle map of a state. A state without a grid gives an empty view. */
export async function createBattleMap(
  state: ScreenState,
  textures: TextureCache
): Promise<ViewContent> {
  const grid = state.grid;
  if (!grid || grid.cellSize <= 0) return { root: new Container(), layout: () => {} };

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
  const view = new BattleMapView(space, size);
  const firstEntityLayer = state.layer.findIndex((l) => l.kind === LayerKind.ENTITIES);
  layers.forEach((layer, i) => {
    if (i === firstEntityLayer && grid.show) view.root.addChild(createGridLayer(grid, size));
    if (layer) view.root.addChild(layer);
  });
  if (firstEntityLayer < 0 && grid.show) view.root.addChild(createGridLayer(grid, size));
  view.setCamera(state.camera);

  return view;
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
