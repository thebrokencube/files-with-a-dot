package main

import (
	"os"
)

var version = "dev" // overridden at build time via -ldflags -X main.version (see Makefile/VERSION)

func main() {
	os.Exit(buildRoot().Execute(os.Args[1:]))
}
