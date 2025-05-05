import { type Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

export function createClient(): Transport {
  return createConnectTransport({
    baseUrl: `${window.location.protocol}//${window.location.hostname}:8666`,
    credentials: 'same-origin'
  });
}
