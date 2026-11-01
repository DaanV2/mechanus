import { server_websocket_url } from '$lib/config';
import { Session } from '$lib/authentication/session';
import { create, fromBinary, toBinary } from '@bufbuild/protobuf';
import {
  ClientMessageSchema,
  ClientMessagesSchema,
  type ClientMessage
} from '../../proto/screens/v1/client_pb';
import { ServerMessagesSchema, type ServerMessage } from '../../proto/screens/v1/server_pb';
import { ServerEventHandler } from './server-events';

// Subprotocols negotiated with the screen WebSocket, must match the server's authentication package.
const SUBPROTOCOL = 'mechanus.v1';
const BEARER_SUBPROTOCOL_PREFIX = 'mechanus.bearer.';

export class WebsocketHandler {
  readonly screenid: string;
  readonly id: string;
  readonly token?: string;
  readonly socket: WebSocket;
  readonly events: ServerEventHandler;

  /**
   * @param query extra query parameters, e.g. the device identity (`deviceId`, `deviceType`) a browser can't send as headers
   */
  constructor(screenid: string, id: string, token?: string, query?: Record<string, string>) {
    this.events = new ServerEventHandler();
    this.screenid = screenid;
    this.id = id;
    this.token = token;
    let url =
      server_websocket_url() +
      `/api/v1/screen/${encodeURIComponent(this.screenid)}/${encodeURIComponent(this.id)}`;
    if (query && Object.keys(query).length > 0) url += '?' + new URLSearchParams(query).toString();
    this.socket = new WebSocket(url, this.subprotocols());
    console.log('opening websocket on', url);
    this.socket.addEventListener('message', this._received.bind(this));
  }

  getToken() {
    return this.token ?? Session.accessToken();
  }

  /**
   * Browsers can't set an Authorization header on `new WebSocket(...)`, so the bearer token is passed as a
   * subprotocol instead. The server always selects {@link SUBPROTOCOL}, never echoing the token back.
   */
  private subprotocols(): string[] | undefined {
    const token = this.getToken()?.replace(/^Bearer /, '');
    if (!token) return undefined;

    return [SUBPROTOCOL, BEARER_SUBPROTOCOL_PREFIX + token];
  }

  addEventListener<K extends keyof WebSocketEventMap>(
    type: K,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    listener: (this: WebsocketHandler, ev: WebSocketEventMap[K]) => any,
    options?: boolean | AddEventListenerOptions
  ): void {
    this.socket.addEventListener(type, (ev) => listener.call(this, ev), options);
  }

  removeEventListener<K extends keyof WebSocketEventMap>(
    type: K,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    listener: (this: WebsocketHandler, ev: WebSocketEventMap[K]) => any,
    options?: boolean | EventListenerOptions
  ): void {
    this.socket.addEventListener(type, (ev) => listener.call(this, ev), options);
  }

  send(...actions: Pick<ClientMessage, 'id' | 'action'>[]) {
    const msg = create(ClientMessagesSchema, {
      action: actions.map((a) => create(ClientMessageSchema, a))
    });
    console.log('sending message', msg);
    const data = toBinary(ClientMessagesSchema, msg);

    this.socket.send(data);
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private _received(ev: MessageEvent<any>) {
    console.log('received', ev);

    readBinary(ev.data).then((data) => {
      if (data) {
        const msg = fromBinary(ServerMessagesSchema, data);
        console.log('received proto', msg);

        msg.action.forEach(this._handleMessage.bind(this));
      } else {
        console.log('cant process message from server', ev);
      }
    });
  }

  private _handleMessage(msg: ServerMessage) {
    if (msg.action.case === undefined) {
      return;
    }

    this.events.call(msg.action.case, msg.id, msg.action.value);
  }
}

async function readBinary(data: ArrayBuffer | Blob | string) {
  if (data instanceof ArrayBuffer) {
    return new Uint8Array(data);
  }
  if (data instanceof Blob) {
    return new Uint8Array(await data.bytes());
  }
}
