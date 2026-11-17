# user-management Specification

## Purpose

Lets admins create and manage accounts and roles, and invite players who then choose their own credentials.

## Requirements

### Requirement: Admin user list
The system SHALL let admins list all user accounts on the `/users` page.

#### Scenario: Admin opens users
- **WHEN** an admin opens `/users`
- **THEN** every account is listed with its roles

#### Scenario: Non-admin is refused
- **WHEN** a user without the admin role requests the user list
- **THEN** the system responds `PermissionDenied`

### Requirement: Admin account management
The system SHALL let admins create accounts, reset a user's password and grant or remove roles.

#### Scenario: Create account
- **WHEN** an admin creates an account with a username and password
- **THEN** the new user can log in with those credentials

#### Scenario: Change roles
- **WHEN** an admin grants a user the admin role
- **THEN** that user can call admin-only endpoints with their next token

### Requirement: Single-use invites
The system SHALL let admins create an invite link, also shown as a QR code, that can be used once and expires after a chosen duration: 7 days by default and at most 30 days.

#### Scenario: Create invite
- **WHEN** an admin creates an invite with the default expiry
- **THEN** the system returns a link valid for 7 days

#### Scenario: Expiry above maximum
- **WHEN** an admin requests an invite valid for more than 30 days
- **THEN** the system rejects the request

### Requirement: Sign up through an invite
The system SHALL let the holder of a valid invite pick their own username and password on `/users/signup?invite=<id>`, creating an account with the `user` role and consuming the invite.

#### Scenario: Use invite
- **WHEN** someone opens a valid invite link and submits credentials
- **THEN** an account with the `user` role is created and the invite can no longer be used

#### Scenario: Expired or used invite
- **WHEN** someone opens an expired or already used invite
- **THEN** sign-up is refused
