## 1. Package Boundaries

- [x] 1.1 Create `internal/github` and move repository discovery, alert models, and GitHub alert retrieval with package-local tests.
- [x] 1.2 Create `internal/providers` and move provider interfaces, implementations, and package-local tests.
- [x] 1.3 Create `internal/app` for configuration, orchestration, planning, rendering, and application tests.
- [x] 1.4 Reduce root `main.go` to the executable entry point and remove other root Go source and test files.

## 2. Verification

- [x] 2.1 Update imports, fakes, and renderer integration to preserve all existing behaviors.
- [x] 2.2 Run Go formatting, tests, vet, build, CLI help, and OpenSpec validation.
