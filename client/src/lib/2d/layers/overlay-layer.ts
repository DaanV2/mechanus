import { Container, Graphics, Text } from 'pixi.js';
import { OverlayKind, type ScreenState } from '../../../proto/screens/v1/state_pb';
import type { ScreenSize, ViewContent } from './layer';

const MARGIN = 16;
const PADDING = 12;
const FONT_SIZE = 20;

/**
 * Draws the view's overlays on top of it, in screen pixels. The `entities` overlay lists the entities on
 * the map; the `initiative` tracker and custom components have nothing to show yet.
 */
export function createOverlays(state: ScreenState | undefined): ViewContent {
  const root = new Container({ label: 'overlays' });
  const layouts: ((screen: ScreenSize) => void)[] = [];

  for (const overlay of state?.view?.overlay ?? []) {
    if (overlay.kind !== OverlayKind.ENTITIES) continue;

    const names = (state?.layer ?? []).flatMap((l) => l.entity.map((e) => e.name));
    if (names.length === 0) continue;

    const panel = createEntityList(names);
    root.addChild(panel);
    layouts.push((screen) => panel.position.set(screen.width - panel.width - MARGIN, MARGIN));
  }

  return { root, layout: (screen) => layouts.forEach((layout) => layout(screen)) };
}

function createEntityList(names: string[]): Container {
  const panel = new Container({ label: 'entities' });
  const text = new Text({
    text: names.join('\n'),
    style: { fontSize: FONT_SIZE, fill: '#ffffff' }
  });
  text.position.set(PADDING, PADDING);

  const background = new Graphics()
    .roundRect(0, 0, text.width + PADDING * 2, text.height + PADDING * 2, 8)
    .fill({ color: '#000000', alpha: 0.7 });
  panel.addChild(background, text);

  return panel;
}
