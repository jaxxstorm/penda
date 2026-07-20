## Why

Penda needs repeatable tagged releases and package distribution.

## What Changes

- Add GoReleaser v2 configuration for GitHub Releases, Homebrew Cask, and Scoop.
- Add a GitHub Actions workflow that releases `v*` tags.

## Capabilities

### New Capabilities
- `release-automation`: Publish tagged Penda binaries and package manifests.

### Modified Capabilities

None.

## Impact

- Adds `.goreleaser.yaml` and `.github/workflows/release.yml`.
