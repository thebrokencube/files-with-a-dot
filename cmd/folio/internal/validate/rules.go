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
	if f.Schema < 1 || f.Schema > 3 {
		r.addError("Missing or invalid schema version (expected: 1, 2, or 3, got: %d)", f.Schema)
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

	// Validate type and status fields (schema v3)
	validTypes := map[string]bool{
		"design": true, "plan": true, "track": true,
		"spike": true, "retro": true,
	}
	tier1Types := map[string]bool{
		"design": true, "plan": true, "track": true,
	}
	validStatuses := map[string]bool{
		"active": true, "done": true,
	}

	for _, src := range f.Sources {
		if src.Path == "" {
			continue // skip external sources — no lifecycle state
		}
		if src.Type != "" && !validTypes[src.Type] {
			r.addError("Source '%s': invalid type '%s' (expected: design, plan, track, spike, retro)", src.Path, src.Type)
		}
		if src.Status != "" {
			if !validStatuses[src.Status] {
				r.addError("Source '%s': invalid status '%s' (expected: active, done)", src.Path, src.Status)
			}
			if src.Type != "" && !tier1Types[src.Type] {
				r.addError("Source '%s': status is only valid on design, plan, or track types (got type '%s')", src.Path, src.Type)
			}
			if src.Type == "" {
				r.addError("Source '%s': status requires an explicit type field", src.Path)
			}
		}
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
		fullPath := config.ResolvePath(folioDir, src.Path)
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
			outDir := filepath.Dir(config.ResolvePath(folioDir, out.Path))
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

	// Batch items
	if target.Batch != nil {
		for i, item := range target.Batch.Items {
			prefix := fmt.Sprintf("Target '%s' batch item [%d]", tid, i)
			if item.ID == "" {
				r.addError("%s: missing required field: id", prefix)
			}
			if item.Source != "" {
				fullPath := config.ResolvePath(folioDir, item.Source)
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

	// Batch and forest mutual exclusion
	if target.Batch != nil && target.Forest != nil {
		r.addError("Target '%s': batch and forest are mutually exclusive", tid)
	}

	// Forest validation
	if target.Forest != nil {
		if target.Forest.Root == "" {
			r.addError("Target '%s': forest missing required field: root", tid)
		} else {
			rootPath := config.ResolvePath(folioDir, target.Forest.Root)
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

}

func (r *Result) addError(format string, args ...interface{}) {
	r.Valid = false
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

func (r *Result) addWarning(format string, args ...interface{}) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}
