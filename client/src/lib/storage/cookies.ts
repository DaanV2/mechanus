type cookieKeys = 'access-token' | 'refresh-token';

export interface CookieOptions {
  /** How long the cookie lives, in seconds. Without it the cookie is removed when the browser closes */
  maxAge?: number;
}

export namespace Cookie {
  export function get(key: cookieKeys): string | null {
    const match = document.cookie.match(new RegExp('(^| )' + key + '=([^;]+)'));
    if (!match) return null;

    return decodeURIComponent(match[2]);
  }

  export function set(key: cookieKeys, value: string, options: CookieOptions = {}) {
    const parts = [`${key}=${encodeURIComponent(value)}`, 'path=/', 'SameSite=Lax'];
    if (options.maxAge !== undefined) {
      parts.push(`Max-Age=${Math.max(0, Math.floor(options.maxAge))}`);
    }
    // Only send the token over HTTPS when the page itself was served over HTTPS
    if (location.protocol === 'https:') {
      parts.push('Secure');
    }

    document.cookie = parts.join('; ');
  }

  export function remove(key: cookieKeys) {
    document.cookie = `${key}=; path=/; SameSite=Lax; Max-Age=0`;
  }

  export function getItem<T>(key: cookieKeys): T | null {
    const item = get(key);
    if (!item) return null;

    try {
      return JSON.parse(item) as T;
    } catch (e) {
      console.error("couldn't read cookie: " + key, e);
    }

    return null;
  }

  export function setItem<T>(key: cookieKeys, value: T, options?: CookieOptions) {
    return set(key, JSON.stringify(value), options);
  }
}
