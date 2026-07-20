## ADDED Requirements

### Requirement: Origin remote discovery
Penda SHALL obtain the selected directory's `origin` remote URL using Git before requesting Dependabot alerts.

#### Scenario: Selected directory is a GitHub checkout
- **WHEN** Penda runs against a directory whose `origin` remote points to GitHub
- **THEN** Penda SHALL resolve the repository owner and name from that remote

#### Scenario: Origin remote is unavailable
- **WHEN** the selected directory is not a Git checkout or does not define an `origin` remote
- **THEN** Penda SHALL not call the GitHub API and SHALL exit with a non-zero status and an actionable error

### Requirement: GitHub remote URL normalization
Penda SHALL resolve owner and repository names from GitHub HTTPS and SSH remote URL forms, including an optional `.git` suffix. It SHALL reject remotes that do not identify a `github.com` repository.

#### Scenario: HTTPS remote
- **WHEN** `origin` is `https://github.com/octo-org/example.git`
- **THEN** Penda SHALL resolve owner `octo-org` and repository `example`

#### Scenario: SCP-style SSH remote
- **WHEN** `origin` is `git@github.com:octo-org/example.git`
- **THEN** Penda SHALL resolve owner `octo-org` and repository `example`

#### Scenario: Non-GitHub remote
- **WHEN** `origin` identifies a host other than `github.com`
- **THEN** Penda SHALL not call the GitHub API and SHALL exit with a non-zero status and an actionable error
