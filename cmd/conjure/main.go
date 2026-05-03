package main

import (
	"fmt"
	"os"
)

// Version is overridden at release time via -ldflags "-X main.Version=...".
// Must remain a var, since constants cannot be linker-injected.
var Version = "1.0.0"

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "conjure: %v\n", err)
		os.Exit(1)
	}
}
