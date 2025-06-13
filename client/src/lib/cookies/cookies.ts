// Helper to get a cookie value by name
export function getCookie(name: string): string | undefined {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
  return match ? decodeURIComponent(match[2]) : undefined;
}

export function setCookie(value: string) {
  document.cookie = value;
}
