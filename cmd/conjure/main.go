package main

import (
	"fmt"
	"os"
)

const Version = "0.2.0-dev"

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("conjure %s\n", Version)
		return
	}
	fmt.Fprintln(os.Stderr, "conjure: not yet implemented")
	os.Exit(1)
}
