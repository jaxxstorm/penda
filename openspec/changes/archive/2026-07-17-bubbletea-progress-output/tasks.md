## 1. Renderer Foundation

- [x] 1.1 Add Bubble Tea Bubbles progress dependencies and create a non-interactive output renderer.
- [x] 1.2 Move lifecycle, progress, completion, and failure rendering behind the renderer while preserving runtime behavior and append-only output.

## 2. Progress Presentation

- [x] 2.1 Render planned-update progress with a Bubbles progress bar and plain current, total, remaining, and alert identity text.
- [x] 2.2 Render compact lifecycle and completion cards for discovery, retrieval, and update planning.
- [x] 2.3 Render safe failure cards that retain bounded native command diagnostics and credential redaction.

## 3. Verification

- [x] 3.1 Add renderer tests for lifecycle cards, progress identity and counts, non-interactive output, and safe failures.
- [x] 3.2 Run Go formatting, tests, vet, and OpenSpec validation.

## 4. Live Ticket Refinement

- [x] 4.1 Redraw a single terminal ticket in place without input handling and render a final quantitative run summary.
- [x] 4.2 Add live-ticket and final-summary regression tests.

## 5. Detailed Summary

- [x] 5.1 Record each applied update and render its ecosystem, package/version, native method, and manifest in the final summary.
- [x] 5.2 Add detailed-summary and native-method regression tests.
