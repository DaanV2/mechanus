import { server_grpc_url } from '$lib/config';
import { Code, ConnectError } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { createLoginClient } from '../api/users_v1';
import { Cookie } from '../storage';
import { isExpired, parseJwt, secondsUntilExpiry, type JWTClaims } from './jwt/parse';

/** Refresh the access token when it expires within this many seconds */
const REFRESH_MARGIN_SECONDS = 60;

/** Answers to a refresh that mean the refresh token is no longer accepted */
const REJECTED_CODES = [Code.Unauthenticated, Code.PermissionDenied, Code.InvalidArgument];

export interface TokenResponse {
  /** The access token, without prefix */
  token: string;
  /** The prefix of the token, so far always Bearer */
  type: string;
  /** The refresh token, without prefix */
  refreshToken: string;
}

function tryParse(token: string | null): JWTClaims | undefined {
  if (!token) return undefined;

  try {
    return parseJwt(token);
  } catch (err) {
    console.error('could not parse the stored token', err);
  }

  return undefined;
}

/** Login client that skips the auth interceptor, so refreshing never loops back into itself */
function loginClient() {
  return createLoginClient(
    createConnectTransport({ baseUrl: server_grpc_url(), defaultTimeoutMs: 5000 })
  );
}

let pendingRefresh: Promise<boolean> | undefined;

/** Keeps the access and refresh token of the logged in user in cookies, and refreshes them */
export namespace Session {
  /** Stores the tokens of a login or refresh, the cookies expire together with their tokens */
  export function store(resp: TokenResponse): JWTClaims | undefined {
    const access = tryParse(resp.token);
    if (!access) return undefined;

    Cookie.set('access-token', `${resp.type} ${resp.token}`, {
      maxAge: secondsUntilExpiry(access)
    });

    const refresh = tryParse(resp.refreshToken);
    if (refresh) {
      Cookie.set('refresh-token', resp.refreshToken, { maxAge: secondsUntilExpiry(refresh) });
    }

    return access;
  }

  /** The stored access token with its prefix (eg: `Bearer <jwt>`), or null */
  export function accessToken(): string | null {
    return Cookie.get('access-token');
  }

  /** The stored refresh token without prefix, or null */
  export function refreshToken(): string | null {
    return Cookie.get('refresh-token');
  }

  /** The claims of the stored access token, as long as it hasn't expired */
  export function claims(): JWTClaims | undefined {
    const claims = tryParse(accessToken());
    if (!claims || isExpired(claims)) return undefined;

    return claims;
  }

  /** Whether a refresh token is stored that hasn't expired */
  export function canRefresh(): boolean {
    const claims = tryParse(refreshToken());

    return claims !== undefined && !isExpired(claims);
  }

  export function clear() {
    Cookie.remove('access-token');
    Cookie.remove('refresh-token');
  }

  /**
   * Exchanges the refresh token for new tokens. Concurrent calls share one request.
   * Returns false, and clears the session, when the refresh token is missing or no longer accepted.
   */
  export function refresh(): Promise<boolean> {
    pendingRefresh ??= doRefresh().finally(() => (pendingRefresh = undefined));

    return pendingRefresh;
  }

  async function doRefresh(): Promise<boolean> {
    const token = refreshToken();
    if (!token || !canRefresh()) {
      return false;
    }

    try {
      const resp = await loginClient().refresh({ token });

      return store(resp) !== undefined;
    } catch (err) {
      console.error('could not refresh the session', err);
      // Only forget the tokens when the server refused them. A network error, or a request aborted by leaving
      // the page, says nothing about the session and must not log the user out
      if (err instanceof ConnectError && REJECTED_CODES.includes(err.code)) clear();

      return false;
    }
  }

  /** Refreshes the access token when it is missing or about to expire, and a refresh token is available */
  export async function ensureFresh(): Promise<void> {
    const claims = tryParse(accessToken());
    if (claims && !isExpired(claims, REFRESH_MARGIN_SECONDS)) return;
    if (!canRefresh()) return;

    await refresh();
  }

  /** Logs out the session on the server, and forgets the tokens */
  export async function logout(): Promise<void> {
    const token = refreshToken() ?? accessToken()?.replace(/^Bearer /, '');

    try {
      if (token) await loginClient().logout({ token });
    } catch (err) {
      // The session might already be expired or revoked, forgetting the tokens is what matters
      console.error('could not log out on the server', err);
    } finally {
      clear();
    }
  }
}
