# user-auth Specification

## Purpose

Lets people sign in to Mechanus with a username and password, keeps them signed in safely across token expiry, and lets them end their own sessions.

## Requirements

### Requirement: Password login creates a session
The system SHALL authenticate a user by username and password and, on success, create a session identified by its own JTI and return an RS512-signed access token and a refresh token (`scope: "refresh"`).

#### Scenario: Successful login
- **WHEN** a user submits a correct username and password
- **THEN** the system returns an access token and a refresh token bound to a new session

#### Scenario: Each login is its own session
- **WHEN** the same user logs in from two browsers
- **THEN** each browser holds a separate session that can be revoked on its own

### Requirement: Token lifetimes are configurable
The system SHALL issue access tokens valid for 1 hour and refresh tokens valid for 7 days by default, and SHALL allow both lifetimes to be changed through configuration.

#### Scenario: Default lifetimes
- **WHEN** a user logs in with default configuration
- **THEN** the access token expires after 1 hour and the refresh token after 7 days

### Requirement: Refresh rotates tokens
The system SHALL exchange a valid refresh token for a new access and refresh token pair and SHALL reject a refresh token whose session has been revoked.

#### Scenario: Refresh before expiry
- **WHEN** the client presents a valid refresh token
- **THEN** the system returns a new token pair for the same session

#### Scenario: Revoked session cannot refresh
- **WHEN** the client presents a refresh token of a logged-out session
- **THEN** the system rejects it as unauthenticated

### Requirement: Client keeps the user signed in
The client SHALL check token expiry and refresh before calling the API, and SHALL NOT log the user out when a refresh fails only because of a network error.

#### Scenario: Session survives access token expiry
- **WHEN** the access token has expired but the refresh token is still valid
- **THEN** the client refreshes and the user stays signed in

#### Scenario: Network failure during refresh
- **WHEN** a refresh request fails because the server is unreachable
- **THEN** the user remains signed in and the refresh is retried later

### Requirement: Logout revokes the session
The system SHALL revoke the current session on logout so that its access and refresh tokens are no longer accepted.

#### Scenario: Logout
- **WHEN** a signed-in user logs out
- **THEN** further requests with that session's tokens are rejected

### Requirement: Unauthenticated and forbidden are distinct
The system SHALL answer `Unauthenticated` when a request has no valid credentials and `PermissionDenied` when valid credentials lack the required role.

#### Scenario: Missing token
- **WHEN** an anonymous request calls a user-only endpoint
- **THEN** the system responds `Unauthenticated`

#### Scenario: Missing role
- **WHEN** a non-admin user calls an admin-only endpoint
- **THEN** the system responds `PermissionDenied`

### Requirement: Brute-force protection
The system SHALL lock out repeated failed logins and SHALL reject an unknown username with the same response and comparable timing as a wrong password.

#### Scenario: Unknown user timing
- **WHEN** a login is attempted for a username that does not exist
- **THEN** the response matches a wrong-password response in content and is comparable in timing

#### Scenario: Repeated failures
- **WHEN** a login fails too many times in a row
- **THEN** further attempts are refused for the lockout period

### Requirement: Self-service password change
The system SHALL let a signed-in user change their own password after confirming the current one.

#### Scenario: Change password
- **WHEN** a user submits their current password and a new password
- **THEN** the new password is required at the next login

### Requirement: Open sign-up is opt-in
The system SHALL reject self sign-up without an invite unless `auth.signup-enabled` is set, which is off by default.

#### Scenario: Sign-up disabled
- **WHEN** someone signs up without an invite and `auth.signup-enabled` is off
- **THEN** the system refuses to create the account

### Requirement: Token cookies
The client SHALL store tokens in JS-readable `SameSite=Lax` cookies with a `Max-Age` matching the token lifetime, and SHALL mark them `Secure` only when served over HTTPS.

#### Scenario: Plain HTTP
- **WHEN** Mechanus is served over plain HTTP
- **THEN** token cookies are set without the `Secure` flag

### Requirement: Signing key rotation
The system SHALL support rotating the RSA signing key.

#### Scenario: Key rotated
- **WHEN** the signing key is rotated
- **THEN** new tokens are signed with the new key

### Requirement: Transport options
The system SHALL serve plain HTTP by default, SHALL support optional TLS through configuration, and SHALL honour `X-Forwarded-Proto`, `X-Forwarded-Host` and `X-Forwarded-For` only when `server.trust-proxy-headers` is enabled.

#### Scenario: Behind a reverse proxy
- **WHEN** `server.trust-proxy-headers` is on and a TLS-terminating proxy forwards `X-Forwarded-Proto: https`
- **THEN** the system treats the request as HTTPS, including setting `Secure` cookies

#### Scenario: Untrusted headers
- **WHEN** `server.trust-proxy-headers` is off
- **THEN** forwarded headers are ignored
