import { expect, type APIRequestContext } from "@playwright/test";
import { User } from "../users/create";

/**
 * Calls a ConnectRPC method using the JSON protocol
 * @param request The playwright request context
 * @param method The full method path, eg: "users.v1.UserService/Create"
 * @param body The request message, as JSON
 * @param token The optional JWT to authenticate with
 */
export async function callRPC<T>(
  request: APIRequestContext,
  method: string,
  body: object,
  token?: string,
): Promise<T> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const response = await request.post(`/${method}`, { headers, data: body });
  expect(response.ok(), `${method}: ${await response.text()}`).toBeTruthy();

  return (await response.json()) as T;
}

/**
 * Creates the user account and logs in, returning the JWT
 */
export async function createUserAndLogin(
  request: APIRequestContext,
  user: User,
): Promise<string> {
  await callRPC(request, "users.v1.UserService/Create", {
    username: user.name,
    password: user.password,
  });

  return login(request, user);
}

/**
 * Logs in the user, returning the JWT
 */
export async function login(
  request: APIRequestContext,
  user: User,
): Promise<string> {
  const resp = await callRPC<{ token: string }>(
    request,
    "users.v1.LoginService/Login",
    { username: user.name, password: user.password },
  );

  return resp.token;
}
