package validate

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/graph"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/observe"
)

// ValidationMode controls checks that are safe for reads versus mutations.
type ValidationMode uint8

const (
	ReadOnly ValidationMode = iota
	Mutation
)

// FindingSeverity identifies whether a finding blocks a mutation.
type FindingSeverity string

const (
	FindingError   FindingSeverity = "error"
	FindingWarning FindingSeverity = "warning"
)

// Finding is a stable validation result used to compare projected manifests.
type Finding struct {
	Code     string          `json:"-"`
	Subject  string          `json:"-"`
	Severity FindingSeverity `json:"-"`
	Message  string          `json:"-"`
}

// Result holds the outcome of validation.
type Result struct {
	Valid    bool      `json:"valid"`
	Errors   []string  `json:"errors"`
	Warnings []string  `json:"warnings"`
	Findings []Finding `json:"-"`
}

// Validate runs structural, reference, graph, repository, cross-reference,
// and observation checks using the caller-resolved context.
func Validate(f *config.Folio, ctx config.Context, mode ValidationMode) *Result {
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
		validateSource(r, src, fmt.Sprintf("Project source [%d]", i), ctx, true)
	}
	seenProjectSourcePaths := make(map[string]bool)
	for i, src := range f.Sources {
		if src.Path == "" {
			continue
		}
		if seenProjectSourcePaths[src.Path] {
			r.addError("Project source [%d]: duplicate path: %s", i, src.Path)
			continue
		}
		seenProjectSourcePaths[src.Path] = true
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
		validateTarget(r, f, tid, &target, ctx)
	}
	validateSnapshotFields(r, f, ctx)

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

	// Observation lint is the single source for observation findings.
	for _, issue := range observe.Lint(f.Observations, ctx) {
		finding := Finding{
			Code:     issue.Code,
			Subject:  issue.Subject,
			Severity: FindingError,
			Message:  fmt.Sprintf("Observation #%d: %s", issue.Index, issue.Reason),
		}
		if issue.Severity == observe.SeverityWarning {
			finding.Severity = FindingWarning
		}
		r.addFinding(finding)
	}

	return r
}

