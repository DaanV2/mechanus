# Proposal

## Why

Cross-cutting: unblocks the M0 screen access item and revisits M5 campaign decisions. Several decisions were explicitly made "until a GM role exists": any logged-in user may run scenes and move tokens, only admins create invites, only the campaign owner and admins manage a campaign, and users may open every screen. A GM role lets those rules tighten without making every GM an admin.

## What Changes

- New `gm` role, granted by admins
- Scene control (activate, switch variant, reset, entity controls) limited to GMs and admins
- GMs can create invites; an invite can carry a campaign so the new account joins it on sign-up
- GMs manage the campaigns they run, alongside the owner and admins
- Per-screen access: screens are owned by the room / GM, and users only open screens they are allowed to see

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `user-management`: adds the GM role and GM-created invites that can carry a campaign
- `scene-control`: scene control requires the GM or admin role
- `screen-connection`: users only open screens they are allowed to see
- `campaigns`: GMs manage the campaigns they run

## Impact

- Roles and policy checks in every scene, invite and campaign handler
- GM view and admin role management UI
- Existing users who run scenes today need the `gm` role after the upgrade
