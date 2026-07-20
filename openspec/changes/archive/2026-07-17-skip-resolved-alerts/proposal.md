## Why

Penda reruns native package commands for alerts already fixed in the working tree. Repeated runs should recognize the uncommitted manifest changes made by a prior run and avoid duplicate work.

## What Changes

- Inspect Git diffs for each alert's manifest before invoking a provider.
- Skip an alert when its package and patched version already appear in that manifest's added diff lines.
- Report skipped updates in the final run summary.

## Capabilities

### New Capabilities
- `resolved-alert-detection`: Detect uncommitted manifest fixes for Dependabot alerts and skip redundant provider operations.

### Modified Capabilities

None.

## Impact

- Adds Git diff inspection to provider planning and final summaries.
- Preserves native update behavior for alerts without matching manifest changes.
