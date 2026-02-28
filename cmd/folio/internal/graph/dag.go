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

		// Tree node outputs
		if target.Tree != nil {
			WalkTree(&target.Tree.Root, func(node *config.TreeNode) {
				if node.ID != "" && target.Tree.System != "" {
					key := fmt.Sprintf("ext:%s:%s:%s", target.Tree.System, node.ID, target.Tree.Field)
					m[key] = append(m[key], tid)
				}
			})
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

// WalkTree calls fn for every node in a tree (pre-order traversal).
func WalkTree(node *config.TreeNode, fn func(*config.TreeNode)) {
	fn(node)
	for i := range node.Children {
		WalkTree(&node.Children[i], fn)
	}
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
// Checks target-level sources, batch item sources, and tree node sources.
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

		// Tree node sources
		if target.Tree != nil {
			WalkTree(&target.Tree.Root, func(node *config.TreeNode) {
				addEdgeIfProduced(edges, tid, node.File, "", "", producerMap)
			})
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

// MergeEdges combines explicit blocked_by edges with inferred edges into
// a unified adjacency list. Deduplicates edges.
func MergeEdges(f *config.Folio, inferred map[string][]string) map[string][]string {
	merged := make(map[string][]string)

	// Add explicit blocked_by
	for tid, target := range f.Targets {
		for _, dep := range target.BlockedBy {
			merged[tid] = append(merged[tid], dep)
		}
	}

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
