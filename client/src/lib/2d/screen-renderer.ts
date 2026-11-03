import { Container, Application as PixiApplication } from 'pixi.js';
import type { SplashScreen as SplashScreenState } from '../../proto/screens/v1/splash_screen_pb';
import { ViewKind, type Asset, type ScreenState } from '../../proto/screens/v1/state_pb';
import { SplashScreen } from './components/splashscreen';
import { createBackdrop } from './layers/backdrop';
import { createBattleMap } from './layers/battlemap';
import type { ViewContent } from './layers/layer';
import { createOverlays } from './layers/overlay-layer';
import { TextureCache } from './layers/textures';

/**
 * Renders the screen state pushed by the server: the view (battle map layers, scenery, splash, blank),
 * its overlays and, on top of everything, the splash screen while it is shown.
 *
 * The render loop is Pixi's ticker: every frame it applies what changed since the last one, so a burst
 * of states or resize events causes one rebuild, then Pixi draws the stage.
 */
export class ScreenRenderer {
  private readonly app = new PixiApplication();
  private readonly view = new Container();
  private readonly overlays = new Container();
  private readonly splash = new SplashScreen();
  private readonly textures = new TextureCache();

  private state?: ScreenState;
  private splashState?: SplashScreenState;
  private contents: ViewContent[] = [];
  private stateChanged = false;
  private resized = false;
  // Incremented on every rebuild, so a slow asset load can't draw over a newer state
  private buildID = 0;

  async init(parent: HTMLElement) {
    await this.app.init({ background: '#000000', resizeTo: parent });
    parent.appendChild(this.app.canvas);

    this.app.stage.addChild(this.view, this.overlays, this.splash);
    this.splash.visible = false;
    this.app.renderer.on('resize', () => (this.resized = true));
    this.app.ticker.add(this.tick, this);
  }

  destroy() {
    this.buildID++;
    this.app.destroy(true, { children: true });
    this.textures.clear();
  }

  /** Replaces the screen state; a new state from a scene hides the splash screen. */
  setState(state: ScreenState | undefined) {
    this.state = state;
    if (state?.view && state.view.kind !== ViewKind.BLANK) this.splashState = undefined;
    this.stateChanged = true;
  }

  setSplash(splash: SplashScreenState | undefined) {
    this.splashState = splash;
    this.renderSplash();
  }

  private tick() {
    if (this.stateChanged) {
      this.stateChanged = false;
      void this.build().catch((err) => console.error('could not build screen state', err));
    }
    if (this.resized) {
      this.resized = false;
      this.layout();
    }
  }

  private layout() {
    for (const content of this.contents) content.layout(this.app.screen);
    this.renderSplash();
  }

  /** Loads what the state needs while the old state stays on screen, then swaps them in one go. */
  private async build() {
    const id = ++this.buildID;
    const state = this.state;

    const view = await createView(state, this.textures);
    const overlays = createOverlays(state);
    if (id !== this.buildID) {
      view.root.destroy({ children: true });
      overlays.root.destroy({ children: true });

      return;
    }

    for (const old of this.view.removeChildren()) old.destroy({ children: true });
    for (const old of this.overlays.removeChildren()) old.destroy({ children: true });
    this.textures.release(usedAssets(state));

    this.app.renderer.background.color = state?.environment?.backgroundColor || '#000000';
    this.view.addChild(view.root);
    this.overlays.addChild(overlays.root);
    this.contents = [view, overlays];
    this.layout();
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
}

// TODO: infoboard and custom views
function createView(state: ScreenState | undefined, textures: TextureCache): Promise<ViewContent> {
  switch (state?.view?.kind) {
    case ViewKind.BATTLEMAP:
      return createBattleMap(state, textures);
    case ViewKind.SCENERY:
      return createBackdrop(state.view.asset, 'cover', textures);
    case ViewKind.SPLASH:
      return createBackdrop(state.view.asset, 'contain', textures);
    default:
      return Promise.resolve({ root: new Container(), layout: () => {} });
  }
}

/** The URLs of every asset the state shows. */
function usedAssets(state: ScreenState | undefined): Set<string> {
  const assets: (Asset | undefined)[] = [state?.view?.asset];
  for (const layer of state?.layer ?? []) {
    assets.push(layer.asset);
    assets.push(...layer.entity.map((e) => e.asset));
    assets.push(...layer.instanceGroup.map((g) => g.asset));
  }

  return new Set(assets.flatMap((a) => (a?.url ? [a.url] : [])));
}
