# Proposal

## Why

Milestone M2 (Scene authoring). Scene assets (maps, images, videos, tokens) have to be copied into folders by hand. Authoring in the app needs upload and management.

## What Changes

- Upload assets into a scene folder or a shared folder
- List, rename and delete assets
- Uploaded assets are served through the existing content-hashed asset URLs

## Capabilities

### New Capabilities

- `asset-management`: uploading and managing scene and shared assets

### Modified Capabilities

None.

## Impact

- New asset RPCs / upload endpoint
- GM client asset browser
- `infrastructure/assets`
