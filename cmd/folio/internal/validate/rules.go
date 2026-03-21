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
	if f.Schema != 1 && f.Schema != 2 {
		r.addError("Missing or invalid schema version (expected: 1 or 2, got: %d)", f.Schema)
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
		validateSource(r, src, fmt.Sprintf("Project source [%d]", i), folioDir, true)
	}

	// Source depends_on validation
	sourcePaths := make(map[string]bool)
	for _, src := range f.Sources {
		if src.Path != "" {
			sourcePaths[src.Path] = true
		}
	}
	for _, src := range f.Sources {
		for _, dep := range src.DependsOn {
			if !sourcePaths[dep] {
				r.addError("Source '%s': depends_on references non-existent source: %s", src.Path, dep)
			}
		}
	}
	if sourceDAG := graph.BuildSourceDAG(f.Sources); len(sourceDAG) > 0 {
		if cycle := graph.DetectCycle(sourceDAG); cycle != nil {
			r.addError("Source dependency cycle detected: %s", strings.Join(cycle, " -> "))
		}
	}

	// Targets
	for _, tid := range maputil.SortedKeys(f.Targets) {
		target := f.Targets[tid]
		validateTarget(r, f, tid, &target, folioDir)
	}

	// Duplicate branch values
	branchUsers := make(map[string][]string)
	for _, tid := range maputil.SortedKeys(f.Targets) {
		target := f.Targets[tid]
		if target.Branch != "" {
			branchUsers[target.Branch] = append(branchUsers[target.Branch], tid)
		}
	}
	for _, branch := range maputil.SortedKeys(branchUsers) {
		users := branchUsers[branch]
		if len(users) > 1 {
			r.addError("Duplicate branch '%s' used by targets: %s", branch, strings.Join(users, ", "))
		}
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

	// Repositories
	for _, name := range maputil.SortedKeys(f.Repositories) {
		url := f.Repositories[name]
		if url == "" {
			r.addError("Repository '%s': URL template is empty", name)
		} else if !strings.Contains(url, "{path}") {
			r.addWarning("Repository '%s': URL template missing {path} placeholder", name)
		}
	}

	// Cross-references
	seenFacts := make(map[string]bool)
	for i, xref := range f.CrossReferences {
		prefix := fmt.Sprintf("Cross-reference [%d]", i)
		if xref.Fact == "" {
			r.addError("%s: missing required field: fact", prefix)
		} else if seenFacts[xref.Fact] {
			r.addWarning("%s: duplicate fact '%s'", prefix, xref.Fact)
		}
		seenFacts[xref.Fact] = true
		if xref.SourceOfTruth == "" {
			r.addError("%s: missing required field: source_of_truth", prefix)
		}
	}

	return r
}

func validateSource(r *Result, src config.Source, prefix string, folioDir string, isProjectLevel bool) {
	if src.External != "" && src.Path != "" {
		r.addWarning("%s: source has both 'path' and 'external' set — path is ignored for external sources", prefix)
	}

	if src.External != "" {
		if src.ID == "" {
			r.addError("%s: external source '%s' missing required field: id", prefix, src.External)
		}
		if len(src.DependsOn) > 0 {
			r.addError("%s: depends_on is only valid on local path sources", prefix)
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
		if !isProjectLevel && len(src.DependsOn) > 0 {
			r.addWarning("%s: depends_on on target-level sources is ignored — move to project-level sources", prefix)
		}
	} else {
		r.addError("%s: source must have either 'path' or 'external' set", prefix)
	}
}

func validateTarget(r *Result, f *config.Folio, tid string, target *config.Target, folioDir string) {
	// Target sources
	for _, src := range target.Sources {
		validateSource(r, src, fmt.Sprintf("Target '%s'", tid), folioDir, false)
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

	// Tree, batch, and forest mutual exclusion
	typeCount := 0
	if target.Tree != nil {
		typeCount++
	}
	if target.Batch != nil {
		typeCount++
	}
	if target.Forest != nil {
		typeCount++
	}
	if typeCount > 1 {
		r.addError("Target '%s': tree, batch, and forest are mutually exclusive", tid)
	}

	// Tree validation
	if target.Tree != nil {
		if target.Tree.System == "" {
			r.addError("Target '%s': tree missing required field: system", tid)
		}
		seenIDs := make(map[string]bool)
		validateTreeNode(r, tid, &target.Tree.Root, target, folioDir, seenIDs)
	}

	// Forest validation
	if target.Forest != nil {
		if target.Forest.Root == "" {
			r.addError("Target '%s': forest missing required field: root", tid)
		} else {
			rootPath := filepath.Join(folioDir, target.Forest.Root)
			if info, err := os.Stat(rootPath); err != nil || !info.IsDir() {
				r.addError("Target '%s': forest root directory not found: %s", tid, target.Forest.Root)
			}
		}
	}

	// How field (optional with warning, instructions fallback)
	if target.How == "" && target.Instructions == "" {
		r.addWarning("[workflow] Target '%s': missing 'how' field — target is a data declaration only (cannot be composed)", tid)
	} else if target.How == "" && target.Instructions != "" {
		r.addWarning("Target '%s': 'instructions' is deprecated, rename to 'how'", tid)
	} else if target.How != "" && target.Instructions != "" {
		r.addError("Target '%s': has both 'how' and 'instructions' — remove 'instructions'", tid)
	}
	if target.Transform != "" {
		r.addWarning("Target '%s': 'transform' is deprecated and ignored — remove it", tid)
	}

	// Precompile rule: external outputs require a local path sibling
	if hasExternal && !hasLocal {
		r.addError("Target '%s': precompile rule — external outputs require a local path: sibling for review", tid)
	}

	// PR requires branch
	if target.PR != "" && target.Branch == "" {
		r.addError("Target '%s': pr set without branch — PR requires a branch mapping", tid)
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

	// Deprecated fields
	if node.Transform != "" {
		r.addWarning("%s: 'transform' is deprecated and ignored — remove it", prefix)
	}
	if node.Instructions != "" {
		r.addWarning("%s: 'instructions' is deprecated, rename to 'how'", prefix)
	}

	// compiled_ext mismatch: if tree has compiled_ext, check node files match
	if target.Tree != nil && target.Tree.CompiledExt != "" && node.File != "" {
		ext := filepath.Ext(node.File)
		if ext != "" && ext != target.Tree.CompiledExt {
			r.addWarning("%s: file extension %q doesn't match tree compiled_ext %q", prefix, ext, target.Tree.CompiledExt)
		}
	}

	// Sync mode must be valid if present
	if node.Sync != "" && !config.ValidSyncModes[node.Sync] {
		r.addError("%s: invalid sync mode '%s' (must be: push, pull, both)", prefix, node.Sync)
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
