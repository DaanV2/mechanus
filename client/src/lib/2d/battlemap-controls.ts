import { create } from '@bufbuild/protobuf';
import type { Container, FederatedPointerEvent, FederatedWheelEvent, Rectangle } from 'pixi.js';
import { Vector2Schema, type Entity, type ScreenState } from '../../proto/screens/v1/state_pb';
import type { BattleMapView, GridPoint } from './layers/battlemap';

/** What the controls ask the server to do. The server validates them and sends the result to every screen. */
export interface BattleMapActions {
  moveEntity(entityId: string, position: GridPoint): void;
  setCamera(center: GridPoint, zoom: number): void;
}

/** How long the wheel has to stop before the zoom is sent, so a scroll is sent once. */
const WHEEL_SEND_DELAY_MS = 200;
const WHEEL_ZOOM_STEP = 1.1;

/**
 * Lets a pointer move tokens (snapped to the grid) and pan / zoom the battle map. Moves are shown right
 * away and then sent; the server answers with the same move, or the old state when it refuses it.
 * Returns a function that removes the controls again.
 */
export function attachBattleMapControls(
  stage: Container,
  screen: Rectangle,
  view: BattleMapView,
  state: ScreenState,
  actions: BattleMapActions
): () => void {
  let dragging:
    { layerId: string; entity: Entity; token: Container; dx: number; dy: number } | undefined;
  let panning: { x: number; y: number } | undefined;
  let wheelTimer: ReturnType<typeof setTimeout> | undefined;

  const sendCamera = () => {
    const camera = view.camera();
    actions.setCamera(camera.center, camera.zoom);
  };

  for (const { layerId, entity, token } of view.tokens(state)) {
    token.eventMode = 'static';
    token.cursor = 'grab';
    token.on('pointerdown', (e: FederatedPointerEvent) => {
      e.stopPropagation();
      const p = view.toMap(e.global.x, e.global.y);
      dragging = { layerId, entity, token, dx: p.x - token.x, dy: p.y - token.y };
      token.cursor = 'grabbing';
    });
  }

  const down = (e: FederatedPointerEvent) => {
    panning = { x: e.global.x, y: e.global.y };
  };
  const move = (e: FederatedPointerEvent) => {
    if (dragging) {
      const p = view.toMap(e.global.x, e.global.y);
      dragging.token.position.set(p.x - dragging.dx, p.y - dragging.dy);
    } else if (panning) {
      view.pan(e.global.x - panning.x, e.global.y - panning.y);
      panning = { x: e.global.x, y: e.global.y };
    }
  };
  const up = () => {
    if (dragging) {
      const { layerId, entity, token } = dragging;
      const raw = view.tokenPosition(entity, token);
      const position = { x: Math.round(raw.x), y: Math.round(raw.y) };
      view.moveToken(layerId, { ...entity, position: create(Vector2Schema, position) });
      token.cursor = 'grab';
      dragging = undefined;
      actions.moveEntity(entity.id, position);
    } else if (panning) {
      panning = undefined;
      sendCamera();
    }
  };
  const wheel = (e: FederatedWheelEvent) => {
    view.zoomBy(e.deltaY < 0 ? WHEEL_ZOOM_STEP : 1 / WHEEL_ZOOM_STEP);
    clearTimeout(wheelTimer);
    wheelTimer = setTimeout(sendCamera, WHEEL_SEND_DELAY_MS);
  };

  stage.eventMode = 'static';
  stage.hitArea = screen;
  stage.on('pointerdown', down);
  stage.on('globalpointermove', move);
  stage.on('pointerup', up);
  stage.on('pointerupoutside', up);
  stage.on('wheel', wheel);

  return () => {
    clearTimeout(wheelTimer);
    stage.off('pointerdown', down);
    stage.off('globalpointermove', move);
    stage.off('pointerup', up);
    stage.off('pointerupoutside', up);
    stage.off('wheel', wheel);
  };
}
