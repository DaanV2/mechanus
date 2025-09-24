# Authentication

## Users

When users are logged in, JWT with an coupled JTI are created. These should be valid for an hour and can be used to refresh for a new one.

See [JWT service](server\pkg\authentication\jwt_service.go)

### JWT Fields

Inside of a JWT there are the following fields contains the given information:

```json
{
    // JWT fields: https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.1
    "iss": "mechanus", // Issuer field
    "jti": "<jti>", // A user <-> id that can be revoked to cancel the tokens generated
    "aud": "mechanus", // Audience field
    "exp": 0, // Expiration time
    "iat": 0, // Time the jwt was issued at
    "nbf": 0, // Not before time.
    "sub": "<subject>" // Subject of the jwt, in this case the same as user.id
    // Custom fields
    "user": {
        "id": "<user id>", // Unique user id
        "name": "<name>", // User identifying nam
        "roles": ["<role id>"], // Assigned roles by the server
        "campaigns": ["<campaign id>"]
    },
    "scope": "refresh" | "password"
}
```

## Devices

Devices are not allowed to use JWT, but will receive support for API keys and basic auth for in the future

## Roles

The server knowns some pre built in roles that help manage permissions:

See [Roles](server\pkg\authentication\roles\role.go)
