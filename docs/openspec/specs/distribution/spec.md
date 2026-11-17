# distribution Specification

## Purpose

Defines how Mechanus is shipped and found on a home network: release binaries and images, LAN announcement, storage backends and the about page.

## Requirements

### Requirement: Supported platforms
The project SHALL release a single server binary for Linux, Windows and macOS on amd64 and arm64, and a multi-arch Docker image, each built and smoke-tested in CI.

#### Scenario: Release
- **WHEN** a release is tagged
- **THEN** archives for all six OS / arch pairs and a multi-arch image are published

### Requirement: Release archives ship the web client
Release archives SHALL contain the web client as a `web/` folder next to the binary.

#### Scenario: Unpack an archive
- **WHEN** a user unpacks a release archive
- **THEN** it contains the binary and a `web/` folder with the client

### Requirement: Database backends
The server SHALL run on SQLite, Postgres, MySQL or an in-memory database, chosen by configuration, and SHALL apply migrations on start.

#### Scenario: First start on Postgres
- **WHEN** the server starts against an empty Postgres database
- **THEN** it creates the schema and starts normally

### Requirement: LAN announcement
The server SHALL announce itself on the local network over mDNS.

#### Scenario: Server starts
- **WHEN** the server starts on a LAN
- **THEN** an mDNS browser on the same network sees the Mechanus service

### Requirement: About page with licences
The client SHALL show an about page listing the licences of third-party Go modules and bundled npm packages, generated at build time.

#### Scenario: Open about
- **WHEN** a user opens the about page
- **THEN** the third-party licences are listed
