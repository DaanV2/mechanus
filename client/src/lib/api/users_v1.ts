import { createClient, type Client, type Transport } from '@connectrpc/connect';
import { UserService } from '../../proto/users/v1/users_pb';
import { LoginService } from '../../proto/users/v1/login_pb';

export function createUserClient(transport: Transport): Client<typeof UserService> {
  return createClient(UserService, transport);
}

export function createLoginClient(transport: Transport): Client<typeof LoginService> {
  return createClient(LoginService, transport);
}
