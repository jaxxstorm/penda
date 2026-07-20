## ADDED Requirements

### Requirement: Built-in provider registration
Penda SHALL register Python, npm, and GitHub Actions providers in its default runtime after Dependabot alerts are retrieved.

#### Scenario: Alerts are retrieved by the CLI
- **WHEN** Penda successfully retrieves Dependabot alerts
- **THEN** it SHALL pass the complete alert collection and selected target directory to the Python, npm, and GitHub Actions providers

### Requirement: Provider execution contract
Penda SHALL provide each provider with the target directory and complete mapped alerts. It SHALL stop later provider execution and exit with a non-zero status when a provider returns an error.

#### Scenario: Provider cannot apply a matching alert
- **WHEN** a registered provider returns an error
- **THEN** Penda SHALL not invoke later providers and SHALL exit with a non-zero status

### Requirement: Safe provider targets
Penda SHALL treat alert manifest paths as repository-relative paths and SHALL reject paths that are absolute or resolve outside the selected target directory.

#### Scenario: Alert specifies a traversal path
- **WHEN** an alert manifest path would resolve outside the selected target directory
- **THEN** Penda SHALL not execute a package-management command or write a file for that alert

### Requirement: Targeted provider selection
Each provider SHALL ignore alerts for unsupported ecosystems, unsupported manifest types, or alerts without a first patched version.

#### Scenario: Alert lacks a patched version
- **WHEN** an alert does not contain a first patched version
- **THEN** no provider SHALL modify a manifest or invoke a package-management command for that alert
