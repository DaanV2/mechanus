# Spec Delta

## ADDED Requirements

### Requirement: Scene state per campaign
The system SHALL store scene state per campaign + scene, keeping only the latest state; scenes themselves SHALL stay campaign-independent.

#### Scenario: Same scene, two campaigns
- **WHEN** two campaigns use the `tavern` scene
- **THEN** token moves in one campaign do not appear in the other
