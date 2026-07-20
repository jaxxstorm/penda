## Context

Penda currently has only its Go module definition. This change establishes a non-interactive command that can be used locally or by CI before Dependabot issue discovery and package-manager updates are added. The initial CLI needs one consistent configuration path for its working directory and GitHub credential.

## Goals / Non-Goals

**Goals:**

- Provide a Go executable with flags parsed by `alecthomas/kong`.
- Operate on a caller-selected directory, defaulting to the current working directory.
- Resolve a GitHub token deterministically without printing it.
- Produce concise terminal output through Charmbracelet presentation libraries without requiring terminal interaction.

**Non-Goals:**

- Fetching or processing Dependabot issues.
- Detecting package managers or applying dependency updates.
- Persisting credentials or configuring GitHub authentication.
- Building an interactive TUI.

## Decisions

- Define a root command with `--dir` and `-d` bound to the target directory. The command will use the current working directory when the option is omitted, then validate that the resolved path exists and is a directory. This gives CI and local callers an explicit target while preserving the conventional default.
- Define a `--token` option for a GitHub access token and configure Kong to read `GITHUB_TOKEN` as its environment fallback. A non-empty flag value overrides the environment value. Keeping resolution in Kong avoids duplicated precedence logic and makes the interface visible in generated help.
- Keep the effective token in command configuration and pass it only to future GitHub clients. Do not render it in normal output, errors, or diagnostics to avoid credential leaks in CI logs.
- Separate command parsing and configuration validation from operational work. Future issue discovery and update orchestration can consume a validated configuration without coupling to Kong or terminal rendering.
- Use Lipgloss for styled status and error presentation while retaining ordinary stdout/stderr streams and exit codes. This preserves scriptability; a Bubble Tea application was considered but rejected because Penda is not a TUI.

## Risks / Trade-offs

- [A missing token can prevent access to private repositories or encounter API rate limits] -> Allow configuration to resolve without a token for public-repository use, and let the future GitHub operation report authentication requirements clearly.
- [ANSI styling can reduce log readability] -> Keep messages concise and ensure their text remains meaningful without color support.
- [Directory validation can reject paths that are valid only after later setup] -> Validate only existence and directory type; do not impose repository or package-manager requirements in this bootstrap change.
