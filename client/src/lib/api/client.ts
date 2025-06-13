import type { Interceptor } from '@connectrpc/connect';
import { type Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

export const GRPC_PORT = 8666;

// Interceptor to inject access-token as Authorization header
const cookInjector: Interceptor = (next) => (req) => {
  // Get all cookies as a string
  const cookies = document.cookie;
  if (cookies && cookies.length > 0) {
    console.log('injecting cookies');
    req.header.append('Cookie', cookies);
    const h = Array.from(req.header.values());
    console.log(h);
  }
  return next(req);
};

export function createClient(): Transport {
  return createConnectTransport({
    baseUrl: `${window.location.protocol}//${window.location.hostname}:${GRPC_PORT}`,
    credentials: 'same-origin',
    defaultTimeoutMs: 5000,
    interceptors: [cookInjector]
  });
}
