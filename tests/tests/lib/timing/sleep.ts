
/**
 * The amount of ms to sleep before continuing
 * @param delay 
 * @returns 
 */
export function sleep(delay: number): Promise<void> {
  return new Promise<void>((resolve) => setTimeout(resolve, delay));
}