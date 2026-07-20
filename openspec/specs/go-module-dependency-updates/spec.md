## Purpose

Support safe, targeted Go module dependency updates from Dependabot alerts.

## Requirements

### Requirement: Go module alert updates
Penda SHALL process a Dependabot alert with the `gomod` package ecosystem when it targets an existing `go.mod` manifest within the selected target directory and includes a package name and first patched version.

#### Scenario: Go module dependency has a patch
- **WHEN** a `gomod` alert targets `go.mod` for module `example.com/module` with first patched version `v1.2.3`
- **THEN** Penda SHALL invoke the Go toolchain from the manifest directory to update only `example.com/module` to `v1.2.3`

### Requirement: Targeted Go module command
For each matching Go module alert, Penda SHALL run `go get <module>@<first-patched-version>` from the directory containing the alerted `go.mod` manifest.

#### Scenario: Nested module manifest has a patch
- **WHEN** a `gomod` alert targets `services/api/go.mod` for module `example.com/module` with first patched version `v1.2.3`
- **THEN** Penda SHALL run `go get example.com/module@v1.2.3` with `services/api` as its working directory

### Requirement: Go module update safety
Penda SHALL NOT invoke a Go command or modify Go module files for a non-`gomod` alert, an alert without a first patched version, an alert whose manifest is not `go.mod`, or an alert whose manifest is absent from the selected checkout.

#### Scenario: Go module alert has no patched version
- **WHEN** a `gomod` alert targeting `go.mod` does not include a first patched version
- **THEN** Penda SHALL leave `go.mod` and `go.sum` unchanged

#### Scenario: Go module manifest is absent locally
- **WHEN** a `gomod` alert names a `go.mod` manifest that is absent from the selected checkout
- **THEN** Penda SHALL not invoke the Go toolchain or fail the overall Penda run for that alert

### Requirement: Go module manifest containment
Penda SHALL reject a `gomod` alert whose `go.mod` manifest path is absolute or resolves outside the selected target directory before invoking the Go toolchain.

#### Scenario: Go module alert specifies a traversal path
- **WHEN** a `gomod` alert names `../go.mod` as its manifest path
- **THEN** Penda SHALL not invoke the Go toolchain or modify a file outside the selected target directory
