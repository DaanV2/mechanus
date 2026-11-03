import { Container, Sprite } from 'pixi.js';
import type { Asset } from '../../../proto/screens/v1/state_pb';
import type { ViewContent } from './layer';
import type { TextureCache } from './textures';

/**
 * Draws one image or video centred on the screen, for the scenery and splash views: `contain` shows all
 * of it, `cover` fills the screen.
 */
export async function createBackdrop(
  asset: Asset | undefined,
  mode: 'contain' | 'cover',
  textures: TextureCache
): Promise<ViewContent> {
  const root = new Container({ label: 'backdrop' });
  const texture = await textures.load(asset);
  if (!texture) return { root, layout: () => {} };

  const sprite = new Sprite(texture);
  root.addChild(sprite);

  return {
    root,
    layout(screen) {
      const scaleX = screen.width / texture.width;
      const scaleY = screen.height / texture.height;
      const scale = mode === 'contain' ? Math.min(scaleX, scaleY) : Math.max(scaleX, scaleY);

      sprite.scale.set(scale);
      sprite.position.set(
        (screen.width - texture.width * scale) / 2,
        (screen.height - texture.height * scale) / 2
      );
    }
  };
}
