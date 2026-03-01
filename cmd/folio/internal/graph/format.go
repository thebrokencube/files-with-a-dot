package graph

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// FormatDAG writes a human-readable text representation of the adjacency list.
// Each line: target-id -> [dep1, dep2]
// Targets with no dependencies are also included.
func FormatDAG(w io.Writer, adj map[string][]string, allTargets []string) {
	// Include all targets, even those without edges
	seen := make(map[string]bool)
	for _, t := range allTargets {
		seen[t] = true
	}
	for t := range adj {
		seen[t] = true
	}

	sorted := make([]string, 0, len(seen))
	for t := range seen {
		sorted = append(sorted, t)
	}
	sort.Strings(sorted)

	for _, tid := range sorted {
		deps := adj[tid]
		if len(deps) == 0 {
			fmt.Fprintf(w, "%s\n", tid)
		} else {
			sortedDeps := make([]string, len(deps))
			copy(sortedDeps, deps)
			sort.Strings(sortedDeps)
			fmt.Fprintf(w, "%s -> [%s]\n", tid, strings.Join(sortedDeps, ", "))
		}
	}
}
