import { server_grpc_url } from '$lib/config';
import { Container, Graphics, Sprite, Texture, Application as PixiApplication } from 'pixi.js';
import type { SplashScreen as SplashScreenState } from '../../proto/screens/v1/splash_screen_pb';
import {
  ViewKind,
  type Asset,
  type GridSettings,
  type ScreenState
} from '../../proto/screens/v1/state_pb';
import { SplashScreen } from './components/splashscreen';

/**
 * Renders the screen state pushed by the server: the view (battle map, scenery, splash, blank) and,
 * on top of everything, the splash screen while it is shown.
 */
export class ScreenRenderer {
  private readonly app = new PixiApplication();
  private readonly view = new Container();
  private readonly splash = new SplashScreen();

  private state?: ScreenState;
  private splashState?: SplashScreenState;
  // Incremented on every render, so a slow asset load can't draw over a newer state
  private renderID = 0;

  async init(parent: HTMLElement) {
    await this.app.init({ background: '#000000', resizeTo: parent });
    parent.appendChild(this.app.canvas);

    this.app.stage.addChild(this.view, this.splash);
    this.splash.visible = false;
    this.app.renderer.on('resize', () => void this.render());
  }

  destroy() {
    this.app.destroy(true, { children: true });
  }

  /** Replaces the screen state; a new state from a scene hides the splash screen. */
  setState(state: ScreenState | undefined) {
    this.state = state;
    if (state?.view && state.view.kind !== ViewKind.BLANK) this.splashState = undefined;
    void this.render();
  }

  setSplash(splash: SplashScreenState | undefined) {
    this.splashState = splash;
    this.renderSplash();
  }

  private renderSplash() {
    const s = this.splashState ?? this.state?.splashScreen;
    this.splash.visible = s?.show ?? false;
    if (!s?.show) return;

    this.splash.title.text = s.title ?? '';
    this.splash.subtitle.text = s.subtitle ?? '';
    this.splash.background
      .clear()
      .rect(0, 0, 1, 1)
      .fill(s.backgroundHex || '#000000');
    this.splash.title.anchor.set(0.5);
    this.splash.subtitle.anchor.set(0.5);
    this.splash.handleResize(this.app.screen);
  }

  private async render() {
    const id = ++this.renderID;
    const view = this.state?.view;
    const screen = this.app.screen;

    const children: Container[] = [];
    switch (view?.kind) {
      case ViewKind.BATTLEMAP: {
        const map = await loadSprite(view.asset);
        if (map) {
          fit(map, screen, 'contain');
          children.push(map);
          if (this.state?.grid?.show) children.push(drawGrid(this.state.grid, map));
        }
        break;
      }
      case ViewKind.SCENERY:
      case ViewKind.SPLASH: {
        const image = await loadSprite(view.asset);
        if (image) {
          fit(image, screen, view.kind === ViewKind.SCENERY ? 'cover' : 'contain');
          children.push(image);
        }
        break;
      }
      // TODO: infoboard and custom views, overlays and entity layers
      default:
        break;
    }
    if (id !== this.renderID) return;

    this.app.renderer.background.color = this.state?.environment?.backgroundColor || '#000000';
    for (const old of this.view.removeChildren()) old.destroy();
    this.view.addChild(...children);
    this.renderSplash();
  }
}

/** Asset URLs from the server are relative to the server, which may not be where the client is served from. */
function assetURL(asset: Asset): string {
  return new URL(asset.url, server_grpc_url()).toString();
}

async function loadSprite(asset: Asset | undefined): Promise<Sprite | undefined> {
  if (!asset?.url) return undefined;

  try {
    if (asset.mediaType.startsWith('video/')) {
      const video = document.createElement('video');
      video.crossOrigin = 'anonymous';
      video.src = assetURL(asset);
      video.loop = true;
      video.muted = true;
      video.playsInline = true;
      await video.play();

      return new Sprite(Texture.from(video));
    }

    const image = new Image();
    image.crossOrigin = 'anonymous';
    image.src = assetURL(asset);
    await image.decode();

    return new Sprite(Texture.from(image));
  } catch (err) {
    console.error('could not load asset', asset.url, err);

    return undefined;
  }
}

/** Scales and centres a sprite on the screen: `contain` shows all of it, `cover` fills the screen. */
function fit(sprite: Sprite, screen: { width: number; height: number }, mode: 'contain' | 'cover') {
  const scaleX = screen.width / sprite.texture.width;
  const scaleY = screen.height / sprite.texture.height;
  const scale = mode === 'contain' ? Math.min(scaleX, scaleY) : Math.max(scaleX, scaleY);

  sprite.scale.set(scale);
  sprite.position.set(
    (screen.width - sprite.texture.width * scale) / 2,
    (screen.height - sprite.texture.height * scale) / 2
  );
}

/** Draws grid lines over a battle map; cell size and offsets are in pixels of the map image. */
function drawGrid(grid: GridSettings, map: Sprite): Graphics {
  const g = new Graphics();
  const cell = grid.cellSize > 1 ? grid.cellSize : 70;
  const { width, height } = map.texture;

  for (let x = grid.xOffset % cell; x <= width; x += cell) g.moveTo(x, 0).lineTo(x, height);
  for (let y = grid.yOffset % cell; y <= height; y += cell) g.moveTo(0, y).lineTo(width, y);
  g.stroke({ width: 1, color: grid.lineColor || '#000000', alpha: 0.5 });

  g.position.copyFrom(map.position);
  g.scale.copyFrom(map.scale);

  return g;
}
