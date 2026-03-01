package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/graph"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
)

// Result holds the outcome of validation.
type Result struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// Validate runs all validation rules against a parsed Folio and its directory.
func Validate(f *config.Folio, folioDir string) *Result {
	r := &Result{Valid: true}

	// Schema version
	if f.Schema != 1 {
		r.addError("Missing or invalid schema version (expected: 1, got: %d)", f.Schema)
	}

	// Project name
	if f.Project == "" {
		r.addError("Missing required field: project")
	}

	// Deprecated context_sources
	if f.ContextSources != nil {
		r.addWarning("Deprecated key 'context_sources' found — migrate to 'sources'")
	}

	// Project-level sources
	for i, src := range f.Sources {
		validateSource(r, src, fmt.Sprintf("Project source [%d]", i), folioDir)
	}

	// Targets
	for _, tid := range maputil.SortedKeys(f.Targets) {
		target := f.Targets[tid]
		validateTarget(r, f, tid, &target, folioDir)
	}

	// Output map collisions
	outputMap := graph.BuildOutputMap(f)
	for _, key := range maputil.SortedKeys(outputMap) {
		producers := outputMap[key]
		if len(producers) > 1 {
			r.addError("Output collision: '%s' produced by multiple targets: %s", key, strings.Join(producers, ", "))
		}
	}

	// Edge inference + cycle detection
	singleProducerMap := graph.SingleProducerMap(outputMap)
	inferred := graph.InferEdges(f, singleProducerMap)
	merged := graph.MergeEdges(f, inferred)

	if cycle := graph.DetectCycle(merged); cycle != nil {
		r.addError("Dependency cycle detected: %s", strings.Join(cycle, " -> "))
	}

	return r
}

func validateSource(r *Result, src config.Source, prefix string, folioDir string) {
	if src.External != "" && src.Path != "" {
		r.addWarning("%s: source has both 'path' and 'external' set — path is ignored for external sources", prefix)
	}

	if src.External != "" {
		if src.ID == "" {
			r.addError("%s: external source '%s' missing required field: id", prefix, src.External)
		}
	} else if src.Path != "" {
		fullPath := filepath.Join(folioDir, src.Path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			r.addError("%s: file not found: %s", prefix, src.Path)
		}
		for j, df := range src.DerivedFrom {
			if df.External == "" {
				r.addError("%s: derived_from[%d] missing required field: external", prefix, j)
			}
		}
	} else {
		r.addError("%s: source must have either 'path' or 'external' set", prefix)
	}
}

func validateTarget(r *Result, f *config.Folio, tid string, target *config.Target, folioDir string) {
	// Target sources
	for _, src := range target.Sources {
		validateSource(r, src, fmt.Sprintf("Target '%s'", tid), folioDir)
	}

	// Output paths and external fields
	hasExternal := false
	hasLocal := false
	for _, out := range target.Outputs {
		if out.Path != "" {
			hasLocal = true
			outDir := filepath.Dir(filepath.Join(folioDir, out.Path))
			if info, err := os.Stat(outDir); err != nil || !info.IsDir() {
				r.addError("Target '%s': output parent directory not found: %s", tid, filepath.Dir(out.Path))
			}
		}
		if out.External != "" {
			hasExternal = true
			if out.ID == "" {
				r.addError("Target '%s': external output '%s' missing required field: id", tid, out.External)
			}
		}
	}

	// blocked_by references exist
	for _, dep := range target.BlockedBy {
		if _, exists := f.Targets[dep]; !exists {
			r.addError("Target '%s': blocked_by references non-existent target: %s", tid, dep)
		}
	}

	// Batch items
	if target.Batch != nil {
		for i, item := range target.Batch.Items {
			prefix := fmt.Sprintf("Target '%s' batch item [%d]", tid, i)
			if item.ID == "" {
				r.addError("%s: missing required field: id", prefix)
			}
			if item.Source != "" {
				fullPath := filepath.Join(folioDir, item.Source)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					r.addError("Target '%s': batch item source not found: %s", tid, item.Source)
				}
			}
			out := target.Batch.ResolveItemOutput(item)
			if out.External == "" {
				r.addError("%s: resolved output missing required field: external (set on item or batch-level system)", prefix)
			}
			if out.ID == "" {
				r.addError("%s: resolved output missing required field: id", prefix)
			}
		}
	}

	// Tree and batch mutual exclusion
	if target.Tree != nil && target.Batch != nil {
		r.addError("Target '%s': tree and batch are mutually exclusive", tid)
	}

	// Tree validation
	if target.Tree != nil {
		if target.Tree.System == "" {
			r.addError("Target '%s': tree missing required field: system", tid)
		}
		seenIDs := make(map[string]bool)
		validateTreeNode(r, tid, &target.Tree.Root, target, folioDir, seenIDs)
	}

	// Transform field (required)
	if target.Transform == "" {
		r.addError("Target '%s': missing required field: transform", tid)
	} else if !config.ValidTransforms[target.Transform] {
		r.addError("Target '%s': invalid transform '%s' (must be: distill, extract, adapt, compose)", tid, target.Transform)
	}

	// Precompile rule: external outputs require a local path sibling
	if hasExternal && !hasLocal {
		r.addError("Target '%s': precompile rule — external outputs require a local path: sibling for review", tid)
	}
}

func validateTreeNode(r *Result, tid string, node *config.TreeNode, target *config.Target, folioDir string, seenIDs map[string]bool) {
	prefix := fmt.Sprintf("Target '%s' tree node", tid)
	if node.ID != "" {
		prefix = fmt.Sprintf("Target '%s' tree node '%s'", tid, node.ID)
	}

	// ID required and unique within tree
	if node.ID == "" {
		r.addError("%s: missing required field: id", prefix)
	} else if seenIDs[node.ID] {
		r.addError("%s: duplicate tree node ID", prefix)
	}
	seenIDs[node.ID] = true

	// File optional; if present, must exist
	if node.File != "" {
		fullPath := filepath.Join(folioDir, node.File)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			r.addError("%s: file not found: %s", prefix, node.File)
		}
	}

	// Transform override must be valid if present
	if node.Transform != "" && !config.ValidTransforms[node.Transform] {
		r.addError("%s: invalid transform '%s'", prefix, node.Transform)
	}

	// Recurse into children
	for i := range node.Children {
		validateTreeNode(r, tid, &node.Children[i], target, folioDir, seenIDs)
	}
}

func (r *Result) addError(format string, args ...interface{}) {
	r.Valid = false
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

func (r *Result) addWarning(format string, args ...interface{}) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

