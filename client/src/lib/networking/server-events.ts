import type { Action } from '@sveltejs/kit';
import type { ServerMessage } from '../../proto/screens/v1/server_pb';

type ActionCases = UnionToMap<
  Exclude<
    ServerMessage['action'],
    {
      case: undefined;
      value?: undefined;
    }
  >,
  false
>;
type EventMapping = UnionToMap<
  Exclude<
    ServerMessage['action'],
    {
      case: undefined;
      value?: undefined;
    }
  >,
  true
>;

type UnionToMap<U extends { case: string; value: any }, WrapValue extends boolean = false> = {
  [K in U as K['case']]: WrapValue extends true ? Array<ActionHandler<K['value']>> : K['value'];
};

type ActionCase = keyof ActionCases;
type ActionType<T extends ActionCase> = ActionCases[T];
type Handler<T> = (id: string | undefined, msg: T) => void;
type ActionHandler<K extends ActionCase> = Handler<ActionCases[K]>;

export class ServerEventHandler {
  readonly events: Partial<Record<ActionCase, Array<ActionType<any>>>>;

  constructor() {
    this.events = {};
  }

  add<K extends ActionCase>(event: K, handler: ActionHandler<K>) {
    const events = this.events[event];
    if (events) {
      events.push(handler);
    } else {
      this.events[event] = [handler];
    }
  }

  call<K extends ActionCase>(event: K, id: string | undefined, msg: ActionType<K>) {
    this.events[event]?.forEach((fn) => fn.call(undefined, id, msg));
  }
}
