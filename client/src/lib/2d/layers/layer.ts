import type { Container } from 'pixi.js';
import type { GridSettings } from '../../../proto/screens/v1/state_pb';

export interface ScreenSize {
  width: number;
  height: number;
}

/** What a view draws: a tree of Pixi containers that can be fitted to the screen again after a resize. */
export interface ViewContent {
  root: Container;
  layout(screen: ScreenSize): void;
}

/**
 * Converts grid units into map pixels. Grid cell (0, 0) starts at the grid's offset, images are placed
 * from the map's origin and so aren't moved by it.
 */
export class MapSpace {
  readonly cell: number;
  readonly xOffset: number;
  readonly yOffset: number;

  constructor(grid: GridSettings) {
    this.cell = grid.cellSize;
    this.xOffset = grid.xOffset;
    this.yOffset = grid.yOffset;
  }

  x(cells: number) {
    return this.xOffset + cells * this.cell;
  }

  y(cells: number) {
    return this.yOffset + cells * this.cell;
  }
}

export function degreesToRadians(degrees: number) {
  return (degrees * Math.PI) / 180;
}
