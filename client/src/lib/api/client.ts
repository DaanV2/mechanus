import { server_grpc_url } from '$lib/config';
import { Code, ConnectError, type Interceptor, type Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { LoginService } from '../../proto/users/v1/login_pb';
import { Session } from '../authentication/session';

// Interceptor that keeps the access token fresh and injects it as the Authorization header
const tokenInjector: Interceptor = (next) => async (req) => {
  await Session.ensureFresh();

  const send = () => {
    const token = Session.accessToken();
    if (token && token.length > 0) {
      req.header.set('Authorization', token);
    } else {
      req.header.delete('Authorization');
    }

    return next(req);
  };

  try {
    return await send();
  } catch (err) {
    // The access token got rejected (expired / revoked), refresh once and retry
    if (
      err instanceof ConnectError &&
      err.code === Code.Unauthenticated &&
      req.service.typeName !== LoginService.typeName &&
      Session.canRefresh() &&
      (await Session.refresh())
    ) {
      return send();
    }

    throw err;
  }
};

export function createClient(): Transport {
  return createConnectTransport({
    baseUrl: server_grpc_url(),
    defaultTimeoutMs: 5000,
    interceptors: [tokenInjector]
  });
}
