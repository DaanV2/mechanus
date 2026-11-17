# Spec Delta

## ADDED Requirements

### Requirement: Characters
The system SHALL let campaign members create characters in a campaign, each owned by one player.

#### Scenario: Create a character
- **WHEN** a player creates a character in a campaign
- **THEN** it is listed in the campaign with that player as owner

### Requirement: Characters on scene entities
The system SHALL let a scene entity reference a character, making that character's player its default controller.

#### Scenario: Linked entity
- **WHEN** an entity references a player's character
- **THEN** that player can move the entity's token
