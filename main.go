package main

import (
	"os"

	"github.com/jaxxstorm/penda/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr, os.Getwd, os.Getenv))
}
