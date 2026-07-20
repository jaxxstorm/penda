## 1. Go Module Provider

- [x] 1.1 Add a Go module provider to the built-in provider registry with the existing injectable provider operations.
- [x] 1.2 Filter for complete `gomod` alerts targeting `go.mod`, resolve the manifest safely, and skip manifests absent from the selected checkout.
- [x] 1.3 Invoke `go get <module>@<first-patched-version>` from the affected module directory, reusing resolved-alert and duplicate-update protections.

## 2. Verification And Documentation

- [x] 2.1 Add provider tests for registered Go module support, exact command arguments and working directory, nested modules, and duplicate or previously resolved alerts.
- [x] 2.2 Add provider tests confirming incomplete, unsupported, missing, and traversal-path Go module alerts do not run the Go toolchain or modify files outside the target directory.
- [x] 2.3 Update README ecosystem and native-tool requirement documentation for Go modules.
- [x] 2.4 Run Go formatting and the full test suite.
