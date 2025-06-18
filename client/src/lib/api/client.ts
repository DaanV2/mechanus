import { type Interceptor, type Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { Cookie } from '../storage';

export const GRPC_PORT = 8666;

// Interceptor to inject access-token as Authorization header
const tokenInjector: Interceptor = (next) => (req) => {
  // Get all cookies as a string
  const token = Cookie.get('access-token');
  if (token && token.length > 0) {
    req.header.append('Authorization', token);
  }
  return next(req);
};

export function createClient(): Transport {
  return createConnectTransport({
    baseUrl: `${window.location.protocol}//${window.location.hostname}:${GRPC_PORT}`,
    credentials: 'same-origin',
    defaultTimeoutMs: 5000,
    interceptors: [tokenInjector]
  });
}
