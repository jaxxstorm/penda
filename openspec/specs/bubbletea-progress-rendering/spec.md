## Purpose

Render non-interactive and live terminal progress safely using Bubble Tea components.

## Requirements

### Requirement: Bubble Tea progress rendering
Penda SHALL render planned update progress with the Charmbracelet Bubbles progress component without launching an interactive Bubble Tea program.

#### Scenario: Update begins
- **WHEN** Penda begins a planned dependency update
- **THEN** it SHALL render a Bubble Tea progress bar with the current planned update number, total planned updates, and remaining update count

### Requirement: Update identity display
Penda SHALL include the active alert's ecosystem, package name, first patched version, and manifest path with its progress output.

#### Scenario: npm update is applied
- **WHEN** Penda applies an npm update for a package-lock manifest
- **THEN** the progress output SHALL identify npm, the package, target version, and manifest path

### Requirement: Lifecycle status cards
Penda SHALL render a consistent status card for repository discovery, alert retrieval, update planning, completion, and failures.

#### Scenario: Alerts are planned
- **WHEN** Penda has retrieved open Dependabot alerts
- **THEN** it SHALL render a planning status containing both the open alert count and the planned unique update count

### Requirement: Non-interactive output
Penda SHALL NOT enter an alternate screen, wait for user input, or require an interactive terminal to complete. Its progress output SHALL remain append-only and include plain current, total, and remaining text when terminal styling is unavailable.

#### Scenario: CI run
- **WHEN** Penda runs without an interactive terminal
- **THEN** it SHALL complete using ordinary output and process exit statuses without requiring input

### Requirement: Live terminal ticket
When stdout is an interactive terminal, Penda SHALL redraw one progress ticket in place as its lifecycle and update state changes. It SHALL NOT enter an alternate screen or accept keyboard or mouse input.

#### Scenario: Multiple updates are applied in a terminal
- **WHEN** Penda applies multiple planned updates to an interactive terminal
- **THEN** it SHALL update the existing progress ticket rather than append a separate ticket for every update

### Requirement: Final run summary
Penda SHALL render a final summary containing open alerts, planned updates, completed updates, failed updates, and unattempted updates. For each completed update, it SHALL include the ecosystem, package, patched version, native update method, and manifest path.

#### Scenario: One update fails after others complete
- **WHEN** Penda completes some planned updates and one update fails
- **THEN** its final summary SHALL show the completed and failed update counts with the bounded failure diagnostic

#### Scenario: npm updates complete
- **WHEN** Penda completes an npm development dependency update
- **THEN** its final summary SHALL identify the package, patched version, manifest, and `npm install --save-dev` as the update method

### Requirement: Safe failure cards
Penda SHALL display failure cards with bounded native command diagnostics and SHALL NOT include configured credentials in the rendered output.

#### Scenario: Native package command fails
- **WHEN** a native package command fails
- **THEN** Penda SHALL render the command, working directory, and bounded diagnostic output without rendering the configured GitHub token
