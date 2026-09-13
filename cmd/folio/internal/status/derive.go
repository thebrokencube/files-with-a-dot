package status

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/graph"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/taxonomy"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/touch"
)

// OutputStatus represents the derived status of a single output.
type OutputStatus struct {
	Type   string `json:"type"`
	Path   string `json:"path,omitempty"`
	System string `json:"system,omitempty"`
	ID     string `json:"id,omitempty"`
	Status string `json:"status"`
	Cause  string `json:"cause,omitempty"`
}

// BatchItemStatus holds derived status for a single batch item.
type BatchItemStatus struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	System string `json:"system"`
	ExtID  string `json:"ext_id"`
	Status string `json:"status"`
}

// TargetStatus holds derived status for a target.
type TargetStatus struct {
	Sources    []string          `json:"sources"`
	Outputs    []OutputStatus    `json:"outputs"`
	BatchItems []BatchItemStatus `json:"batch_items,omitempty"`
	Final      bool              `json:"final,omitempty"`
}

// SourceInfo holds classified info about a project-level source.
type SourceInfo struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

// LifecycleSummary counts artifacts by lifecycle stage.
type LifecycleSummary struct {
	Observations int `json:"observations"`
	Spikes       int `json:"spikes"`
	Designs      int `json:"designs"`
	Plans        int `json:"plans"`
	Retros       int `json:"retros"`
	References   int `json:"references"`
}

// ProjectStatus holds the complete derived status for a folio project.
type ProjectStatus struct {
	Project   string                  `json:"project"`
	Lifecycle LifecycleSummary        `json:"lifecycle"`
	Sources   []SourceInfo            `json:"sources"`
	Targets   map[string]TargetStatus `json:"targets"`
}

// DeriveLifecycleSummary counts sources by lifecycle stage.
func DeriveLifecycleSummary(f *config.Folio) LifecycleSummary {
	ls := LifecycleSummary{Observations: len(f.Observations)}
	for _, src := range f.Sources {
		if src.Path == "" {
			continue
		}
		t := taxonomy.InferType(src.Path)
		switch taxonomy.StageForType(t) {
		case taxonomy.StageSpike:
			ls.Spikes++
		case taxonomy.StageDesign:
			ls.Designs++
		case taxonomy.StagePlan:
			ls.Plans++
		case taxonomy.StageRetro:
			ls.Retros++
		case taxonomy.StageReference:
			ls.References++
		}
	}
	return ls
}

// Derive computes project status using a legacy local context.
func Derive(f *config.Folio, folioDir string) *ProjectStatus {
	return DeriveWithContext(f, legacyContext(folioDir))
}

// DeriveWithContext computes project status from authored composition state.
func DeriveWithContext(f *config.Folio, ctx config.Context) *ProjectStatus {
	ps := &ProjectStatus{
		Project:   f.Project,
		Lifecycle: DeriveLifecycleSummary(f),
		Targets:   make(map[string]TargetStatus),
	}
	for _, src := range f.Sources {
		ps.Sources = append(ps.Sources, ClassifySource(src))
	}

	for _, tid := range maputil.SortedKeys(f.Targets) {
		target := f.Targets[tid]
		ts := TargetStatus{Final: target.Final}
		for _, src := range target.Sources {
			if src.Path != "" {
				ts.Sources = append(ts.Sources, src.Path)
			}
		}
		for _, output := range target.Outputs {
			if output.External != "" {
				ts.Outputs = append(ts.Outputs, OutputStatus{
					Type:   "external",
					System: output.External,
					ID:     output.ID,
					Status: "unknown",
				})
				continue
			}
			if output.Path == "" {
				continue
			}
			freshness, cause := DeriveLocalStatus(ctx, &target, output)
			ts.Outputs = append(ts.Outputs, OutputStatus{
				Type:   "local",
				Path:   output.Path,
				Status: freshness,
				Cause:  cause,
			})
		}
		if target.Batch != nil {
			for _, item := range target.Batch.Items {
				output := target.Batch.ResolveItemOutput(item)
				itemStatus := BatchItemStatus{
					ID:     item.ID,
					Source: item.Source,
					System: output.External,
					ExtID:  output.ID,
					Status: "unknown",
				}
				if item.Source != "" {
					itemStatus.Status = deriveBatchItemStatus(ctx, item)
				}
				ts.BatchItems = append(ts.BatchItems, itemStatus)
			}
		}
		ps.Targets[tid] = ts
	}
	return ps
}

