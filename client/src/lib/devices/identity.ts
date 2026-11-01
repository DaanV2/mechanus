import { LocalStorage } from '$lib/storage';

/** Hardware types the server understands for X-Device-Type / `deviceType`; anything else is stored as `other`. */
export type HardwareType = 'screen' | 'table' | 'tablet' | 'phone' | 'desktop' | 'other';

/**
 * Returns this browser's stable device ID, generating and persisting one on first use.
 * The screen WebSocket uses it as the screen ID, so the room's bindings decide what this device shows.
 */
export function deviceId(): string {
  const existing = LocalStorage.get('device-id');
  if (existing) return existing;

  const id = randomId();
  LocalStorage.set('device-id', id);

  return id;
}

/** A random hex ID. `crypto.randomUUID` needs a secure context, which a LAN server over plain HTTP isn't. */
export function randomId(): string {
  const bytes = crypto.getRandomValues(new Uint8Array(16));

  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
}
