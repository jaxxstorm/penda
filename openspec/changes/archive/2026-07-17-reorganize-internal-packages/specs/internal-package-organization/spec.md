## ADDED Requirements

### Requirement: Minimal root executable
Penda SHALL keep `main.go` as the only root Go source file. It SHALL delegate CLI execution to an internal application package.

#### Scenario: CLI is built from the module root
- **WHEN** Penda is built or run from the module root
- **THEN** the root executable SHALL invoke the internal application package and preserve the existing CLI interface

### Requirement: GitHub package boundary
Penda SHALL place Git remote discovery, repository identity, Dependabot alert data, and GitHub REST retrieval in `internal/github`.

#### Scenario: Application retrieves alerts
- **WHEN** the application resolves a repository and retrieves Dependabot alerts
- **THEN** it SHALL use types and clients from `internal/github`

### Requirement: Provider package boundary
Penda SHALL place provider contracts, built-in provider registration, and Python/npm/GitHub Actions update behavior in `internal/providers`.

#### Scenario: Application applies planned updates
- **WHEN** the application processes retrieved alerts
- **THEN** it SHALL invoke providers from `internal/providers`

### Requirement: Behavioral preservation
The package reorganization SHALL preserve command flags, open-alert filtering, provider execution, output rendering, native command diagnostics, and credential redaction.

#### Scenario: Existing CLI workflow runs after reorganization
- **WHEN** Penda is invoked with a target directory and GitHub token
- **THEN** it SHALL retain the prior observable behavior and process exit statuses

### Requirement: Package-local tests
Penda SHALL keep GitHub, provider, and application tests alongside their owning internal packages.

#### Scenario: Test suite runs
- **WHEN** `go test ./...` is executed
- **THEN** it SHALL run tests for the internal GitHub, provider, and application packages
