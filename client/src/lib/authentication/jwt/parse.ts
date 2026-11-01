export interface JWTClaims {
  iss: string;
  sub: string;
  aud: string;
  exp: number;
  nbf: number;
  jti: string;
  user: JWtUser;
  grant: 'password' | 'refresh';
  scope: '*' | 'refresh';
}

export interface JWtUser {
  id: string;
  name: string;
  roles: string[];
  Campaigns: string[];
}

export function parseJwt(token: string): JWTClaims {
  const base64Url = token.replace(/^Bearer /, '').split('.')[1];
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  const jsonPayload = decodeURIComponent(
    window
      .atob(base64)
      .split('')
      .map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
      })
      .join('')
  );

  return JSON.parse(jsonPayload);
}

/** Seconds until the token expires, negative once expired */
export function secondsUntilExpiry(claims: Pick<JWTClaims, 'exp'>): number {
  return claims.exp - Date.now() / 1000;
}

/** Whether the token expired, or will within the given margin in seconds */
export function isExpired(claims: Pick<JWTClaims, 'exp'>, marginSeconds = 0): boolean {
  return typeof claims.exp !== 'number' || secondsUntilExpiry(claims) <= marginSeconds;
}
