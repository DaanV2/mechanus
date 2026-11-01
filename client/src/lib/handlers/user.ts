import { type JWTClaims } from '../authentication/jwt/parse';
import { Session, type TokenResponse } from '../authentication/session';
import { createLoginClient, createUserClient } from '../stores/clients';

export type UserState = { loggedin: false; data?: undefined } | { loggedin: true; data: JWTClaims };

export namespace UserState {
  export const LOGGED_OUT: UserState = { loggedin: false };
}

function getCurrentUser(): UserState {
  // An expired access token doesn't count as logged in, requests refresh it before they are sent
  const claims = Session.claims();
  if (!claims) return UserState.LOGGED_OUT;

  return { loggedin: true, data: claims };
}

function updateCurrentUser(data: TokenResponse): UserState {
  if (data === undefined || data.token === '') return UserState.LOGGED_OUT;

  const claims = Session.store(data);
  if (!claims) return UserState.LOGGED_OUT;

  return { loggedin: true, data: claims };
}

export class UserHandler {
  constructor() {}

  get current() {
    return getCurrentUser();
  }

  /**
   * The current user, refreshing the access token first when it expired but the session can still be refreshed.
   * Use this on page load, {@link current} only looks at the access token.
   */
  async restore(): Promise<UserState> {
    await Session.ensureFresh();

    return getCurrentUser();
  }

  /** Creates an account, then logs in as it */
  async create(username: string, password: string) {
    await createUserClient().create({ username, password });

    return this.login(username, password);
  }

  /**
   * Creates an account for someone else with the given roles, without logging in as it. Requires the admin role,
   * unless open sign-up is enabled on the server
   */
  async createAccount(username: string, password: string, roles: string[]) {
    const resp = await createUserClient().create({ username, password });
    const id = resp.user?.id ?? '';
    // New accounts start with the user role
    if (roles.length > 0 && !(roles.length === 1 && roles[0] === 'user')) {
      await this.setRoles(id, roles);
    }

    return id;
  }

  async login(username: string, password: string) {
    const login = await createLoginClient().login({ username, password });

    return updateCurrentUser(login);
  }

  /** Logs out the current session on the server and forgets its tokens */
  async logout() {
    await Session.logout();
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

  /** Resets the password of the given user, requires the admin role. Signs out all their sessions */
  async resetPassword(id: string, newPassword: string) {
    await createUserClient().resetPassword({ id, newPassword });
  }

  /** Changes the password of the logged in user, signs out all their other sessions */
  async changePassword(currentPassword: string, newPassword: string) {
    await createUserClient().changePassword({ currentPassword, newPassword });
  }

  /** Replaces the roles of the given user, requires the admin role */
  async setRoles(id: string, roles: string[]) {
    const resp = await createUserClient().setRoles({ id, roles });

    return resp.user;
  }
}

export const userHandler = new UserHandler();
