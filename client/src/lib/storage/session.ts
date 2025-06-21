type sessionStorageKeys = 'access-token' | 'refresh-token';

export namespace SessionStorage {
  export function get(key: sessionStorageKeys): string | null {
    return sessionStorage.getItem(key);
  }
  export function set(key: sessionStorageKeys, value: string) {
    return sessionStorage.setItem(key, value);
  }
  export function remove(key: sessionStorageKeys) {
    return sessionStorage.removeItem(key);
  }
  export function getItem<T>(key: sessionStorageKeys): T | null {
    const item = get(key);
    if (!item) return null;

    try {
      return JSON.parse(item) as T;
    } catch (e) {
      console.error("couldn't read local storage: " + key, e);
    }

    return null;
  }

  export function setItem<T>(key: sessionStorageKeys, value: T) {
    return set(key, JSON.stringify(value));
  }
}
