import { server_grpc_url } from '$lib/config';
import { LocalStorage } from '$lib/storage';
import { Code, ConnectError } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import type { DeviceTokens } from '../../proto/devices/v1/devices_pb';
import { createDeviceClient } from '../api/devices_v1';
import { isExpired, parseJwt } from '../authentication/jwt/parse';
import { deviceId, type HardwareType } from './identity';

/** Refresh the access token when it expires within this many seconds */
const REFRESH_MARGIN_SECONDS = 60;
const REFRESH_TIMEOUT_MS = 5000;

/** Answers to a refresh that mean the device is no longer paired */
const REJECTED_CODES = [Code.Unauthenticated, Code.PermissionDenied, Code.InvalidArgument];

function tryExp(token: string | null): { exp: number } | undefined {
  if (!token) return undefined;

  try {
    return parseJwt(token);
  } catch (err) {
    console.error('could not parse the stored device token', err);
  }

  return undefined;
}

/**
 * Device client that sends no credentials: the routes it calls need none, and a stale cookie of a user on this browser
 * would get them rejected, as an invalid credential is never treated as no credential.
 * Has no default timeout, the pairing stream stays open as long as the code is shown.
 */
function deviceClient() {
  return createDeviceClient(
    createConnectTransport({
      baseUrl: server_grpc_url(),
      fetch: (input, init) => fetch(input, { ...init, credentials: 'omit' })
    })
  );
}

let pendingRefresh: Promise<boolean> | undefined;

/** Keeps the device tokens of this browser, obtained by pairing. They outlive the login of any user on this browser */
export namespace DeviceCredentials {
  /** Starts pairing: streams the codes to show, ending with the device tokens */
  export function pair(hardware: HardwareType, signal: AbortSignal) {
    return deviceClient().pair(
      {},
      { signal, headers: { 'X-Device-ID': deviceId(), 'X-Device-Type': hardware } }
    );
  }

  export function store(tokens: DeviceTokens) {
    LocalStorage.set('device-token', tokens.token);
    LocalStorage.set('device-refresh-token', tokens.refreshToken);
  }

  /** The stored access token without prefix, or null */
  export function accessToken(): string | null {
    return LocalStorage.get('device-token');
  }

  /** Whether this browser holds a refresh token that hasn't expired, so as far as it knows the device is paired */
  export function canRefresh(): boolean {
    const claims = tryExp(LocalStorage.get('device-refresh-token'));

    return claims !== undefined && !isExpired(claims);
  }

  export function clear() {
    LocalStorage.remove('device-token');
    LocalStorage.remove('device-refresh-token');
  }

  /**
   * Exchanges the refresh token for new tokens. Concurrent calls share one request.
   * Returns false when that didn't work, and forgets the tokens when the server refused them: the device got revoked
   * and has to pair again. A network error says nothing about that, and keeps them.
   */
  export function refresh(): Promise<boolean> {
    pendingRefresh ??= doRefresh().finally(() => (pendingRefresh = undefined));

    return pendingRefresh;
  }

  async function doRefresh(): Promise<boolean> {
    const token = LocalStorage.get('device-refresh-token');
    if (!token || !canRefresh()) {
      clear();

      return false;
    }

    try {
      const resp = await deviceClient().refresh({ token }, { timeoutMs: REFRESH_TIMEOUT_MS });
      if (!resp.tokens) return false;
      store(resp.tokens);

      return true;
    } catch (err) {
      console.error('could not refresh the device token', err);
      if (err instanceof ConnectError && REJECTED_CODES.includes(err.code)) clear();

      return false;
    }
  }

  /** Refreshes the access token when it is missing or about to expire */
  export async function ensureFresh(): Promise<void> {
    const claims = tryExp(accessToken());
    if (claims && !isExpired(claims, REFRESH_MARGIN_SECONDS)) return;

    await refresh();
  }
}
