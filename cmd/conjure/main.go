package main

import (
	"fmt"
	"os"
)

// Version is overridden at release time via -ldflags "-X main.Version=...".
// Must remain a var, since constants cannot be linker-injected.
var Version = "0.3.0-dev"

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "conjure: %v\n", err)
		os.Exit(1)
	}
}
