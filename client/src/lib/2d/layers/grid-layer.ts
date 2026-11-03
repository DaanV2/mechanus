import { Graphics } from 'pixi.js';
import type { GridSettings } from '../../../proto/screens/v1/state_pb';
import type { ScreenSize } from './layer';

/**
 * Draws the grid lines in map pixels: `columns` x `rows` cells from the offset, or, on a map without a
 * size, every line that falls on the map.
 */
export function createGridLayer(grid: GridSettings, size: ScreenSize): Graphics {
  const g = new Graphics();
  g.label = 'grid';

  const cell = grid.cellSize;
  const sized = grid.columns > 0 && grid.rows > 0;
  const left = sized ? grid.xOffset : grid.xOffset % cell;
  const top = sized ? grid.yOffset : grid.yOffset % cell;
  const right = sized ? grid.xOffset + grid.columns * cell : size.width;
  const bottom = sized ? grid.yOffset + grid.rows * cell : size.height;

  for (let x = left; x <= right; x += cell) g.moveTo(x, top).lineTo(x, bottom);
  for (let y = top; y <= bottom; y += cell) g.moveTo(left, y).lineTo(right, y);
  g.stroke({ pixelLine: true, color: grid.lineColor || '#000000', alpha: 0.5 });

  return g;
}
