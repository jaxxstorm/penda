## Context

The current root package contains the executable entry point plus GitHub transport, alert models, provider implementations, terminal rendering, and all tests. Penda needs boundaries that let GitHub callers and update providers evolve independently while keeping the root executable intentionally small.

## Goals / Non-Goals

**Goals:**

- Leave `main.go` as the only root Go source file.
- Isolate GitHub transport and repository discovery in `internal/github`.
- Isolate provider contracts and implementations in `internal/providers`.
- Keep command orchestration and rendering in `internal/app`.
- Preserve all existing behavior and test coverage.

**Non-Goals:**

- Changing CLI flags, API endpoints, provider selection, update commands, or output semantics.
- Making internal packages public APIs outside the Penda module.
- Adding dependency injection frameworks or changing external dependencies.

## Decisions

- `main.go` will call `app.Run` with process arguments, streams, working-directory lookup, and environment lookup. It will contain no CLI, transport, provider, or renderer logic.
- `internal/github` will own `Repository`, `Alert`, remote URL normalization, `DiscoverRepository`, and the GitHub Dependabot alert client. The package will expose narrow client interfaces and avoid imports from application or providers.
- `internal/providers` will import `internal/github` for alert data. It will own the provider interface, built-in provider registration, safe manifest resolution, native command execution, and update planning identity helpers needed by providers.
- `internal/app` will depend on GitHub and providers. It will parse Kong flags, coordinate repository discovery and alert retrieval, maintain progress and final summary state, and render Bubble Tea/Bubbles cards.
- Tests will move with their production code. Application tests will use exported internal-package types and fake clients/providers; GitHub and provider tests will remain package-local to retain direct coverage of parsing and safety helpers.

## Risks / Trade-offs

- [The refactor can create import cycles] -> GitHub remains independent, providers depend only on GitHub, and app is the only package that imports both.
- [Moving unexported tests can reduce coverage] -> Preserve existing test cases and add package-level fakes at the app boundary.
- [Exporting internal types broadens package APIs] -> Export only the types and methods required across internal package boundaries.
