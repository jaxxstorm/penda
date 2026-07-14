## Context

Penda currently renders lifecycle stages, progress, completion, and failures through separate Lipgloss helpers. The result is readable but visually fragmented, and planned update progress competes with diagnostic text. The CLI must remain non-interactive and usable in CI.

## Goals / Non-Goals

**Goals:**

- Use Bubble Tea's Bubbles progress component to render consistent planned-update progress.
- Render one compact status card for lifecycle stages, current update details, completion, and failures.
- Preserve append-only output, standard exit statuses, and readable non-terminal logs.
- Retain bounded, redacted native command diagnostics after failures.

**Non-Goals:**

- Starting a Bubble Tea `Program`, entering an alternate screen, or accepting keyboard input.
- Adding spinners, timers, or continuously redrawing terminal output.
- Changing alert selection, update planning, provider execution, or error semantics.

## Decisions

- Add `github.com/charmbracelet/bubbles/progress` and instantiate its progress model directly. Penda will call `ViewAs` for each known progress update rather than run a Bubble Tea event loop. This uses Bubble Tea presentation components without becoming an interactive TUI.
- Consolidate output into a small renderer with stage, progress, success, and error methods. The runtime will retain its existing lifecycle boundaries and delegate rendering to this component.
- Render progress as a compact card containing Penda's stage label, a Bubbles bar, `current/total` planned updates, remaining count, and the ecosystem, package, target version, and manifest. Each record remains append-only for CI log durability.
- Render errors beneath a consistent failure card and preserve safe command diagnostics. Never render configured tokens or unbounded command output.
- Use a neutral ASCII fallback bar when color is disabled or the output is not a terminal. Bubble Tea markup is cosmetic only; all essential progress text remains present in the fallback.
- When stdout is a terminal, clear and redraw the prior card in place with ANSI cursor controls. This is a display-only ticket: it does not enter an alternate screen, create a Bubble Tea program, or read keyboard or mouse input. Render a final summary card with planned, completed, failed, and unattempted update counts.

## Risks / Trade-offs

- [Bubbles adds a dependency for a non-interactive renderer] -> Use only the progress component and avoid the Bubble Tea program runtime.
- [Repeated cards can lengthen CI logs] -> Keep each progress event to a compact single-line card and avoid redraw loops.
- [Terminal color support varies] -> Include current/total, remaining count, and update identity as plain text independent of styling.
- [Renderer changes can obscure failures] -> Test rendered success and error output for the existing command and redaction details.
