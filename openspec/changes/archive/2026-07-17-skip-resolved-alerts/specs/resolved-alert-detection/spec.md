## ADDED Requirements

### Requirement: Manifest-scoped resolved alert detection
Before applying an alert, Penda SHALL inspect the selected repository's uncommitted diff for the alert manifest. It SHALL skip the update when an added line contains both the alert package and first patched version.

#### Scenario: Prior npm update is uncommitted
- **WHEN** `package-lock.json` has an added diff line containing `undici` and `6.27.0`
- **THEN** Penda SHALL not run npm again for the matching alert

### Requirement: Unresolved alerts remain actionable
Penda SHALL invoke the native provider when the manifest diff does not contain both the alert package and patched version.

#### Scenario: Manifest changed for another dependency
- **WHEN** an alert manifest diff does not contain the alert package and patched version together
- **THEN** Penda SHALL apply the alert with its native provider

### Requirement: Skipped update summary
Penda SHALL include skipped resolved updates in the final summary count.

#### Scenario: Run skips an already fixed alert
- **WHEN** Penda skips an alert due to a matching manifest diff
- **THEN** the final summary SHALL include it in the skipped or not attempted count
