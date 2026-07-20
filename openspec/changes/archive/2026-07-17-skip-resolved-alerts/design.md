## Context

Providers currently rerun every planned update even when a prior Penda run already changed the matching manifest. Penda already relies on Git repositories and alert manifest paths, so Git's working-tree diff can identify unapplied work without persisted state.

## Goals / Non-Goals

**Goals:**

- Skip an alert when its manifest diff adds both the alert package and first patched version.
- Keep detection scoped to the selected repository and alert manifest.
- Display skipped updates separately from completed, failed, and unattempted updates.

**Non-Goals:**

- Detecting fixes that are committed or made outside Git.
- Treating arbitrary manifest modifications as resolved alerts.
- Changing native provider commands for unresolved alerts.

## Decisions

- Providers will obtain a manifest-scoped `git diff -- <path>` through an injectable Git diff operation.
- Detection considers only added diff lines and requires both the package name and first patched version, avoiding skips for unrelated changes.
- A matching diff bypasses the provider command and emits a skipped update event. The application summary includes each skipped package, method, and manifest.

## Risks / Trade-offs

- [Formatting can hide a package or version across separate lines] -> Fall back to a native update when a complete matching added line is unavailable.
- [A user manually adds the same version] -> Treat it as resolved because the desired manifest state is already present.
