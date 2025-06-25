type localStorageKeys = 'access-token' | 'refresh-token';

export namespace LocalStorage {
  export function get(key: localStorageKeys): string | null {
    return localStorage.getItem(key);
  }
  export function set(key: localStorageKeys, value: string) {
    return localStorage.setItem(key, value);
  }
  export function remove(key: localStorageKeys) {
    return localStorage.removeItem(key);
  }
  export function getItem<T>(key: localStorageKeys): T | null {
    const item = get(key);
    if (!item) return null;

    try {
      return JSON.parse(item) as T;
    } catch (e) {
      console.error("couldn't read local storage: " + key, e);
    }

    return null;
  }

  export function setItem<T>(key: localStorageKeys, value: T) {
    return set(key, JSON.stringify(value));
  }
}
