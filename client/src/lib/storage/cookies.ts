type cookieKeys = 'access-token' | 'refresh-token';

export namespace Cookie {
  export function get(key: cookieKeys): string | null {
    const match = document.cookie.match(new RegExp('(^| )' + key + '=([^;]+)'));

    if (match) {
      console.log('got a cookie', match);
      return decodeURIComponent(match[2]);
    }
    console.log('no match?', document.cookie);

    return null;
  }
  export function set(key: cookieKeys, value: string) {
    const msg = `${key}=${value}`;
    console.log('setting cookie', msg);
    document.cookie = msg;
  }
  export function getItem<T>(key: cookieKeys): T | null {
    const item = get(key);
    if (!item) return null;

    try {
      return JSON.parse(item) as T;
    } catch (e) {
      console.error("couldn't read local storage: " + key, e);
    }

    return null;
  }

  export function setItem<T>(key: cookieKeys, value: T) {
    return set(key, JSON.stringify(value));
  }
}