func legacyContext(folioDir string) config.Context {
	return config.Context{FolioPath: filepath.Join(folioDir, "folio.yml"), WorkRoot: folioDir}
}

// DeriveLocalStatus derives one local output state and its human-readable cause.
func DeriveLocalStatus(ctx config.Context, target *config.Target, output config.Output) (string, string) {
	if target == nil || output.Path == "" {
		return "unknown", "composition not recorded"
	}
	if target.Forest != nil {
		return "unknown", "freshness delegated to jf"
	}
	fullOutput, err := ctx.ResolvePath(output.Path)
	if err != nil {
		return "missing", "output missing"
	}
	if info, statErr := os.Stat(fullOutput); statErr != nil || !info.Mode().IsRegular() {
		return "missing", "output missing"
	}
	if hasExternalInput(target) {
		return "unknown", "external source freshness unverified"
	}
	if target.ComposedAt == "" && target.InputsSHA256 == "" {
		return "unknown", "composition not recorded"
	}
	if !completeStamp(target) {
		return "unknown", "incomplete composition snapshot"
	}
	digest, err := touch.InputDigest(ctx, target)
	if err != nil {
		return "stale", missingInputCause(ctx, target)
	}
	if digest != target.InputsSHA256 {
		return "stale", "source content changed since compose"
	}
	return "clean", ""
}

func deriveBatchItemStatus(ctx config.Context, item config.BatchItem) string {
	if item.InputsSHA256 == "" {
		return "unknown"
	}
	if _, err := ctx.ResolvePath(item.Source); err != nil {
		return "missing"
	}
	if _, err := os.Stat(mustResolve(ctx, item.Source)); err != nil {
		return "missing"
	}
	digest, err := touch.ItemInputDigest(ctx, item.Source)
	if err != nil {
		return "stale"
	}
	if digest != item.InputsSHA256 {
		return "stale"
	}
	return "clean"
}

func mustResolve(ctx config.Context, path string) string {
	resolved, err := ctx.ResolvePath(path)
	if err != nil {
		return ""
	}
	return resolved
}

func hasExternalInput(target *config.Target) bool {
	for _, source := range target.Sources {
		if source.External != "" {
			return true
		}
	}
	return false
}

func completeStamp(target *config.Target) bool {
	if target.ComposedAt == "" || target.InputsSHA256 == "" || !strings.HasSuffix(target.ComposedAt, "Z") {
		return false
	}
	if _, err := time.Parse(time.RFC3339, target.ComposedAt); err != nil {
		return false
	}
	if len(target.InputsSHA256) != 64 || target.InputsSHA256 != strings.ToLower(target.InputsSHA256) {
		return false
	}
	for _, value := range target.InputsSHA256 {
		if (value < '0' || value > '9') && (value < 'a' || value > 'f') {
			return false
		}
	}
	return true
}

func missingInputCause(ctx config.Context, target *config.Target) string {
	for _, source := range target.Sources {
		if source.Path == "" || source.External != "" {
			continue
		}
		if pathMissing(ctx, source.Path) {
			return fmt.Sprintf("source %s missing", source.Path)
		}
	}
	if target.Batch != nil {
		for _, item := range target.Batch.Items {
			if item.Source != "" && pathMissing(ctx, item.Source) {
				return fmt.Sprintf("source %s missing", item.Source)
			}
		}
	}
	return "source content unavailable"
}

