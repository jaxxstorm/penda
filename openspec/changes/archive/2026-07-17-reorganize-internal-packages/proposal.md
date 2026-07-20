## Why

Penda's root package currently combines command orchestration, GitHub API access, provider implementations, rendering, and tests. This makes the CLI difficult to extend and prevents future callers from using GitHub access or providers without importing the executable package.

## What Changes

- Keep `main.go` as the sole root Go source file and make it only start the application.
- Move Git remote discovery, repository identity, alert types, and GitHub REST access to `internal/github`.
- Move provider contracts, safe filesystem and command operations, and Python/npm/GitHub Actions providers to `internal/providers`.
- Move CLI configuration, orchestration, update planning, and rendering to `internal/app`.
- Move unit tests beside their owning internal packages without changing observable CLI behavior.

## Capabilities

### New Capabilities
- `internal-package-organization`: Organize Penda's executable, GitHub access, providers, and application orchestration behind internal package boundaries.

### Modified Capabilities

None.

## Impact

- Replaces root-level implementation files with internal packages and package-local tests.
- Preserves CLI flags, GitHub alert retrieval, provider execution, progress rendering, and safe error output.
