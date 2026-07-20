## Context

Penda's provider runtime already passes all Dependabot alerts and the selected repository directory to built-in providers. Python, npm, and GitHub Actions providers filter their matching alerts, validate manifest paths within the target checkout, skip incomplete or already-resolved alerts, and use their ecosystem's native tooling for targeted updates. Go module alerts are reported by Dependabot with the `gomod` ecosystem and name the affected `go.mod` file, but no registered provider currently handles them.

## Goals / Non-Goals

**Goals:**

- Apply a first patched Go module version for a `gomod` Dependabot alert targeting an in-repository `go.mod`.
- Use the Go toolchain from the affected module directory so `go.mod` and `go.sum` remain tool-managed.
- Match existing provider safety behavior for invalid paths, missing manifests, incomplete alerts, duplicate work, and previously resolved changes.
- Keep provider execution testable without a Go toolchain installed.

**Non-Goals:**

- Supporting non-module Go dependency managers or manifests other than `go.mod`.
- Running `go get -u`, `go mod tidy`, or another broad dependency update command.
- Resolving alerts without a first patched version, indirect version selection beyond the Go toolchain's normal targeted command, or reverting successful Go tool changes.

## Decisions

- Add a `goModuleProvider` to the built-in provider list after the existing providers. It will receive the same provider operations and alert collection as the other ecosystems, avoiding a separate update pipeline.
- Match only alerts whose ecosystem is `gomod`, whose manifest basename is `go.mod`, and which include a package name and first patched version. This mirrors Dependabot's ecosystem identifier and prevents Go commands for unrelated files or incomplete alert data.
- Resolve the manifest through the existing repository-relative path guard and require it to exist. A missing `go.mod` is skipped because an alert can outlive the dependency or branch state in the selected checkout; an unsafe path remains an error before any command is run.
- Invoke `go get <module>@<first-patched-version>` from the directory containing the alerted `go.mod`. `go get` is Go's targeted module-version operation and lets the toolchain update both `go.mod` and `go.sum`; it avoids manual module-file editing and broad upgrade flags.
- Reuse the existing resolved-manifest check and per-command de-duplication pattern so an already-applied alert or duplicate alert does not run `go get` again. Provider tests will inject command and filesystem operations and assert the working directory and exact command arguments.

## Risks / Trade-offs

- [The Go toolchain can update dependent module requirements while resolving the requested module] -> Invoke only `go get <module>@<version>` and do not run broad upgrade or tidy commands.
- [The selected checkout no longer contains the alert's module] -> Skip missing `go.mod` files without failing the complete Penda run.
- [A Go binary may be unavailable or reject the selected version] -> Return the native command error for matching alerts so CI and users receive the failing command context.
- [A malformed manifest path could escape the checkout] -> Reuse the existing manifest resolver before inspection or command execution.
