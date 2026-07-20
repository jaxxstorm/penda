## Purpose

Hand off retrieved Dependabot alerts to registered providers safely.

## Requirements

### Requirement: Provider-facing alert contract
Penda SHALL define a provider interface that accepts an execution context and the complete collection of mapped Dependabot alerts.

#### Scenario: Future provider is registered
- **WHEN** Penda has successfully retrieved alerts and one or more providers are registered
- **THEN** Penda SHALL invoke each registered provider with the same complete alert collection

### Requirement: Read-only initial handoff
The initial Penda command SHALL register no providers that modify repository files or invoke package-management tooling.

#### Scenario: Alerts are retrieved
- **WHEN** Penda successfully retrieves Dependabot alerts
- **THEN** Penda SHALL complete without modifying dependency manifests, lockfiles, or other repository files

### Requirement: Retrieval result reporting
Penda SHALL emit a concise terminal status message containing the number of retrieved Dependabot alerts after completing provider handoff.

#### Scenario: No alerts are found
- **WHEN** Penda retrieves an empty alert collection
- **THEN** Penda SHALL report that zero Dependabot alerts were retrieved and exit successfully

### Requirement: Provider failure handling
Penda SHALL stop provider processing and exit with a non-zero status when a provider returns an error.

#### Scenario: Provider cannot process alerts
- **WHEN** a registered provider returns an error
- **THEN** Penda SHALL not invoke later providers and SHALL return a non-zero status
