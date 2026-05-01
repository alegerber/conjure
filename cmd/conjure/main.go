package main

import (
	"fmt"
	"os"
)

const Version = "0.2.0-dev"

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "conjure: %v\n", err)
		os.Exit(1)
	}
}
