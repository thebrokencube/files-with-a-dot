package output

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// PrintValidateTerminal renders validation results to a terminal.
func PrintValidateTerminal(w io.Writer, r *validate.Result, folioPath string, color bool) {
	p := dendrik.NewPalette(color)

	if len(r.Errors) > 0 {
		fmt.Fprintf(w, "%s%sValidation failed%s (%d error(s))\n\n", p.Red, p.Bold, p.Reset, len(r.Errors))
		for _, err := range r.Errors {
			fmt.Fprintf(w, "  %s✗%s %s\n", p.Red, p.Reset, err)
		}
	}

	if len(r.Warnings) > 0 {
		fmt.Fprintln(w)
		for _, warn := range r.Warnings {
			fmt.Fprintf(w, "  %s!%s %s\n", p.Yellow, p.Reset, warn)
		}
	}

	if len(r.Errors) == 0 {
		fmt.Fprintf(w, "%s%sValid%s — %s\n", p.Green, p.Bold, p.Reset, folioPath)
	}
}

// PrintStatusTerminal renders project status to a terminal.
func PrintStatusTerminal(w io.Writer, ps *status.ProjectStatus, causedBy map[string]string, color bool) {
	p := dendrik.NewPalette(color)

	fmt.Fprintf(w, "%s%s%s\n", p.Bold, ps.Project, p.Reset)

	// Lifecycle summary
	ls := ps.Lifecycle
	fmt.Fprintf(w, "%sLifecycle:%s %d observations | %d spikes | %d designs | %d plans | %d retros | %d references\n",
		p.Dim, p.Reset,
		ls.Observations, ls.Spikes, ls.Designs, ls.Plans, ls.Retros, ls.References)
	fmt.Fprintln(w)

	// Project-level sources
	if len(ps.Sources) > 0 {
		fmt.Fprintln(w, "Sources:")
		for _, src := range ps.Sources {
			fmt.Fprintf(w, "  %-28s %s\n", src.Label, src.Kind)
		}
		fmt.Fprintln(w)
	}

	// Targets
	if len(ps.Targets) == 0 {
		fmt.Fprintf(w, "%sNo targets defined.%s\n", p.Dim, p.Reset)
	} else {
		fmt.Fprintln(w, "Targets:")
		for _, tid := range maputil.SortedKeys(ps.Targets) {
			ts := ps.Targets[tid]
			for _, out := range ts.Outputs {
				s := out.Status
				c := statusColor(s, p)
				var detail string
				if out.Type == "external" {
					detail = fmt.Sprintf("external: %s %s", out.System, out.ID)
				} else {
					detail = out.Path
				}

				// Annotate transitive staleness
				annotation := ""
				if cause, ok := causedBy[tid]; ok {
					annotation = fmt.Sprintf(" << %s", cause)
				}

				fmt.Fprintf(w, "  %-24s %s%s%s (%s)%s\n", tid, c, s, p.Reset, detail, annotation)
			}

			// Show sources
			if len(ts.Sources) > 0 {
				srcList := ""
				for i, s := range ts.Sources {
					if i > 0 {
						srcList += ", "
					}
					srcList += s
				}
				fmt.Fprintf(w, "  %s                        sources: %s%s\n", p.Dim, srcList, p.Reset)
			}

			// Show tree
			if ts.TreeRoot != nil {
				printTreeNode(w, ts.TreeRoot, p, "    ", true)
			}

			// Show batch items
			if len(ts.BatchItems) > 0 {
				for i, item := range ts.BatchItems {
					connector := "├── "
					if i == len(ts.BatchItems)-1 {
						connector = "└── "
					}
					c := statusColor(item.Status, p)
					fmt.Fprintf(w, "    %s%s%s%s  %s → %s:%s\n", connector, c, item.Status, p.Reset, item.ID, item.System, item.ExtID)
				}
			}
		}
	}

	// Observations
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Observations: %s%d%s\n", p.Bold, ps.Lifecycle.Observations, p.Reset)
}

func printTreeNode(w io.Writer, node *status.TreeNodeStatus, p dendrik.Palette, indent string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	label := node.ID
	if node.Label != "" {
		label = fmt.Sprintf("%s (%s)", node.ID, node.Label)
	}

	c := statusColor(node.Status, p)
	annotation := ""
	if node.CausedBy != "" {
		annotation = fmt.Sprintf(" %s<< %s%s", p.Dim, node.CausedBy, p.Reset)
	}

	fmt.Fprintf(w, "%s%s%s%s%s %s%s\n", indent, connector, c, node.Status, p.Reset, label, annotation)

	childIndent := indent + "│   "
	if isLast {
		childIndent = indent + "    "
	}

	for i := range node.Children {
		printTreeNode(w, &node.Children[i], p, childIndent, i == len(node.Children)-1)
	}
}

func statusColor(s string, p dendrik.Palette) string {
	switch s {
	case "clean":
		return p.Green
	case "stale":
		return p.Yellow
	case "missing":
		return p.Red
	case "unknown":
		return p.Yellow
	default:
		return p.Reset
	}
}

