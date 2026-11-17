# screen-views Specification

## Purpose

Defines what a device screen shows: the full-screen device view driven entirely by the screen state the server pushes, and the view kinds it can render.

## Requirements

### Requirement: Device view
The client SHALL provide a full-screen device view at `/views/devices/` that renders whatever screen state the server sends, reconnects automatically, and accepts `?device=` / `?type=` overrides.

#### Scenario: TV opens the device view
- **WHEN** a paired TV opens `/views/devices/`
- **THEN** it connects with its device credentials and shows the view from its screen state

#### Scenario: Unpaired device
- **WHEN** a device without credentials opens the device view
- **THEN** it is sent to `/devices/pair`

### Requirement: Screen state replaces the screen
The system SHALL send a complete screen state (view, overlays, layers, entities, grid and camera) on initial setup and whenever a scene instruction changes, and the client SHALL replace what it shows with it.

#### Scenario: Variant switch
- **WHEN** a new screen state arrives for a device
- **THEN** the device shows the new state in full, with nothing left over from the old one

### Requirement: View kinds
The client SHALL render the `battlemap` (layers and grid), `scenery` (image or video), `splash` and `blank` views.

#### Scenario: Scenery video
- **WHEN** the screen state names the `scenery` view with a video asset
- **THEN** the device plays the video full screen

### Requirement: Entities overlay on battle maps
The system SHALL support an `entities` overlay on battle map views that lists the names of the visible entities.

#### Scenario: Overlay enabled
- **WHEN** a battle map instruction enables the `entities` overlay
- **THEN** the screen lists the names of the visible entities and no hidden ones

### Requirement: Content-hashed scene assets
The system SHALL serve scene assets under `/assets/scenes/<scene id>/...` and shared assets under `/assets/shared/...` with content-hashed URLs, and screen state SHALL reference assets by those URLs.

#### Scenario: Asset changed on disk
- **WHEN** an asset file's content changes
- **THEN** its URL in new screen states changes so screens do not show a cached old version
