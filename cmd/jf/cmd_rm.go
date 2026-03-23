package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runRm(args []string) int {
	fs := dendrik.NewFlagSet("rm")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	keys := fs.GetArgs()
	if len(keys) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: jf rm <KEY>...\n")
		return dendrik.ExitUserError
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
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}
