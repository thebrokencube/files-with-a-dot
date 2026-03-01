package status

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/graph"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
)

// OutputStatus represents the derived status of a single output.
type OutputStatus struct {
	Type   string `json:"type"` // "local" or "external"
	Path   string `json:"path,omitempty"`
	System string `json:"system,omitempty"`
	ID     string `json:"id,omitempty"`
	Status string `json:"status"` // "clean", "stale", "missing", "unknown"
}

// BatchItemStatus holds derived status for a single batch item.
type BatchItemStatus struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	System string `json:"system"`
	ExtID  string `json:"ext_id"`
	Status string `json:"status"` // clean, stale, missing, unknown
}

// TargetStatus holds derived status for a target.
type TargetStatus struct {
	Sources    []string          `json:"sources"`
	Outputs    []OutputStatus    `json:"outputs"`
	TreeRoot   *TreeNodeStatus   `json:"tree,omitempty"`
	BatchItems []BatchItemStatus `json:"batch_items,omitempty"`
}

// TreeNodeStatus holds derived status for a single tree node.
type TreeNodeStatus struct {
	ID       string           `json:"id"`
	Label    string           `json:"label"`
	File     string           `json:"file"`
	Status   string           `json:"status"`              // clean, stale, missing
	CausedBy string           `json:"caused_by,omitempty"` // child ID that caused staleness
	Children []TreeNodeStatus `json:"children,omitempty"`
}

// SourceInfo holds classified info about a project-level source.
type SourceInfo struct {
	Kind  string `json:"kind"` // "primary", "external", "derived", "code", "unknown"
	Label string `json:"label"`
}

// ProjectStatus holds the complete derived status for a folio project.
type ProjectStatus struct {
	Project string                  `json:"project"`
	Sources []SourceInfo            `json:"sources"`
	Targets map[string]TargetStatus `json:"targets"`
	Tasks   int                     `json:"tasks"`
	Pending int                     `json:"pending"`
}

// Derive computes the full status of a folio project.
func Derive(f *config.Folio, folioDir string) *ProjectStatus {
	ps := &ProjectStatus{
		Project: f.Project,
		Targets: make(map[string]TargetStatus),
		Tasks:   len(f.Tasks),
		Pending: len(f.Pending),
	}

	// Classify project-level sources
	for _, src := range f.Sources {
		ps.Sources = append(ps.Sources, ClassifySource(src))
	}

	// Derive target statuses
	for _, tid := range maputil.SortedKeys(f.Targets) {
		target := f.Targets[tid]
		ts := TargetStatus{}

		// Collect local source paths
		var sourcePaths []string
		for _, src := range target.Sources {
			if src.Path != "" {
				sourcePaths = append(sourcePaths, src.Path)
				ts.Sources = append(ts.Sources, src.Path)
			}
		}

		// Derive status for each output
		for _, out := range target.Outputs {
			if out.External != "" {
				ts.Outputs = append(ts.Outputs, OutputStatus{
					Type:   "external",
					System: out.External,
					ID:     out.ID,
					Status: "unknown",
				})
			} else if out.Path != "" {
				status := DeriveLocalStatus(folioDir, out.Path, sourcePaths)
				ts.Outputs = append(ts.Outputs, OutputStatus{
					Type:   "local",
					Path:   out.Path,
					Status: status,
				})
			}
		}

		// Tree status derivation
		if target.Tree != nil {
			manifestMtime := getManifestMtime(folioDir, target.Outputs)
			ts.TreeRoot = deriveTreeNodeStatus(folioDir, &target.Tree.Root, manifestMtime)
		}

		// Batch item status derivation
		if target.Batch != nil {
			manifestMtime := getManifestMtime(folioDir, target.Outputs)
			for _, item := range target.Batch.Items {
				out := target.Batch.ResolveItemOutput(item)
				bis := BatchItemStatus{
					ID:     item.ID,
					Source: item.Source,
					System: out.External,
					ExtID:  out.ID,
				}
				if item.Source == "" {
					bis.Status = "unknown"
				} else {
					srcPath := filepath.Join(folioDir, item.Source)
					srcInfo, err := os.Stat(srcPath)
					if err != nil {
						bis.Status = "missing"
					} else if manifestMtime.IsZero() || srcInfo.ModTime().After(manifestMtime) {
						bis.Status = "stale"
					} else {
						bis.Status = "clean"
					}
				}
				ts.BatchItems = append(ts.BatchItems, bis)
			}
		}

		ps.Targets[tid] = ts
	}

	return ps
}

