/**
 *
 * @param timeout The amount of milliseconds to sleep
 * @returns
 */
export function sleep(timeout: number) {
  return new Promise((resolve) => setTimeout(resolve, timeout));
}
