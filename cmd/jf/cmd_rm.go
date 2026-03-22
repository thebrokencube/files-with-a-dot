package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"os"
	"path/filepath"
)

func runRm(args []string) int {
	fs := flag.NewFlagSet("rm", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan for forest.yml")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	keys := fs.Args()
	if len(keys) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: jf rm <KEY>...\n")
		return 1
	}

	f, roots, code := loadForestOrFail(*dir, false)
	if code != 0 {
		return code
	}

	failed := false
	for _, key := range keys {
		node, err := forest.Resolve(roots, key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %s\n", key, err)
			failed = true
			continue
		}

		if len(node.Children) > 0 {
			fmt.Fprintf(os.Stderr, "✗ %s: has %d children — remove children first\n", key, len(node.Children))
			failed = true
			continue
		}

		path := filepath.Join(f.Dir, node.File)
		if err := os.Remove(path); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %s\n", key, err)
			failed = true
			continue
		}

		fmt.Printf("✓ removed %s (%s)\n", key, node.File)
	}

	if failed {
		return 1
	}
	return 0
}
