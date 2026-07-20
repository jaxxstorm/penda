# Penda

Penda applies available fixes for open Dependabot alerts directly in a local
GitHub repository. It supports Go modules, npm, pip, Pipenv, Poetry, uv, and
GitHub Actions dependencies.

## Install

### Homebrew

macOS and Linux:

```sh
brew install --cask jaxxstorm/tap/penda
```

### Scoop

Windows:

```powershell
scoop bucket add jaxxstorm https://github.com/jaxxstorm/scoop-bucket
scoop install penda
```

### GitHub Releases

Download the archive for your operating system and architecture from the
[releases page](https://github.com/jaxxstorm/penda/releases). Release archives
are named `penda_<version>_<os>_<arch>.tar.gz` for macOS and Linux, and
`penda_<version>_windows_<arch>.zip` for Windows.

For example, on Linux ARM64:

```sh
tar -xzf penda_<version>_linux_arm64.tar.gz
install -m 0755 penda ~/.local/bin/penda
```

Verify a downloaded archive against the `checksums.txt` file attached to the
same release before installing it.

### From source

With the Go version declared in [`go.mod`](go.mod) installed:

```sh
go install github.com/jaxxstorm/penda@latest
```

## Usage

Penda must run against a local Git repository whose `origin` remote points to
GitHub. Set `GITHUB_TOKEN` to a token that can read that repository's
Dependabot alerts, then run Penda from the repository root:

```sh
export GITHUB_TOKEN=github_pat_...
cd path/to/repository
penda
```

Use `--dir` to target a different repository directory, or `--token` to pass a
token for one invocation:

```sh
penda --dir path/to/repository
penda --token github_pat_...
```

Penda reads open Dependabot alerts and updates the affected manifest or lock
file to the first patched version. Review the resulting changes and run your
project's tests before committing them.

## Requirements

- Git, with an `origin` remote that identifies a `github.com` repository.
- A GitHub token with permission to read Dependabot alerts for that repository.
- The native package tool required by the alerts being fixed: `go`, `npm`,
  `python` and `pip`, `pipenv`, `poetry`, or `uv`. GitHub Actions dependency
  updates do not require an additional tool.

## Development

Run the complete Go package during development:

```sh
go run . --help
```

Use `go run .` rather than `go run main.go`, which compiles only `main.go` and
does not include Penda's other package files.
