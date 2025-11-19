import { server_grpc_url } from '$lib/config';
import { type Interceptor, type Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { Cookie } from '../storage';

// Interceptor to inject access-token as Authorization header
const tokenInjector: Interceptor = (next) => (req) => {
  // Get all cookies as a string
  const token = Cookie.get('access-token');
  if (token && token.length > 0) {
    req.header.set('Authorization', token);
  }
  return next(req);
};

export function createClient(): Transport {
  return createConnectTransport({
    baseUrl: server_grpc_url(),
    defaultTimeoutMs: 5000,
    interceptors: [tokenInjector]
  });
}
