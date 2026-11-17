# Tasks

## 1. Role

- [ ] 1.1 Resolve the open questions in design.md (per-campaign vs global GM, screen ownership) and update the specs
- [ ] 1.2 Add the `gm` role and admin UI to grant it; verify with tests and a Playwright role-management test
- [ ] 1.3 Add a migration or upgrade note for existing users who run scenes; verify on a copy of an existing database

## 2. Permissions

- [ ] 2.1 Require GM or admin for scene control RPCs; verify `PermissionDenied` for plain users with tests
- [ ] 2.2 Let GMs create invites, optionally carrying a campaign; verify the new account joins the campaign with a Playwright test
- [ ] 2.3 Let GMs manage the campaigns they run; verify with tests
- [ ] 2.4 Restrict screen access per the chosen ownership model; verify `403` cases with tests
- [ ] 2.5 Update docs/architecture/authentication.md
- [ ] 2.6 Run `just server lint` and `just server test` and fix every reported issue
- [ ] 2.7 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in the milestone list in docs/openspec/README.md.
