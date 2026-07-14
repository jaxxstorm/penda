# Penda

Penda applies targeted local dependency fixes for Dependabot alerts.

## Run locally

Run the complete Go package during development:

```sh
go run . --help
```

`go run main.go` compiles only `main.go`, so it does not include Penda's other
package files. Use `go run .` instead.
