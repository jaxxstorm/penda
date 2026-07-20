## ADDED Requirements

### Requirement: Workflow action reference updates
The GitHub Actions provider SHALL process GitHub Actions ecosystem alerts that target `.github/workflows` YAML files and include a first patched version.

#### Scenario: Workflow action has a patch
- **WHEN** an alert for action `actions/checkout` targets `.github/workflows/ci.yml` with first patched version `v4.1.0`
- **THEN** the provider SHALL update matching `uses: actions/checkout@...` references in that workflow to `actions/checkout@v4.1.0`

### Requirement: Targeted workflow edits
The GitHub Actions provider SHALL change only matching action references in the alert's named workflow manifest and SHALL preserve unrelated workflow content.

#### Scenario: Workflow contains another action
- **WHEN** an alert targets `actions/checkout` and the same workflow also uses `actions/setup-node`
- **THEN** the provider SHALL not modify the `actions/setup-node` reference

### Requirement: GitHub Actions update safety
The GitHub Actions provider SHALL NOT modify files outside `.github/workflows` or update an alert without a first patched version.

#### Scenario: Alert targets a non-workflow manifest
- **WHEN** a GitHub Actions alert names a manifest outside `.github/workflows`
- **THEN** the provider SHALL leave repository files unchanged
