## ADDED Requirements

### Requirement: Supported Python ecosystems
The Python provider SHALL process alerts whose package ecosystem is pip, Pipenv, Poetry, or uv and whose manifest is a supported Python manifest within the target directory. It SHALL select Poetry and uv tooling for `poetry.lock` and `uv.lock` manifests even when GitHub reports the alert ecosystem as pip.

#### Scenario: Python alert targets a supported manifest
- **WHEN** an alert for pip, Pipenv, Poetry, or uv names a supported Python manifest and has a first patched version
- **THEN** the Python provider SHALL attempt a targeted update for that package and version

### Requirement: pip requirements updates
For a pip alert targeting a requirements file, the Python provider SHALL update only the matching direct requirement to the first patched version and SHALL run `python -m pip install -r <manifest>`.

#### Scenario: pip alert has a patched version
- **WHEN** a pip alert targets `requirements.txt` for package `example` with first patched version `1.2.3`
- **THEN** the provider SHALL pin the matching `example` requirement to `1.2.3` and install that requirements file

### Requirement: Pipenv and Poetry updates
For Pipenv and Poetry alerts, the Python provider SHALL invoke the respective native package manager from the target manifest directory with the alert package and first patched version.

#### Scenario: Poetry alert has a patched version
- **WHEN** a Poetry alert targets `pyproject.toml` for package `example` with first patched version `1.2.3`
- **THEN** the provider SHALL invoke Poetry to add `example` at version `1.2.3` from that manifest directory

### Requirement: Python lockfile updates
For a Poetry lockfile alert, the Python provider SHALL invoke `poetry update` for the alert package. For a uv lockfile alert, it SHALL invoke `uv lock --upgrade-package` for the alert package.

#### Scenario: uv lockfile alert has a patch
- **WHEN** an alert targets `uv.lock` for package `example` with first patched version `1.2.3`
- **THEN** the provider SHALL invoke uv to upgrade only `example` in that lockfile's project

### Requirement: Missing Python manifests
The Python provider SHALL skip an alert when its named Python manifest does not exist in the selected target directory.

#### Scenario: Alert manifest is absent locally
- **WHEN** a Python alert names a manifest that is absent from the selected checkout
- **THEN** the provider SHALL not invoke a package-management command or fail the overall Penda run for that alert

### Requirement: Python update safety
The Python provider SHALL NOT run a broad upgrade command or modify a Python manifest for an alert without a first patched version.

#### Scenario: Python alert has no patched version
- **WHEN** a Python alert does not include a first patched version
- **THEN** the provider SHALL leave Python manifests and lockfiles unchanged
