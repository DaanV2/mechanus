export interface User {
  name: string;
  password: string;
}

export namespace User {
  export function createRandom(): User {
    const id = crypto.randomUUID();

    return {
      name: "user-" + id,
      password: "password-" + id,
    };
  }

  export function createAdmin(): User {
    return {
      name: "admin",
      password:  "admin",
    };
  }
}
