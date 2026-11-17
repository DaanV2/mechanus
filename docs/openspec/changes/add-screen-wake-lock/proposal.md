# Proposal

## Why

Milestone M0 (Foundation). TVs and tablets showing a scene must not dim or sleep. The Screen Wake Lock API only works in a secure context, and Mechanus serves plain HTTP by default, so screens need a fallback. Neither is implemented today.

## What Changes

- The device view keeps the screen awake while it is open
- Uses the Screen Wake Lock API over HTTPS, and a fallback (for example a hidden looping video) over plain HTTP
- Re-acquires the lock when the page becomes visible again

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `screen-views`: the device view keeps the screen awake

## Impact

- `client/src` device view only
