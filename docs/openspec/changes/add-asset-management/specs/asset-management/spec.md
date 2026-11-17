# Spec Delta

## Purpose

Lets the GM upload, list and remove the image, video and token assets that scenes and battle maps use.

## ADDED Requirements

### Requirement: Upload assets
The system SHALL let the GM upload files into a scene's folder or a shared folder.

#### Scenario: Upload a map image
- **WHEN** the GM uploads `cave.webp` to the `cave` scene
- **THEN** it is available as `maps/cave.webp` to that scene

### Requirement: Manage assets
The system SHALL let the GM list, rename and delete uploaded assets.

#### Scenario: Delete an asset
- **WHEN** the GM deletes an unused shared asset
- **THEN** it is no longer listed or served
