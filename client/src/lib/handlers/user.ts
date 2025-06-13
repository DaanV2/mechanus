import { KEY_ACCESS_TOKEN } from '../cookies';
import { getCookie } from '../cookies/cookies';

export class UserHandler {
  IsLoggedIn(): boolean {
    const c = getCookie(KEY_ACCESS_TOKEN);

    return typeof c === 'string' && c.length > 0;
  }
}
