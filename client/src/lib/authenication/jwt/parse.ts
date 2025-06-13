export interface JWTClaims {
  iss: string;
  sub: string;
  aud: string;
  exp: number;
  nbf: number;
  jti: string;
  user: JWtUser;
  scope: 'password' | 'refresh';
}

export interface JWtUser {
  id: string;
  name: string;
  roles: string[];
  Campaigns: string[];
}

export function parseJwt(token: string): JWTClaims {
  const base64Url = token.split('.')[1];
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
