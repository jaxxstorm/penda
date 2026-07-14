## Why

Penda's current line-by-line Lipgloss output exposes useful data but does not make the update lifecycle, current target, or completion state easy to scan. Users need a clearer progress page while retaining non-interactive, CI-safe command behavior.

## What Changes

- Replace ad-hoc status and progress lines with a Bubble Tea/Bubbles-backed progress renderer.
- Present discovery, retrieval, update progress, completion, and failure as a consistent visual page.
- Redraw a single live ticket in interactive terminals while retaining append-only CI output.
- Preserve plain, append-only output when stdout is not an interactive terminal; do not create an interactive TUI or require input.
- Keep command diagnostics and credential redaction intact in the failure view.

## Capabilities

### New Capabilities
- `bubbletea-progress-rendering`: Render Penda's operation lifecycle and planned update progress with Bubble Tea components in interactive terminals and readable fallback output in CI.

### Modified Capabilities

None.

## Impact

- Replaces output helpers in the CLI runtime.
- Adds Charmbracelet Bubble Tea and Bubbles dependencies.
- Extends renderer tests for terminal and non-terminal output.
