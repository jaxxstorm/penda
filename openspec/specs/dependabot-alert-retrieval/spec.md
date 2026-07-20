## Purpose

Retrieve and map repository Dependabot alerts through GitHub's API.

## Requirements

### Requirement: Authenticated Dependabot alert retrieval
Penda SHALL require a configured GitHub token before requesting alerts and SHALL use that token to authenticate requests to GitHub's repository Dependabot alerts endpoint.

#### Scenario: Token is configured
- **WHEN** Penda resolves a GitHub repository and a non-empty token
- **THEN** Penda SHALL request that repository's Dependabot alerts using the token in an authorization header

#### Scenario: Token is not configured
- **WHEN** Penda resolves a GitHub repository without a token from `--token` or `GITHUB_TOKEN`
- **THEN** Penda SHALL not call the GitHub API and SHALL exit with a non-zero status and an actionable authentication error

### Requirement: Complete alert collection
Penda SHALL retrieve every page of the GitHub Dependabot alerts response and combine the alerts into one collection before provider handoff.

#### Scenario: Alerts span multiple pages
- **WHEN** GitHub returns a pagination link for an additional alerts page
- **THEN** Penda SHALL request the linked pages until no additional page remains

#### Scenario: Repository has no alerts
- **WHEN** GitHub returns an empty Dependabot alerts response
- **THEN** Penda SHALL continue successfully with an empty alert collection

### Requirement: Alert data mapping
Penda SHALL map each retrieved alert to a provider-facing model that includes its number, state, dependency package name, manifest path, dependency scope, advisory identifiers and severity, vulnerable version range, and first patched version when GitHub provides one.

#### Scenario: Alert includes a patched version
- **WHEN** GitHub returns an alert with a first patched version
- **THEN** Penda SHALL include that version in the provider-facing alert model

### Requirement: Safe API failure reporting
Penda SHALL report GitHub API failures with a non-zero status and SHALL NOT include the configured token or raw API response body in output.

#### Scenario: GitHub rejects the request
- **WHEN** GitHub returns a non-success response for an alerts request
- **THEN** Penda SHALL exit with a non-zero status and an error that identifies the failed GitHub operation without exposing credentials or the response body
