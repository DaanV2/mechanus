# Tasks

## 1. DeviceService

- [ ] 1.1 Move each `DeviceService` policy into its handler and remove it from `grpc.PolicyInterceptor`; verify existing device tests still pass
- [ ] 1.2 Add tests that call each `DeviceService` method with no / wrong credentials and assert `Unauthenticated` / `PermissionDenied`

## 2. Other handlers

- [ ] 2.1 Split handlers that mix checks and work into entry + private variants; verify with `just server test`
- [ ] 2.2 Settle the screen WebSocket question and update docs/architecture/authentication.md accordingly
- [ ] 2.3 Run `just server lint` and `just server test` and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in the milestone list in docs/openspec/README.md.
