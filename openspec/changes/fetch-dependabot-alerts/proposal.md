## Why

Penda cannot plan dependency updates until it knows which Dependabot alerts apply to the repository selected by the user. It needs a reliable bridge from a local Git checkout to GitHub alert data while keeping future package-manager update logic decoupled from GitHub retrieval.

## What Changes

- Discover the GitHub repository owner and name from the selected directory's Git remote URL.
- Retrieve every Dependabot alert for that repository through the GitHub REST API using Penda's configured token.
- Introduce a provider interface and hand retrieved alerts to registered providers without implementing dependency patching yet.
- Report concise retrieval results and actionable failures without exposing credentials.

## Capabilities

### New Capabilities
- `github-repository-discovery`: Resolve a GitHub repository identity from a selected local checkout.
- `dependabot-alert-retrieval`: Retrieve all Dependabot alerts for the resolved repository from GitHub.
- `alert-provider-handoff`: Pass retrieved alert data through an extensible provider interface without applying changes.

### Modified Capabilities

None.

## Impact

- Extends the Go command runtime after directory validation.
- Adds Git remote inspection and authenticated GitHub REST API access.
- Establishes the data boundary that future package-manager providers will consume.
