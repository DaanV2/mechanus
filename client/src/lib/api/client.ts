import { type Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import {} from 'svelte/reactivity/window';

export const GRPC_PORT = 8666;

export function createClient(): Transport {
  return createConnectTransport({
    baseUrl: `${window.location.protocol}//${window.location.hostname}:${GRPC_PORT}`,
    credentials: 'same-origin',
    defaultTimeoutMs: 5000
  });
}
