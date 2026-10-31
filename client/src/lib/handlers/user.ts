import type { LoginResponse } from '../../proto/users/v1/login_pb';
import { parseJwt, type JWTClaims } from '../authentication/jwt/parse';
import { Cookie } from '../storage';
import { createLoginClient, createUserClient } from '../stores/clients';

export type UserState = { loggedin: false; data?: undefined } | { loggedin: true; data: JWTClaims };

export namespace UserState {
  export const LOGGED_OUT: UserState = { loggedin: false };
}

function getCurrentUser(): UserState {
  const jwt = Cookie.get('access-token');
  if (!jwt) {
    return {
      loggedin: false
    };
  }

  return parseToken(jwt);
}

function updateCurrentUser(data: LoginResponse): UserState {
  if (data === undefined || data.token === '') return { loggedin: false };

  Cookie.set('access-token', `${data.type} ${data.token}`);

  return parseToken(data.token);
}

function parseToken(jwt: string): UserState {
  try {
    return {
      loggedin: true,
      data: parseJwt(jwt)
    };
  } catch (err) {
    console.log('something is weird with the jwt', jwt, err);
  }

  return {
    loggedin: false
  };
}

export class UserHandler {
  constructor() {}

  get current() {
    return getCurrentUser();
  }

  async create(username: string, password: string) {
    await createUserClient().create({ username, password });

    return this.login(username, password);
  }

  async login(username: string, password: string) {
    const login = await createLoginClient().login({ username, password });

    return updateCurrentUser(login);
  }

  async logout() {
    // TODO;
  }

  async serverData(id?: string) {
    return createUserClient().get({ id: id ?? this.current.data?.user.id });
  }

  /** Whether the logged in user has the admin role */
  get isAdmin(): boolean {
    return this.current.data?.user.roles?.includes('admin') ?? false;
  }

  /** Lists all users, requires the admin role */
  async list() {
    const resp = await createUserClient().list({});

    return resp.users;
  }

  /** Resets the password of the given user, requires the admin role */
  async resetPassword(id: string, newPassword: string) {
    await createUserClient().resetPassword({ id, newPassword });
  }
}

export const userHandler = new UserHandler();
