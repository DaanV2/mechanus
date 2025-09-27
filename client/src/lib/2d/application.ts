import * as pixi from 'pixi.js';

export class Application {
  // Create a new application
  public readonly _app: pixi.Application;

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

    // this._app.renderer.on('resize', () => this.layers.handleResize(this._app.renderer));

    // STEP: activate splashscreen first before anything
    // this.layers.activate('splashScreen', this._app.stage);
  }

  destroy() {
    document.body.removeChild(this._app.canvas);
    this._app.destroy();
  }

  get stage() {
    return this._app.stage;
  }

  start() {}
}