func validateSource(r *Result, src config.Source, prefix string, ctx config.Context, isProjectLevel bool) {
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
		fullPath, err := ctx.ResolvePath(src.Path)
		if err != nil {
			r.addError("%s: %s", prefix, err)
		} else if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
			// External stores are read-only and may be uncloned — a missing
			// target warns rather than fails the whole validation.
			if config.IsExternalStorePath(src.Path, ctx.Registry) {
				r.addWarning("%s: external source not found (store may be uncloned): %s", prefix, src.Path)
			} else {
				r.addError("%s: file not found: %s", prefix, src.Path)
			}
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

func validateTarget(r *Result, f *config.Folio, tid string, target *config.Target, ctx config.Context) {
	// Target sources
	for _, src := range target.Sources {
		validateSource(r, src, fmt.Sprintf("Target '%s'", tid), ctx, false)
	}

	// Output paths and external fields
	hasExternal := false
	hasLocal := false
	for _, out := range target.Outputs {
		if out.Path != "" {
			hasLocal = true
			resolved, err := ctx.ResolvePath(out.Path)
			if err != nil {
				r.addError("Target '%s': %s", tid, err)
			} else if info, statErr := os.Stat(filepath.Dir(resolved)); statErr != nil || !info.IsDir() {
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
				fullPath, err := ctx.ResolvePath(item.Source)
				if err != nil {
					r.addError("Target '%s': %s", tid, err)
				} else if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
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
			rootPath, err := ctx.ResolvePath(target.Forest.Root)
			if err != nil {
				r.addError("Target '%s': %s", tid, err)
			} else if info, statErr := os.Stat(rootPath); statErr != nil || !info.IsDir() {
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
func validateSnapshotFields(r *Result, f *config.Folio, ctx config.Context) {
	for _, tid := range maputil.SortedKeys(f.Targets) {
		target := f.Targets[tid]
		hasComposedAt := target.ComposedAt != ""
		hasDigest := target.InputsSHA256 != ""
		if hasComposedAt != hasDigest {
			r.addError("Target '%s': composed_at and inputs_sha256 must appear as a pair", tid)
		}
		if hasComposedAt && hasDigest {
			if !validCompositionTimestamp(target.ComposedAt) {
				r.addError("Target '%s': composed_at must be a UTC RFC 3339 value", tid)
			}
			if !validDigest(target.InputsSHA256) {
				r.addError("Target '%s': inputs_sha256 must be a lowercase 64-character SHA-256 value", tid)
			}
		}
		if target.Batch != nil {
			for i, item := range target.Batch.Items {
				if item.InputsSHA256 == "" {
					continue
				}
				if item.Source == "" {
					r.addError("Target '%s' batch item [%d]: inputs_sha256 requires a source", tid, i)
				}
				if !validDigest(item.InputsSHA256) {
					r.addError("Target '%s' batch item [%d]: inputs_sha256 must be a lowercase 64-character SHA-256 value", tid, i)
				}
			}
		}
		if !target.Final {
			continue
		}
		if !hasComposedAt || !hasDigest {
			r.addError("Target '%s': final requires a complete composition snapshot", tid)
		}
		if target.Batch != nil {
			r.addError("Target '%s': final is not valid for batch targets", tid)
		}
		if target.Forest != nil {
			r.addError("Target '%s': final is not valid for forest targets", tid)
		}
		hasLocal, hasExternal := false, false
		for _, output := range target.Outputs {
			switch {
			case output.Path != "":
				hasLocal = true
				resolved, err := ctx.ResolvePath(output.Path)
				if err != nil {
					continue
				}
				info, statErr := os.Stat(resolved)
				if statErr != nil || !info.Mode().IsRegular() {
					r.addError("Target '%s': final requires existing local review copy: %s", tid, output.Path)
				}
			case output.External != "":
				hasExternal = true
			}
		}
		if !hasLocal {
			r.addError("Target '%s': final requires a local review-copy output", tid)
		}
		if !hasExternal {
			r.addError("Target '%s': final requires a direct external output", tid)
		}
	}
}

func validCompositionTimestamp(value string) bool {
	if !strings.HasSuffix(value, "Z") {
		return false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return err == nil && parsed.Location() == time.UTC
}

func validDigest(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func (r *Result) addError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	r.addFinding(Finding{
		Code:     findingCode(message),
		Subject:  strings.TrimSpace(message),
		Severity: FindingError,
		Message:  message,
	})
}

func (r *Result) addWarning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	r.addFinding(Finding{
		Code:     findingCode(message),
		Subject:  strings.TrimSpace(message),
		Severity: FindingWarning,
		Message:  message,
	})
}

func (r *Result) addFinding(finding Finding) {
	r.Findings = append(r.Findings, finding)
	if finding.Severity == FindingError {
		r.Valid = false
		r.Errors = append(r.Errors, finding.Message)
		return
	}
	r.Warnings = append(r.Warnings, finding.Message)
}

func findingCode(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.HasPrefix(lower, "observation #"):
		return "observation"
	case strings.HasPrefix(lower, "project source"), strings.HasPrefix(lower, "source "):
		return "source"
	case strings.HasPrefix(lower, "target "):
		return "target"
	case strings.HasPrefix(lower, "output collision"):
		return "output-collision"
	case strings.Contains(lower, "dependency cycle"):
		return "dependency-cycle"
	case strings.HasPrefix(lower, "repository "):
		return "repository"
	case strings.HasPrefix(lower, "cross-reference"):
		return "cross-reference"
	case strings.Contains(lower, "schema"):
		return "schema"
	default:
		return "validation"
	}
}

// Delta validates the projected result against a baseline. Blocking findings
// that already existed are advisory until the caller fixes the baseline.
func Delta(before, after *Result, subjectAliases map[string]string) *Result {
	result := &Result{Valid: true}
	baseline := make(map[string]int)
	for _, finding := range findingsOf(before) {
		if finding.Severity == FindingError {
			baseline[findingKey(finding, subjectAliases)]++
		}
	}
	for _, finding := range findingsOf(after) {
		current := finding
		if current.Severity == FindingError {
			key := findingKey(current, subjectAliases)
			if baseline[key] > 0 {
				baseline[key]--
				current.Severity = FindingWarning
			}
		}
		result.addFinding(current)
	}
	return result
}

func findingsOf(result *Result) []Finding {
	if result == nil {
		return nil
	}
	if len(result.Findings) > 0 {
		return result.Findings
	}
	var findings []Finding
	for _, message := range result.Errors {
		findings = append(findings, Finding{
			Code:     findingCode(message),
			Subject:  strings.TrimSpace(message),
			Severity: FindingError,
			Message:  message,
		})
	}
	for _, message := range result.Warnings {
		findings = append(findings, Finding{
			Code:     findingCode(message),
			Subject:  strings.TrimSpace(message),
			Severity: FindingWarning,
			Message:  message,
		})
	}
	return findings
}

func findingKey(finding Finding, aliases map[string]string) string {
	subject := finding.Subject
	for from, to := range aliases {
		subject = strings.ReplaceAll(subject, from, to)
	}
	return finding.Code + "\x00" + subject
}
