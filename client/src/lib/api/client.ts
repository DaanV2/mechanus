import type { Interceptor } from '@connectrpc/connect';
import { type Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { getCookie, KEY_ACCESS_TOKEN } from '../cookies';

export const GRPC_PORT = 8666;

// Interceptor to inject access-token as Authorization header
const accessTokenInterceptor: Interceptor = (next) => async (req) => {
  const token = getCookie(KEY_ACCESS_TOKEN);
  if (token) {
    req.header.set('Authorization', `Bearer ${token}`);
  }

  const response = await next(req);
  const s = response.header.get('Set-Cookie');
  if (s) {
    console.log('setting cookie');
    document.cookie = s;
  }

  return response;
};

export function createClient(): Transport {
  return createConnectTransport({
    baseUrl: `${window.location.protocol}//${window.location.hostname}:${GRPC_PORT}`,
    credentials: 'same-origin',
    defaultTimeoutMs: 5000,
    interceptors: [accessTokenInterceptor]
  });
}
