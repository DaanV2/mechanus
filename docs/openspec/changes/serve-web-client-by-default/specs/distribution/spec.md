# Spec Delta

## ADDED Requirements

### Requirement: Web client served by default
The server SHALL serve the web client when started from a release without `--server.static.folder`; the flag SHALL still override the location.

#### Scenario: Run from archive
- **WHEN** a user runs the binary from an unpacked release archive with no flags
- **THEN** opening the server in a browser shows the client
