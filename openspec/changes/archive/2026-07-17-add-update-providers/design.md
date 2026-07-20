## Context

Penda retrieves Dependabot alerts into a typed model and passes them to an empty provider list. Providers need the repository target directory as well as the alert collection to update the exact manifest identified by an alert. The first concrete providers cover the Python, npm, and GitHub Actions ecosystems without attempting broad dependency refreshes.

## Goals / Non-Goals

**Goals:**

- Register Python, npm, and GitHub Actions providers in the default CLI runtime.
- Carry alert ecosystem and target directory information through the provider boundary.
- Apply only alerts with a known first patched version to their declared manifest.
- Use ecosystem-native commands where available and make only targeted manifest edits.
- Prevent providers from operating outside the selected repository.

**Non-Goals:**

- Supporting Python ecosystems beyond pip, Pipenv, Poetry, and uv.
- Supporting Node package managers other than npm.
- Updating arbitrary YAML fields or GitHub Actions outside workflow `uses:` references.
- Resolving alerts without a first patched version, indirect dependency upgrades, or broad dependency refreshes.
- Rolling back a successfully completed native package-manager command.

## Decisions

- Extend the alert model with the Dependabot package ecosystem and change the provider contract to receive the selected target directory plus the complete alert collection. Built-in providers filter by ecosystem and manifest, while the command continues to stop on the first provider error.
- Register the three built-in providers in the default runtime in Python, npm, then GitHub Actions order. Existing test runtimes can inject a provider list, preserving isolated tests.
- Treat each alert's manifest path as a repository-relative path. Reject absolute paths and traversal outside the target directory before invoking a command or writing a file.
- Require a non-empty first patched version before a provider changes anything. Providers skip alerts for other ecosystems, unsupported manifests, or alerts without a patch version; this is safer than guessing a compatible update.
- The Python provider supports pip requirements files, Pipenv `Pipfile`, Poetry `pyproject.toml` and `poetry.lock`, and uv `uv.lock`. For pip requirements files it pins the matching direct requirement to the patched version and runs `python -m pip install -r <manifest>`. For Pipenv and Poetry it runs `pipenv install <package>==<version>`, `poetry add <package>@<version>`, or `poetry update <package>` from the manifest directory. For uv it runs `uv lock --upgrade-package <package>`. These tools update their native manifest and lockfile formats.
- GitHub may label alerts for uv and Poetry lockfiles as the `pip` ecosystem. The Python provider selects the native tool from the manifest file name in addition to the alert ecosystem. When an alert manifest is absent from the selected checkout, it skips the alert because the branch no longer contains the vulnerable dependency.
- The npm provider runs `npm install <package>@<version>` from the `package.json` directory, using `--save-dev` for development-scoped alerts. npm owns package.json and package-lock.json updates, avoiding manual lockfile edits.
- The GitHub Actions provider updates only `uses: owner/action@ref` values in an alert's `.github/workflows/*.yml` or `.yaml` manifest when the action name matches the alert package. It writes the patched version atomically and preserves unrelated workflow lines.
- Abstract command execution and filesystem writes behind narrow interfaces so provider tests do not require installed ecosystem tools or mutate fixtures.

## Risks / Trade-offs

- [Native tools may be missing or use incompatible project configuration] -> Return a concise provider error for matching alerts; unsupported ecosystems and manifests remain untouched.
- [Package manager commands can alter transitive dependencies] -> Invoke only targeted install commands with the first patched version; do not use broad update commands.
- [Text replacement in requirements and workflow files can match unintended content] -> Match direct dependency declarations or `uses:` values anchored to the alert package and manifest path, with focused fixture tests.
- [Malformed alert manifest paths could escape the repository] -> Validate paths before every filesystem or command operation.
- [Multiple alerts can target one manifest] -> Process alerts in deterministic order and use native tools or atomic writes so each change is visible before the next alert.
