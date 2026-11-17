# Spec Delta

## ADDED Requirements

### Requirement: GM role
The system SHALL support a `gm` role that admins can grant and remove; admins SHALL keep every permission a GM has.

#### Scenario: Grant GM
- **WHEN** an admin grants a user the `gm` role
- **THEN** that user can control scenes with their next token

### Requirement: GM invites carry a campaign
The system SHALL let GMs create invites, which create `user` accounts only, and SHALL let an invite name a campaign the new account joins on sign-up.

#### Scenario: Invite into a campaign
- **WHEN** a GM creates an invite for their campaign and someone signs up with it
- **THEN** the new account has the `user` role and is a player in that campaign
