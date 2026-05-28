package main

import (
	"github.com/ralvarez/m2p/internal/cli"
)

// Injected by goreleaser via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.Execute(version, commit, date)
}
