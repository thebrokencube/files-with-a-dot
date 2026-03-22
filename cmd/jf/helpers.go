package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"jf/internal/output"
	"os"
	"strings"
)

// parseFlags wraps fs.Parse with trailing flag detection.
// Returns error on parse failure or flags found after positional arguments.
func parseFlags(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return err
	}
	for _, arg := range fs.Args() {
		if strings.HasPrefix(arg, "-") {
			err := fmt.Errorf("unknown flag %q after positional arguments (flags must come before arguments)", arg)
			fmt.Fprintln(os.Stderr, err)
			return err
		}
	}
	return nil
}

// loadForest finds and discovers a forest from the given directory.
// Returns the forest config, discovered roots, or a descriptive error.
func loadForest(dir string) (*forest.Forest, []*forest.Node, error) {
	f, err := forest.FindForest(dir)
	if err != nil {
		return nil, nil, err
	}
	if f == nil {
		return nil, nil, fmt.Errorf("No forest.yml found (searched up from %s)", dir)
	}

	roots, err := forest.Discover(f)
	if err != nil {
		return nil, nil, fmt.Errorf("Discovery failed: %s", err)
	}

	return f, roots, nil
}

// loadForestOrFail calls loadForest and reports the error.
// If jsonOut is true, errors go to stdout as JSON; otherwise stderr.
// Returns exit code 1 on failure, 0 on success.
func loadForestOrFail(dir string, jsonOut bool) (*forest.Forest, []*forest.Node, int) {
	f, roots, err := loadForest(dir)
	if err != nil {
		if jsonOut {
			output.Error(err.Error(), dir)
		} else {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		}
		return nil, nil, 1
	}
	return f, roots, 0
}
