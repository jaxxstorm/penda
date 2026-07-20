## Purpose

Configure GitHub authentication without exposing credentials.

## Requirements

### Requirement: GitHub token configuration
Penda SHALL accept a GitHub access token through a `--token` command-line option or the `GITHUB_TOKEN` environment variable.

#### Scenario: Token is supplied on the command line
- **WHEN** Penda is invoked with a non-empty `--token` value
- **THEN** Penda SHALL use that value as the GitHub access token

#### Scenario: Token is supplied through the environment
- **WHEN** Penda is invoked without a non-empty `--token` value and `GITHUB_TOKEN` is set
- **THEN** Penda SHALL use the `GITHUB_TOKEN` value as the GitHub access token

### Requirement: Token precedence
Penda SHALL prefer a non-empty `--token` value over `GITHUB_TOKEN` when both are supplied.

#### Scenario: Both token sources are set
- **WHEN** Penda is invoked with a non-empty `--token` value and `GITHUB_TOKEN` is set
- **THEN** Penda SHALL use the `--token` value as the GitHub access token

### Requirement: Credential-safe output
Penda SHALL NOT include the configured GitHub access token in normal output, error messages, or diagnostics.

#### Scenario: Command reports its configuration
- **WHEN** Penda emits status or error output after resolving a GitHub access token
- **THEN** the output SHALL not contain the token value
