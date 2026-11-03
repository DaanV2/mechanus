import { Sprite } from 'pixi.js';
import type { Layer } from '../../../proto/screens/v1/state_pb';
import type { TextureCache } from './textures';
import type { MapSpace } from './layer';

/** Draws an image layer in map pixels, undefined when its image can't be loaded. */
export async function createImageLayer(
  layer: Layer,
  space: MapSpace,
  textures: TextureCache
): Promise<Sprite | undefined> {
  const texture = await textures.load(layer.asset);
  if (!texture) return undefined;

  const sprite = new Sprite(texture);
  sprite.label = layer.id;
  sprite.alpha = layer.opacity;
  sprite.position.set((layer.position?.x ?? 0) * space.cell, (layer.position?.y ?? 0) * space.cell);
  if (layer.size) sprite.setSize(layer.size.x * space.cell, layer.size.y * space.cell);

  return sprite;
}
