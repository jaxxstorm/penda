## Context

GoReleaser v2.17 documents `homebrew_casks` and `scoops` publishers with repository-scoped tokens.

## Goals / Non-Goals

**Goals:** Publish Penda releases, casks, and Scoop manifests from version tags.

**Non-Goals:** Signing, notarization, or prereleases.

## Decisions

- Build Penda for macOS, Linux, and Windows on amd64 and arm64.
- Use `HOMEBREW_TOKEN` for the external tap and Scoop bucket.
- Trigger releases only for `v*` tags with full Git history.

## Risks / Trade-offs

- [External repository writes require token access] -> The workflow reads `HOMEBREW_TOKEN` only from GitHub Actions secrets.
