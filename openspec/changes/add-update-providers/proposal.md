## Why

Penda can now retrieve Dependabot alerts but does not perform any updates. Adding providers for the common Python, npm, and GitHub Actions ecosystems turns alert data into narrowly scoped local dependency fixes.

## What Changes

- Register concrete update providers with the alert-processing runtime.
- Add a Python provider for Dependabot's pip, Pipenv, Poetry, and uv projects.
- Add an npm provider for Node.js dependency alerts.
- Add a GitHub Actions provider for workflow action-reference alerts.
- Ensure providers only act on alerts and manifests they support, using native tooling and preserving unrelated dependencies.

## Capabilities

### New Capabilities
- `provider-registration`: Register and invoke the built-in update providers for retrieved alerts.
- `python-dependency-updates`: Apply supported Python dependency fixes with pip, Pipenv, Poetry, and uv.
- `npm-dependency-updates`: Apply supported npm dependency fixes with npm.
- `github-actions-updates`: Update supported GitHub Actions references in workflow files.

### Modified Capabilities

None.

## Impact

- Extends the alert provider pipeline from read-only handoff to local repository updates.
- Requires discovery of Python, Node.js, and GitHub Actions manifest files and execution of native package-management commands.
- Adds safety checks, provider selection, command execution, and filesystem tests.
