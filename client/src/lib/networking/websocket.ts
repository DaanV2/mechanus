import { server_websocket_url } from '$lib/config';
import { Cookie } from '$lib/storage';
import { create, fromBinary, toBinary } from '@bufbuild/protobuf';
import {
  ClientMessageSchema,
  ClientMessagesSchema,
  type ClientMessage
} from '../../proto/screens/v1/client_pb';
import { ServerMessagesSchema, type ServerMessage } from '../../proto/screens/v1/server_pb';
import { ServerEventHandler } from './server-events';

export class WebsocketHandler {
  readonly screenid: string;
  readonly id: string;
  readonly token?: string;
  readonly socket: WebSocket;
  readonly events: ServerEventHandler;

  constructor(screenid: string, id: string, token?: string) {
    this.events = new ServerEventHandler();
    this.screenid = screenid;
    this.id = id;
    this.token = token;
    const url = server_websocket_url() + `/api/v1/screen/${this.screenid}/${this.id}`;
    this.socket = new WebSocket(url);
    console.log('opening websocket on', url);
    this.socket.addEventListener('message', this._received.bind(this));
  }

  getToken() {
    return this.token ?? Cookie.get('access-token');
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
