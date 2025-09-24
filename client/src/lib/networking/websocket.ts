import { server_websocket_url } from '$lib/config';
import { Cookie } from '$lib/storage';

export class WebsocketHandler {
  readonly screenid: string;
  readonly id: string;
  readonly token?: string;
  socket: WebSocket;

  constructor(screenid: string, id: string, token?: string) {
    this.screenid = screenid;
    this.id = id;
    this.token = token;
    const url = server_websocket_url() + `/api/v1/screen/${this.screenid}/${this.id}`;
    this.socket = new WebSocket(url);
  }

  getToken() {
    return this.token ?? Cookie.get('access-token');
  }
}
