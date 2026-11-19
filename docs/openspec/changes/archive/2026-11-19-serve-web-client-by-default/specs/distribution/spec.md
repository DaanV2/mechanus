# Spec Delta

## ADDED Requirements

### Requirement: Web client served by default
The server SHALL serve the web client when started from a release without `--server.static.folder`; the flag SHALL still override the location.

#### Scenario: Run from archive
- **WHEN** a user runs the binary from an unpacked release archive with no flags
- **THEN** opening the server in a browser shows the client

## MODIFIED Requirements

### Requirement: Release archives ship the web client
Release archives SHALL ship the web client inside the server binary, so an archive needs no separate `web/` folder; on start the server writes the client to its static folder.

#### Scenario: Unpack an archive
- **WHEN** a user unpacks a release archive
- **THEN** it contains the binary, which holds the web client, and no `web/` folder
