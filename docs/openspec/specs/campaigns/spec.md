# campaigns Specification

## Purpose

Lets users create campaigns, see the campaigns they play in, and lets the campaign owner manage its players.

## Requirements

### Requirement: Campaign overview
The system SHALL show users the campaigns they are a member of and show admins every campaign.

#### Scenario: Player opens campaigns
- **WHEN** a user opens the campaign overview
- **THEN** only campaigns they own or play in are listed

#### Scenario: Admin opens campaigns
- **WHEN** an admin opens the campaign overview
- **THEN** every campaign is listed

### Requirement: Create a campaign
The system SHALL let any user create a campaign, becoming its owner.

#### Scenario: Create
- **WHEN** a user creates a campaign
- **THEN** they are recorded as its owner and see it in their overview

### Requirement: Owner and admin manage a campaign
The system SHALL let only the owner and admins edit or delete a campaign and add or remove players; other members SHALL only see it and leave it. Campaigns without an owner SHALL be managed by admins only.

#### Scenario: Member tries to edit
- **WHEN** a member who is not the owner edits the campaign
- **THEN** the system responds `PermissionDenied`

#### Scenario: Member leaves
- **WHEN** a member leaves a campaign
- **THEN** it no longer appears in their overview

### Requirement: Owner cannot be removed
The system SHALL refuse to remove the owner from their campaign's players.

#### Scenario: Remove owner
- **WHEN** an admin tries to remove the owner from the player list
- **THEN** the request is refused

### Requirement: Add players by username
The system SHALL add players by the username of an existing account; the campaign page SHALL link to the admin invite flow for people without an account.

#### Scenario: Unknown username
- **WHEN** the owner adds a username that does not exist
- **THEN** the system reports that the user was not found

### Requirement: Soft delete
The system SHALL soft delete campaigns, keeping their player links so they can be restored from the database.

#### Scenario: Delete
- **WHEN** the owner deletes a campaign
- **THEN** it disappears from every overview while its data remains in the database
