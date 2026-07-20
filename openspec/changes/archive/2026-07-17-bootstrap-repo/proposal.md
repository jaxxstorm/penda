## Why

Penda needs an executable foundation before it can discover and apply Dependabot fixes. Users and CI jobs need a predictable way to select a repository directory and authenticate to GitHub without exposing credentials unnecessarily.

## What Changes

- Bootstrap the Go CLI entry point using `alecthomas/kong`.
- Add a `--dir` / `-d` option to select the directory Penda operates on.
- Accept a GitHub token from an explicit CLI option or the `GITHUB_TOKEN` environment variable, with the explicit option taking precedence.
- Provide non-interactive, readable terminal output suitable for local use and CI.

## Capabilities

### New Capabilities
- `command-line-configuration`: Configure the target directory and validate CLI input.
- `github-authentication`: Resolve GitHub credentials from a CLI option or `GITHUB_TOKEN`.

### Modified Capabilities

None.

## Impact

- Adds the initial Go command structure and dependencies on `alecthomas/kong` and Charmbracelet output libraries.
- Establishes the CLI interface consumed by local users and CI pipelines.
