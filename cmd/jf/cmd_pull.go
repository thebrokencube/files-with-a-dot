package main

import (
	"flag"
	"fmt"
	"jf/internal/pipeline"
	"os"
)

func runPull(args []string) int {
	fs := flag.NewFlagSet("pull", flag.ContinueOnError)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	positional := fs.Args()
	if len(positional) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: jf pull <KEY> <FILE>\n")
		return 1
	}

	key := positional[0]
	filePath := positional[1]

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	out, err := p.View(key, "description", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: acli error\n  %s\n", key, err)
		return 2
	}

	if err := os.WriteFile(filePath, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: write error\n  %s\n", filePath, err)
		return 1
	}

	fmt.Printf("✓ Pulled %s description -> %s\n", key, filePath)
	return 0
}
