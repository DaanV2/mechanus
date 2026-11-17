# Proposal

## Why

Milestone M0 (Foundation). The project rule is that security policies are enforced in the message handler, not an interceptor. `DeviceService` still relies on `grpc.PolicyInterceptor`, and some handlers mix checks with the work they guard.

## What Changes

- Move `DeviceService` authorization off `grpc.PolicyInterceptor` into its handlers
- Split handlers that mix checks and work into a public entry function (`Create(ctx)`) that validates, and a private variant (`create(ctx)`) that does the work
- Decide whether the screen WebSocket's `http.RequirePolicy` moves too
- No observable behaviour change: same endpoints, same `Unauthenticated` / `PermissionDenied` answers

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None.

## Impact

- `server/application` services, `server/infrastructure/grpc`, `components/servers.go`
- docs/architecture/authentication.md
