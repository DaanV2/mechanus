import { timestampDate } from '@bufbuild/protobuf/wkt';
import { createInviteClient } from '../stores/clients';
import { userHandler } from './user';

/** The query parameter of the sign-up page that holds the invite ID */
export const INVITE_PARAM = 'invite';

export class InviteHandler {
  /** Creates an invite valid for the given number of seconds (0 uses the server default), requires the admin role */
  async create(validForSeconds = 0) {
    const resp = await createInviteClient().create({ validForSeconds: BigInt(validForSeconds) });

    return resp.invite;
  }

  /** Lists the invites that can still be used, requires the admin role */
  async list() {
    const resp = await createInviteClient().list({});

    return resp.invites;
  }

  /** Revokes an invite so it can't be used anymore, requires the admin role */
  async revoke(id: string) {
    await createInviteClient().revoke({ id });
  }

  /** When the invite can still be used, its expiry, otherwise undefined */
  async check(id: string): Promise<Date | undefined> {
    const resp = await createInviteClient().check({ id });
    if (!resp.valid || !resp.expiresAt) return undefined;

    return timestampDate(resp.expiresAt);
  }

  /** Creates an account with the invite, then logs in as it */
  async accept(id: string, username: string, password: string) {
    await createInviteClient().accept({ id, username, password });

    return userHandler.login(username, password);
  }

  /** The sign-up link of an invite, on the server this page was loaded from */
  link(id: string): string {
    const url = new URL('/users/signup', window.location.origin);
    url.searchParams.set(INVITE_PARAM, id);

    return url.toString();
  }
}

export const inviteHandler = new InviteHandler();
