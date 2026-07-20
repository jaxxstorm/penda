## Context

The bootstrap command validates a selected directory and resolves an optional GitHub token, but it does not yet inspect the checkout or contact GitHub. Penda needs Dependabot alert data as the input to future package-manager-specific updates while avoiding a dependency between GitHub transport, repository discovery, and update providers.

## Goals / Non-Goals

**Goals:**

- Resolve a public GitHub owner and repository name from the target checkout's `origin` remote.
- Retrieve every Dependabot alert using GitHub's repository alerts REST endpoint.
- Use the resolved token for authenticated API requests and never include it in output or errors.
- Hand a stable, typed alert model to registered providers without implementing a provider that modifies dependencies.

**Non-Goals:**

- Applying dependency updates, changing manifests, or running package-manager commands.
- Supporting GitHub Enterprise Server or a configurable GitHub API host.
- Discovering alternate remotes when `origin` is missing.
- Interactively authenticating or storing GitHub credentials.

## Decisions

- Invoke `git -C <target-dir> remote get-url origin` to discover the checkout remote. This supports standard repositories and worktrees without reimplementing Git configuration parsing. A missing remote, Git failure, or non-GitHub remote produces an actionable error before any API call.
- Normalize supported GitHub remote URL forms, including HTTPS and SSH URLs, into an owner and repository name, removing an optional `.git` suffix. Limit the first version to `github.com`; an API-base abstraction can be introduced when Enterprise support is required.
- Use the GitHub REST endpoint `GET /repos/{owner}/{repo}/dependabot/alerts` with `per_page=100`, following `Link` pagination until all pages are collected. Use the standard library HTTP client and narrowly scoped response structs rather than adding a GitHub SDK, because Penda currently requires one endpoint and a domain model independent of any SDK.
- Require a non-empty resolved token before requesting alerts. Send it only in the `Authorization: Bearer` header with GitHub's `Accept: application/vnd.github+json` header. Map non-success responses to status-aware errors without copying response bodies or request headers into terminal output.
- Define a domain `Alert` model containing the alert number, state, dependency package name, manifest path, scope, advisory identifiers and severity, vulnerable version range, and first patched version. The GitHub client maps API responses to this model so providers do not depend on GitHub response structures.
- Define a provider interface that receives a context and the complete alert set. The command orchestrator retrieves alerts and invokes each registered provider in order; the initial provider list is empty, so this change performs no patches. This creates a deliberate handoff point for future package-manager providers.
- Render a concise count of retrieved alerts after a successful handoff. Error output names the failing operation but never includes a token or raw GitHub response body.

## Risks / Trade-offs

- [Dependabot alerts require repository permissions and GitHub API access] -> Fail before the request when no token is configured and surface HTTP status failures without exposing credentials.
- [Git remote URLs have many valid forms] -> Support common GitHub HTTPS and SSH forms with table-driven tests; reject unsupported forms explicitly rather than guessing.
- [Large repositories have more alerts than one API page] -> Request the maximum supported page size and follow pagination links until exhausted.
- [A provider failure can leave future updates partially applied] -> Keep this change read-only and define provider invocation to stop and return an error; future mutating providers must define rollback behavior.