func pathMissing(ctx config.Context, logicalPath string) bool {
	resolved, err := ctx.ResolvePath(logicalPath)
	if err != nil {
		return true
	}
	_, err = os.Stat(resolved)
	return err != nil
}

// ClassifySource categorizes a project-level source entry.
func ClassifySource(src config.Source) SourceInfo {
	if src.External != "" {
		kind := "external"
		if src.External == "github" {
			kind = "code"
		}
		label := src.External
		if src.ID != "" {
			label += " " + src.ID
		}
		return SourceInfo{Kind: kind, Label: label}
	}
	if src.Path != "" {
		if len(src.DerivedFrom) > 0 {
			label := src.Path
			oldestAge := -1
			for _, df := range src.DerivedFrom {
				if df.Cached != "" {
					if age := daysSince(df.Cached); age >= 0 && (oldestAge < 0 || age > oldestAge) {
						oldestAge = age
					}
				}
			}
			if oldestAge >= 0 {
				label += fmt.Sprintf(" (cached %dd ago)", oldestAge)
			}
			return SourceInfo{Kind: "derived", Label: label}
		}
		return SourceInfo{Kind: "primary", Label: src.Path}
	}
	return SourceInfo{Kind: "unknown", Label: "(unrecognized entry)"}
}

// DeriveWithDAG computes status and transitive freshness causes.
func DeriveWithDAG(f *config.Folio, folioDir string) (*ProjectStatus, map[string]string, map[string]string) {
	return DeriveWithDAGWithContext(f, legacyContext(folioDir))
}

// DeriveWithDAGWithContext computes status and terminal-aware propagation.
func DeriveWithDAGWithContext(f *config.Folio, ctx config.Context) (*ProjectStatus, map[string]string, map[string]string) {
	ps := DeriveWithContext(f, ctx)
	outputMap := graph.BuildOutputMap(f)
	producerMap := graph.SingleProducerMap(outputMap)
	merged := graph.MergeEdges(f, graph.InferEdges(f, producerMap))

	localStatuses := make(map[string]string, len(ps.Targets))
	terminal := make(map[string]bool, len(ps.Targets))
	for tid, target := range f.Targets {
		terminal[tid] = target.Final
		if localStatus, ok := targetStatus(ps.Targets[tid]); ok {
			localStatuses[tid] = localStatus
		}
	}
	propagated, causedBy := graph.PropagateStaleness(localStatuses, merged, terminal)
	causedStatus := make(map[string]string, len(causedBy))
	for tid, upstream := range causedBy {
		causedStatus[tid] = propagated[upstream]
	}

	for tid, targetStatus := range ps.Targets {
		if propagated[tid] == localStatuses[tid] || terminal[tid] {
			continue
		}
		for i := range targetStatus.Outputs {
			if targetStatus.Outputs[i].Type != "local" {
				continue
			}
			if graph.StatusRank(propagated[tid]) > graph.StatusRank(targetStatus.Outputs[i].Status) {
				targetStatus.Outputs[i].Status = propagated[tid]
				if upstream := causedBy[tid]; upstream != "" {
					targetStatus.Outputs[i].Cause = fmt.Sprintf("blocked by target %s", upstream)
				}
			}
		}
		ps.Targets[tid] = targetStatus
	}
	return ps, causedBy, causedStatus
}

func targetStatus(targetStatus TargetStatus) (string, bool) {
	worst := "clean"
	hasLocal := false
	for _, output := range targetStatus.Outputs {
		if output.Type != "local" {
			continue
		}
		hasLocal = true
		if graph.StatusRank(output.Status) > graph.StatusRank(worst) {
			worst = output.Status
		}
	}
	if !hasLocal {
		return "", false
	}
	return worst, true
}

// daysSince computes the number of days since a YYYY-MM-DD date string.
func daysSince(dateStr string) int {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return -1
	}
	days := time.Since(t).Hours() / 24
	return int(math.Floor(days))
}
