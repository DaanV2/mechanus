import { ConnectError } from '@connectrpc/connect';

export type MechanusError = ConnectError | Error | string | null;
