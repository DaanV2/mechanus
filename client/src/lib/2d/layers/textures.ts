import { server_grpc_url } from '$lib/config';
import { Texture } from 'pixi.js';
import type { Asset } from '../../../proto/screens/v1/state_pb';

/**
 * Loads the textures of a screen state. The same asset is only loaded once, and textures the current
 * state no longer uses can be released, as a map switch would otherwise keep all of them on the GPU.
 */
export class TextureCache {
  private readonly textures = new Map<string, Promise<Texture | undefined>>();

  /** Loads the asset, resolves to undefined when it can't be loaded so one broken asset doesn't blank the screen. */
  load(asset: Asset | undefined): Promise<Texture | undefined> {
    if (!asset?.url) return Promise.resolve(undefined);

    let texture = this.textures.get(asset.url);
    if (!texture) {
      texture = loadTexture(asset);
      this.textures.set(asset.url, texture);
    }

    return texture;
  }

  /** Destroys every texture whose URL is not in `keep`. */
  release(keep: ReadonlySet<string>) {
    for (const [url, texture] of this.textures) {
      if (keep.has(url)) continue;
      this.textures.delete(url);
      void texture.then((t) => t?.destroy(true));
    }
  }

  /** Destroys every texture. */
  clear() {
    this.release(new Set());
  }
}

/** Asset URLs from the server are relative to the server, which may not be where the client is served from. */
function assetURL(asset: Asset): string {
  return new URL(asset.url, server_grpc_url()).toString();
}

async function loadTexture(asset: Asset): Promise<Texture | undefined> {
  try {
    if (asset.mediaType.startsWith('video/')) {
      const video = document.createElement('video');
      video.crossOrigin = 'anonymous';
      video.src = assetURL(asset);
      video.loop = true;
      video.muted = true;
      video.playsInline = true;
      await video.play();

      return Texture.from(video);
    }

    const image = new Image();
    image.crossOrigin = 'anonymous';
    image.src = assetURL(asset);
    await image.decode();

    return Texture.from(image);
  } catch (err) {
    console.error('could not load asset', asset.url, err);

    return undefined;
  }
}
