## 1. Provider Runtime Foundation

- [x] 1.1 Extend the alert model with the Dependabot package ecosystem and map it from GitHub responses.
- [x] 1.2 Extend the provider contract and runtime handoff to pass the selected target directory with the complete alert collection.
- [x] 1.3 Add safe repository-relative manifest resolution and injectable command and filesystem operations for providers.
- [x] 1.4 Register Python, npm, and GitHub Actions providers in the default runtime while retaining injectable providers for tests.

## 2. Python Updates

- [x] 2.1 Implement Python provider filtering for pip, Pipenv, and Poetry alerts with supported manifests and first patched versions.
- [x] 2.2 Implement targeted pip requirements pinning and `python -m pip install -r` execution.
- [x] 2.3 Implement targeted Pipenv and Poetry package commands from the alert manifest directory.
- [x] 2.4 Support Poetry and uv lockfile alerts reported as the pip ecosystem, using targeted native lockfile commands.
- [x] 2.5 Skip Python alerts whose named manifest is absent from the selected checkout.

## 3. npm And GitHub Actions Updates

- [x] 3.1 Implement the npm provider with targeted production and development dependency installation.
- [x] 3.2 Implement the GitHub Actions provider with atomic, targeted workflow `uses:` reference updates.

## 4. Verification

- [x] 4.1 Add fixture-based tests for provider registration, ecosystem filtering, command arguments, manifest traversal rejection, targeted file edits, unchanged unsupported alerts, and provider errors.
- [x] 4.2 Run Go formatting and the full test suite.
- [x] 4.3 Add regression tests for pip-classified Poetry and uv lockfiles and missing Python manifests.
