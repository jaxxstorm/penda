## 1. Command Foundation

- [x] 1.1 Add the Go command entry point and required Kong and Charmbracelet dependencies.
- [x] 1.2 Define the root CLI configuration with `--dir` / `-d` and `--token` options, including `GITHUB_TOKEN` environment fallback.
- [x] 1.3 Resolve the default target directory and validate that the selected path exists and is a directory before executing the command.

## 2. Runtime Behavior

- [x] 2.1 Apply explicit non-empty `--token` precedence over `GITHUB_TOKEN` and keep the resolved value out of all terminal output.
- [x] 2.2 Add concise, non-interactive styled status and error output with correct process exit statuses.

## 3. Verification

- [x] 3.1 Add tests for explicit and default target-directory selection and invalid-directory failure.
- [x] 3.2 Add tests for token environment fallback, flag precedence, and token-safe output.
- [x] 3.3 Run Go formatting and the full test suite.
