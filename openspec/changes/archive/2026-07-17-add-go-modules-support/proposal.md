## Why

Penda supports several Dependabot ecosystems but cannot remediate alerts in Go modules, leaving Go repositories dependent on manual updates. Adding Go module support lets the CLI apply the same targeted, locally reviewable fixes for `go.mod` alerts.

## What Changes

- Register a Go module update provider in the default provider runtime.
- Process Dependabot alerts for `go.mod` manifests that include a first patched version.
- Run a targeted Go module command for the alerted module without performing broad module upgrades.
- Skip unsupported Go alerts, missing manifests, and alerts without a patched version without modifying the repository.

## Capabilities

### New Capabilities
- `go-module-dependency-updates`: Apply targeted Dependabot fixes for dependencies declared in Go module manifests.

### Modified Capabilities

None.

## Impact

- Extends the provider registry and update execution pipeline with Go module handling.
- Adds native `go` command execution, manifest validation, and provider fixture tests.
- Updates user-facing documentation to list Go module support and its native tool requirement.
