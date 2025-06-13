import type { Client } from '@connectrpc/connect';
import { parseJwt, type JWTClaims } from '../authenication/jwt/parse';
import { KEY_ACCESS_TOKEN } from '../cookies';
import { getCookie } from '../cookies/cookies';
import { createUserClient } from '../stores/clients';
import type { UserService } from '../../proto/users/v1/users_connect';

export class UserHandler {
  private _current: JWTClaims | undefined;
  private _client: Client<typeof UserService>;

  constructor() {
    this._current = getCurrent();
    this._client = createUserClient();
  }

  get hasLoggedinUser(): boolean {
    return this._current !== undefined;
  }

  data(): JWTClaims | undefined {
    return this._current;
  }

  async serverData() {
    const d = this.data();
    if (d === undefined) return Promise.reject('not logged in');
    const id = d.user.id;

    return this._client.get({ id: id });
  }
}

function getCurrent(): JWTClaims | undefined {
  const jwt = getCookie(KEY_ACCESS_TOKEN);
  if (jwt === undefined) {
    return undefined;
  }

  try {
    return parseJwt(jwt);
  } catch (err) {
    console.log('something is weird with the jwt', jwt, err);
  }

  return undefined;
}
