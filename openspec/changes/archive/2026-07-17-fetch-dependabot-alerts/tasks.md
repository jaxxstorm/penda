## 1. Repository Discovery

- [x] 1.1 Add a repository identity type and resolve the target directory's `origin` remote with `git -C <dir> remote get-url origin`.
- [x] 1.2 Parse supported GitHub HTTPS and SSH remote URL forms into owner and repository names, rejecting missing, malformed, and non-GitHub remotes before API access.
- [x] 1.3 Add table-driven tests for remote URL parsing and repository discovery failures.

## 2. Dependabot Alert Client

- [x] 2.1 Define the provider-facing alert model and map the required fields from GitHub Dependabot alert responses.
- [x] 2.2 Implement an authenticated GitHub REST client that requests repository Dependabot alerts, follows pagination, and returns safe status-aware errors.
- [x] 2.3 Add HTTP tests for authorization and accept headers, empty responses, multi-page retrieval, alert mapping, missing tokens, and non-success API responses without response-body leakage.

## 3. Provider Handoff And Command Integration

- [x] 3.1 Define the alert provider interface and invoke registered providers in order with the complete alert collection, stopping on a provider error.
- [x] 3.2 Integrate repository discovery, token validation, alert retrieval, empty initial provider registration, and concise alert-count status output into the CLI runtime.
- [x] 3.3 Add command and provider tests that confirm read-only default behavior, provider failure handling, non-zero errors, and credential-safe output.

## 4. Verification

- [x] 4.1 Run Go formatting and the full test suite.
