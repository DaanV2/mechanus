# Proposal

## Why

Milestone M6 (Polish & release). The server already announces itself over mDNS, but clients still need its address typed in.

## What Changes

- Clients find the server on the LAN via mDNS

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `distribution`: adds client-side discovery

## Impact

- client startup / connection flow; browsers cannot browse mDNS directly, so this may need a helper or native shell
