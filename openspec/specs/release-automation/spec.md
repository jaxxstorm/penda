## Purpose

Publish tagged Penda releases and package manifests automatically.

## Requirements

### Requirement: Tagged releases
Penda SHALL publish GitHub Releases when a `v*` tag is pushed.

#### Scenario: Version tag is pushed
- **WHEN** a `v1.2.3` tag is pushed
- **THEN** GitHub Actions SHALL run GoReleaser v2 to publish Penda artifacts

### Requirement: Package manifests
GoReleaser SHALL publish a `penda` Homebrew Cask and Scoop manifest to the configured repositories using `HOMEBREW_TOKEN`.

#### Scenario: Release succeeds
- **WHEN** GoReleaser publishes a Penda release
- **THEN** it SHALL update the configured tap Cask and Scoop bucket manifests
