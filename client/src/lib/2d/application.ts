import * as pixi from 'pixi.js';

export class Application {
  // Create a new application
  private _app: pixi.Application;

  constructor() {
    this._app = new pixi.Application();
  }

  /**
   * Initialize the application
   * @returns
   */
  async init() {
    await this._app.init({
      background: '#1099bb',
      resizeTo: window,
      preference: 'webgpu'
    });

    if (this._app.canvas == undefined) {
      throw new Error('app canvas not initialized');
    }
    document.body.appendChild(this._app.canvas);
  }

  destroy() {
    document.body.removeChild(this._app.canvas);
    this._app.destroy();
  }

  get stage() {
    return this._app.stage
  }
}
