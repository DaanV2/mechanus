import type { WebsocketHandler } from '$lib/networking/websocket';
import { create } from '@bufbuild/protobuf';
import * as pixi from 'pixi.js';
import { InitialSetupRequestSchema } from '../../proto/screens/v1/setup_pb';

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
  async init(conn?: WebsocketHandler) {
    await this._app.init({
      background: '#1099bb',
      resizeTo: window,
      preference: 'webgpu'
    });

    if (this._app.canvas == undefined) {
      throw new Error('app canvas not initialized');
    }
    document.body.appendChild(this._app.canvas);

    // TODO: Implement layer manager to handle resize events across all layers
    // this._app.renderer.on('resize', () => this.layers.handleResize(this._app.renderer));

    // TODO: Implement layer activation system and enable splashscreen on startup
    // this.layers.activate('splashScreen', this._app.stage);

    // Await the connection to open then send initial request

    if (conn) {
      conn.addEventListener('open', () => {
        conn.send({
          action: {
            case: 'initialSetupRequest',
            value: create(InitialSetupRequestSchema, {})
          }
        });
      });
    }
  }

  destroy() {
    document.body.removeChild(this._app.canvas);
    this._app.destroy();
  }

  get stage() {
    return this._app.stage;
  }

  // TODO: Implement start() - begin render loop, process incoming screen state updates
  start() {}
}
