export const GRPC_PORT = 8666;

export function server_grpc_url(): string {
  return `${window.location.protocol}//${window.location.hostname}:${GRPC_PORT}`;
}

export function server_websocket_url(): string {
  const proc = window.location.protocol === 'https:' ? 'wss' : 'ws';

  return `${proc}//${window.location.hostname}:${GRPC_PORT}`;
}
