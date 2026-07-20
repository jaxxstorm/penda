## Purpose

Apply safe, targeted npm dependency updates from Dependabot alerts.

## Requirements

### Requirement: npm alert updates
The npm provider SHALL process npm alerts that target a `package.json` manifest within the selected target directory and include a first patched version.

#### Scenario: Production npm dependency has a patch
- **WHEN** an npm alert targets `package.json` for package `example` with first patched version `1.2.3`
- **THEN** the provider SHALL invoke npm from the manifest directory to install only `example` at `1.2.3`

### Requirement: Development npm dependency updates
The npm provider SHALL use npm's development dependency option when an npm alert identifies a development-scoped dependency.

#### Scenario: Development npm dependency has a patch
- **WHEN** an npm alert for a development-scoped dependency targets `package.json`
- **THEN** the provider SHALL invoke npm to save the patched package as a development dependency

### Requirement: npm update safety
The npm provider SHALL NOT invoke `npm update` or modify `package.json` or lockfiles for alerts without a first patched version or for non-npm ecosystems.

#### Scenario: npm alert has no patched version
- **WHEN** an npm alert does not include a first patched version
- **THEN** the provider SHALL leave `package.json` and lockfiles unchanged
