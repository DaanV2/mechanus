import { createClient } from '$lib/api/client';
import * as users from '$lib/api/users_v1';
import * as screens from '$lib/api/screens_v1';
import type { Transport } from '@connectrpc/connect';

let trans: Transport | undefined = undefined;
export function grpcTransport(): Transport {
  if (!trans) trans = createClient();

  return trans;
}

export const createUserClient = () => users.createUserClient(grpcTransport());
export const createLoginClient = () => users.createLoginClient(grpcTransport());
export const createScreenClient = ()=> screens.createScreenClient(grpcTransport());