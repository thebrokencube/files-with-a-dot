package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"os"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

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
			dendrik.WriteError(os.Stderr, err.Error(), dir)
		} else {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		}
		return nil, nil, dendrik.ExitUserError
	}
	return f, roots, dendrik.ExitOK
}
