import type { Interceptor } from '@connectrpc/connect';
import { type Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

export const GRPC_PORT = 8666;

// Helper to get a cookie value by name
function getCookie(name: string): string | undefined {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
  return match ? decodeURIComponent(match[2]) : undefined;
}

// Interceptor to inject access-token as Authorization header
const accessTokenInterceptor: Interceptor = (next) => async (req) => {
  const token = getCookie('access-token');
  if (token) {
    req.header.set('Authorization', `Bearer ${token}`);
  }
  return next(req);
};

export function createClient(): Transport {
  return createConnectTransport({
    baseUrl: `${window.location.protocol}//${window.location.hostname}:${GRPC_PORT}`,
    credentials: 'same-origin',
    defaultTimeoutMs: 5000,
    interceptors: [accessTokenInterceptor]
  });
}
