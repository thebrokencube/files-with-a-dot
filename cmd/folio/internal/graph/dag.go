package graph

import (
	"fmt"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

// BuildOutputMap creates a map from output key to list of producing target IDs.
// Keys are "path:<relative-path>" or "ext:<system>:<id>:<field>".
// Multiple producers for the same key indicates a collision.
func BuildOutputMap(f *config.Folio) map[string][]string {
	m := make(map[string][]string)
	for tid, target := range f.Targets {
		for _, out := range target.Outputs {
			if out.Path != "" {
				key := fmt.Sprintf("path:%s", out.Path)
				m[key] = append(m[key], tid)
			}
			if out.External != "" && out.ID != "" {
				key := fmt.Sprintf("ext:%s:%s:%s", out.External, out.ID, out.Field)
				m[key] = append(m[key], tid)
			}
		}

		// Batch item outputs
		if target.Batch != nil {
			for _, item := range target.Batch.Items {
				out := target.Batch.ResolveItemOutput(item)
				if out.External != "" && out.ID != "" {
					key := fmt.Sprintf("ext:%s:%s:%s", out.External, out.ID, out.Field)
					m[key] = append(m[key], tid)
				}
			}
		}
	}
	return m
}

// SingleProducerMap returns a map from output key to single producer target ID,
// filtering out any keys with multiple producers (collisions).
func SingleProducerMap(outputMap map[string][]string) map[string]string {
	m := make(map[string]string)
	for k, v := range outputMap {
		if len(v) == 1 {
			m[k] = v[0]
		}
	}
	return m
}

// InferEdges matches each target's sources against the output map to infer
// dependency edges. Returns a map from target ID to list of upstream target IDs.
// Checks target-level sources and batch item sources.
func InferEdges(f *config.Folio, producerMap map[string]string) map[string][]string {
	edges := make(map[string][]string)
	for tid, target := range f.Targets {
		// Target-level sources
		for _, src := range target.Sources {
			addEdgeIfProduced(edges, tid, src.Path, src.External, src.ID, producerMap)
		}

		// Batch item sources
		if target.Batch != nil {
			for _, item := range target.Batch.Items {
				addEdgeIfProduced(edges, tid, item.Source, "", "", producerMap)
			}
		}

	}
	return edges
}

func addEdgeIfProduced(edges map[string][]string, tid, path, external, id string, producerMap map[string]string) {
	if path != "" {
		lookupKey := fmt.Sprintf("path:%s", path)
		if producer, ok := producerMap[lookupKey]; ok && producer != tid {
			edges[tid] = append(edges[tid], producer)
		}
	} else if external != "" && id != "" {
		// Prefix match: sources don't carry field, so match any field suffix.
		prefix := fmt.Sprintf("ext:%s:%s:", external, id)
		for key, producer := range producerMap {
			if strings.HasPrefix(key, prefix) && producer != tid {
				edges[tid] = append(edges[tid], producer)
				break // one producer per external resource is sufficient
			}
		}
	}
}

// MergeEdges deduplicates inferred edges into a unified adjacency list.
func MergeEdges(f *config.Folio, inferred map[string][]string) map[string][]string {
	merged := make(map[string][]string)

	// Add inferred edges
	for tid, deps := range inferred {
		merged[tid] = append(merged[tid], deps...)
	}

	// Deduplicate
	for tid, deps := range merged {
		merged[tid] = dedupe(deps)
	}

	return merged
}

// BuildSourceDAG builds an adjacency list from depends_on fields on sources.
// Skips sources without a path or without depends_on entries.
// Returns a map consumable by DetectCycle.
func BuildSourceDAG(sources []config.Source) map[string][]string {
	adj := make(map[string][]string)
	for _, src := range sources {
		if src.Path == "" || len(src.DependsOn) == 0 {
			continue
		}
		adj[src.Path] = append(adj[src.Path], src.DependsOn...)
	}
	return adj
}

func dedupe(ss []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
