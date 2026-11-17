# Spec Delta

## Purpose

Keeps one continuous play session per campaign so the game resumes exactly where it left off across restarts and campaign switches.

## ADDED Requirements

### Requirement: One session per campaign
The system SHALL keep one continuous session per campaign, without a start or end, holding the active scene and play state.

#### Scenario: Switch campaign
- **WHEN** the GM switches from campaign A to campaign B and back
- **THEN** campaign A resumes with the scene and state it had

### Requirement: Active scene survives restart
The system SHALL restore the current campaign's active scene and its state after a server restart.

#### Scenario: Restart mid-scene
- **WHEN** the server restarts while a scene is active
- **THEN** the same scene is active again with its state and devices are re-instructed