// deriveTreeNodeStatus computes status for a tree node and its children.
// Staleness is based on source mtime vs manifest mtime, with bottom-up propagation.
func deriveTreeNodeStatus(folioDir string, node *config.TreeNode, manifestMtime time.Time) *TreeNodeStatus {
	ns := &TreeNodeStatus{
		ID:    node.ID,
		Label: node.Label,
		File:  node.File,
	}

	// Derive own status: source mtime vs manifest mtime
	if node.File != "" {
		srcPath := filepath.Join(folioDir, node.File)
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			ns.Status = "missing"
		} else if manifestMtime.IsZero() || srcInfo.ModTime().After(manifestMtime) {
			ns.Status = "stale"
		} else {
			ns.Status = "clean"
		}
	} else {
		ns.Status = "unknown"
	}

	// Recurse into children
	for i := range node.Children {
		childStatus := deriveTreeNodeStatus(folioDir, &node.Children[i], manifestMtime)
		ns.Children = append(ns.Children, *childStatus)
	}

	// Bottom-up propagation: child stale/missing → parent stale
	if ns.Status == "clean" {
		for _, child := range ns.Children {
			if child.Status == "stale" || child.Status == "missing" {
				ns.Status = "stale"
				ns.CausedBy = child.ID
				break
			}
		}
	}

	return ns
}

// getManifestMtime returns the mtime of the first local output (manifest file).
// Returns zero time if no local output exists (everything will be stale).
func getManifestMtime(folioDir string, outputs []config.Output) time.Time {
	for _, out := range outputs {
		if out.Path != "" {
			info, err := os.Stat(filepath.Join(folioDir, out.Path))
			if err == nil {
				return info.ModTime()
			}
		}
	}
	return time.Time{}
}

// DeriveLocalStatus computes status for a local output by comparing mtimes.
func DeriveLocalStatus(folioDir, outputPath string, sourcePaths []string) string {
	fullOutput := filepath.Join(folioDir, outputPath)
	outInfo, err := os.Stat(fullOutput)
	if err != nil {
		return "missing"
	}
	outputMtime := outInfo.ModTime()

	for _, src := range sourcePaths {
		fullSrc := filepath.Join(folioDir, src)
		srcInfo, err := os.Stat(fullSrc)
		if err != nil {
			return "stale"
		}
		if srcInfo.ModTime().After(outputMtime) {
			return "stale"
		}
	}

	return "clean"
}

// DeriveLocalCause returns a human-readable reason why a local output is stale.
// Returns "" if clean, "output missing" if the output doesn't exist, or the
// first source path that is newer than the output.
func DeriveLocalCause(folioDir, outputPath string, sourcePaths []string) string {
	fullOutput := filepath.Join(folioDir, outputPath)
	outInfo, err := os.Stat(fullOutput)
	if err != nil {
		return "output missing"
	}
	outputMtime := outInfo.ModTime()

	for _, src := range sourcePaths {
		fullSrc := filepath.Join(folioDir, src)
		srcInfo, err := os.Stat(fullSrc)
		if err != nil {
			return fmt.Sprintf("source %s missing", src)
		}
		if srcInfo.ModTime().After(outputMtime) {
			return fmt.Sprintf("source %s newer than output", src)
		}
	}

	return ""
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
			// Use oldest cached date across all derived_from entries
			oldestAge := -1
			for _, df := range src.DerivedFrom {
				if df.Cached != "" {
					if age := daysSince(df.Cached); age >= 0 {
						if oldestAge < 0 || age > oldestAge {
							oldestAge = age
						}
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

// DeriveWithDAG computes status and applies transitive staleness propagation.
// Returns the ProjectStatus and a causedBy map (target → upstream that caused staleness).
func DeriveWithDAG(f *config.Folio, folioDir string) (*ProjectStatus, map[string]string) {
	ps := Derive(f, folioDir)

	outputMap := graph.BuildOutputMap(f)
	producerMap := graph.SingleProducerMap(outputMap)
	inferred := graph.InferEdges(f, producerMap)
	merged := graph.MergeEdges(f, inferred)

	// Build per-target worst status
	targetStatuses := make(map[string]string)
	for tid, ts := range ps.Targets {
		worst := "clean"
		for _, out := range ts.Outputs {
			worst = worseStatus(worst, out.Status)
		}
		targetStatuses[tid] = worst
	}

	propagated, causedBy := graph.PropagateStaleness(targetStatuses, merged)

	// Apply propagated statuses back to outputs
	for tid, ts := range ps.Targets {
		if propagated[tid] != targetStatuses[tid] {
			for i := range ts.Outputs {
				if ts.Outputs[i].Status == "clean" {
					ts.Outputs[i].Status = "stale"
				}
			}
			ps.Targets[tid] = ts
		}
	}

	return ps, causedBy
}

// StatusRank returns a numeric rank for a status string (higher = worse).
func StatusRank(s string) int {
	switch s {
	case "clean":
		return 0
	case "unknown":
		return 1
	case "stale":
		return 2
	case "missing":
		return 3
	default:
		return -1
	}
}

func worseStatus(a, b string) string {
	if StatusRank(b) > StatusRank(a) {
		return b
	}
	return a
}

// daysSince computes the number of days since a YYYY-MM-DD date string.
// Returns -1 if the date can't be parsed.
func daysSince(dateStr string) int {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return -1
	}
	days := time.Since(t).Hours() / 24
	return int(math.Floor(days))
}