// SourceLabel returns a human-readable label for a source entry.
func SourceLabel(s config.Source) string {
	if s.Path != "" {
		return s.Path
	}
	if s.External != "" {
		if s.ID != "" {
			return s.External + ":" + s.ID
		}
		return s.External
	}
	return "(unknown)"
}

// OutputLabel returns a human-readable label for an output entry.
func OutputLabel(o config.Output) string {
	if o.Path != "" {
		return o.Path
	}
	if o.External != "" {
		if o.ID != "" {
			return o.External + ":" + o.ID
		}
		return o.External
	}
	return "(unknown)"
}

// PrintDAGTerminal renders the target dependency graph to a terminal.
func PrintDAGTerminal(w io.Writer, targets map[string]config.Target, adj map[string][]string, allTargets []string, color bool) {
	p := dendrik.NewPalette(color)

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

	for i, tid := range sorted {
		if i > 0 {
			fmt.Fprintln(w)
		}

		deps := adj[tid]
		if len(deps) == 0 {
			fmt.Fprintf(w, "%s%s%s\n", p.Bold, tid, p.Reset)
		} else {
			sortedDeps := make([]string, len(deps))
			copy(sortedDeps, deps)
			sort.Strings(sortedDeps)
			fmt.Fprintf(w, "%s%s%s -> [%s]\n", p.Bold, tid, p.Reset, strings.Join(sortedDeps, ", "))
		}

		target, ok := targets[tid]
		if !ok {
			continue
		}

		if len(target.Sources) > 0 {
			labels := make([]string, len(target.Sources))
			for j, s := range target.Sources {
				labels[j] = SourceLabel(s)
			}
			fmt.Fprintf(w, "  %ssources:%s %s\n", p.Dim, p.Reset, strings.Join(labels, ", "))
		}

		if len(target.Outputs) > 0 {
			labels := make([]string, len(target.Outputs))
			for j, o := range target.Outputs {
				labels[j] = OutputLabel(o)
			}
			fmt.Fprintf(w, "  %soutputs:%s %s\n", p.Dim, p.Reset, strings.Join(labels, ", "))
		}
	}
}

// PrintBranchDAG renders branch topology derived from targets with Branch set.
// Convenience wrapper that builds topology internally.
func PrintBranchDAG(w io.Writer, targets map[string]config.Target, color bool) {
	bt := BuildBranchTopology(targets)
	PrintBranchDAGFromTopology(w, bt, color, false)
}

// PrintBranchDAGFromTopology renders a pre-built branch topology to a terminal.
// When showStatus is true, status and stale_via annotations are included.
func PrintBranchDAGFromTopology(w io.Writer, bt *BranchTopology, color bool, showStatus bool) {
	if len(bt.Roots) == 0 {
		return
	}

	p := dendrik.NewPalette(color)

	var printNode func(node *BranchNode, indent string, isLast bool)
	printNode = func(node *BranchNode, indent string, isLast bool) {
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		statusStr := ""
		if showStatus && node.Status != "" {
			c := statusColor(node.Status, p)
			statusStr = fmt.Sprintf("  %s%s%s", c, node.Status, p.Reset)
			if node.StaleVia != "" {
				statusStr += fmt.Sprintf(" %s<< %s%s", p.Dim, node.StaleVia, p.Reset)
			}
		}

		fmt.Fprintf(w, "%s%s%s%s%s%s\n", indent, connector, p.Bold, node.ID, p.Reset, statusStr)

		prStr := ""
		if node.PR != "" {
			prStr = fmt.Sprintf("  PR: %s", node.PR)
		}
		fmt.Fprintf(w, "%s    %sbranch: %s (base: %s)%s%s\n", indent, p.Dim, node.Branch, node.Base, prStr, p.Reset)

		childIndent := indent + "│   "
		if isLast {
			childIndent = indent + "    "
		}

		for i, kid := range node.Children {
			printNode(kid, childIndent, i == len(node.Children)-1)
		}
	}

	for _, root := range bt.Roots {
		fmt.Fprintf(w, "%s%s%s\n", p.Bold, root.Base, p.Reset)
		for i, kid := range root.Children {
			printNode(kid, "", i == len(root.Children)-1)
		}
	}
}

// StaleEntry holds display data for a single stale target.
type StaleEntry struct {
	ID      string   `json:"id"`
	Status  string   `json:"status"`
	Branch  string   `json:"branch,omitempty"`
	Outputs []string `json:"outputs"`
	Cause   string   `json:"cause"`
}

// PrintStaleTerminal renders stale targets to a terminal.
func PrintStaleTerminal(w io.Writer, entries []StaleEntry, color bool) {
	p := dendrik.NewPalette(color)

	for i, e := range entries {
		if i > 0 {
			fmt.Fprintln(w)
		}

		c := statusColor(e.Status, p)
		fmt.Fprintf(w, "%s%s%s  %s%s%s\n", p.Bold, e.ID, p.Reset, c, e.Status, p.Reset)

		for _, out := range e.Outputs {
			fmt.Fprintf(w, "  %soutput:%s %s\n", p.Dim, p.Reset, out)
		}

		if e.Cause != "" {
			fmt.Fprintf(w, "  %scause:%s %s\n", p.Dim, p.Reset, e.Cause)
		}
	}
}
