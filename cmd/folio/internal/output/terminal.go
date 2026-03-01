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
)

const (
	ansiRed    = "\033[0;31m"
	ansiGreen  = "\033[0;32m"
	ansiYellow = "\033[0;33m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiReset  = "\033[0m"
)

type palette struct {
	red, green, yellow, bold, dim, reset string
}

func newPalette(color bool) palette {
	if !color {
		return palette{}
	}
	return palette{
		red:    ansiRed,
		green:  ansiGreen,
		yellow: ansiYellow,
		bold:   ansiBold,
		dim:    ansiDim,
		reset:  ansiReset,
	}
}

// PrintValidateTerminal renders validation results to a terminal.
func PrintValidateTerminal(w io.Writer, r *validate.Result, folioPath string, color bool) {
	p := newPalette(color)

	if len(r.Errors) > 0 {
		fmt.Fprintf(w, "%s%sValidation failed%s (%d error(s))\n\n", p.red, p.bold, p.reset, len(r.Errors))
		for _, err := range r.Errors {
			fmt.Fprintf(w, "  %s✗%s %s\n", p.red, p.reset, err)
		}
	}

	if len(r.Warnings) > 0 {
		fmt.Fprintln(w)
		for _, warn := range r.Warnings {
			fmt.Fprintf(w, "  %s!%s %s\n", p.yellow, p.reset, warn)
		}
	}

	if len(r.Errors) == 0 {
		fmt.Fprintf(w, "%s%sValid%s — %s\n", p.green, p.bold, p.reset, folioPath)
	}
}

// PrintStatusTerminal renders project status to a terminal.
func PrintStatusTerminal(w io.Writer, ps *status.ProjectStatus, causedBy map[string]string, color bool) {
	p := newPalette(color)

	fmt.Fprintf(w, "%s%s%s\n\n", p.bold, ps.Project, p.reset)

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
		fmt.Fprintf(w, "%sNo targets defined.%s\n", p.dim, p.reset)
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

				fmt.Fprintf(w, "  %-24s %s%s%s (%s)%s\n", tid, c, s, p.reset, detail, annotation)
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
				fmt.Fprintf(w, "  %s                        sources: %s%s\n", p.dim, srcList, p.reset)
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
					fmt.Fprintf(w, "    %s%s%s%s  %s → %s:%s\n", connector, c, item.Status, p.reset, item.ID, item.System, item.ExtID)
				}
			}
		}
	}

	// Tasks and pending
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Tasks: %s%d%s open\n", p.bold, ps.Tasks, p.reset)
	fmt.Fprintf(w, "Pending: %s%d%s notes\n", p.bold, ps.Pending, p.reset)
}

func printTreeNode(w io.Writer, node *status.TreeNodeStatus, p palette, indent string, isLast bool) {
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
		annotation = fmt.Sprintf(" %s<< %s%s", p.dim, node.CausedBy, p.reset)
	}

	fmt.Fprintf(w, "%s%s%s%s%s %s%s\n", indent, connector, c, node.Status, p.reset, label, annotation)

	childIndent := indent + "│   "
	if isLast {
		childIndent = indent + "    "
	}

	for i := range node.Children {
		printTreeNode(w, &node.Children[i], p, childIndent, i == len(node.Children)-1)
	}
}

func statusColor(s string, p palette) string {
	switch s {
	case "clean":
		return p.green
	case "stale":
		return p.yellow
	case "missing":
		return p.red
	case "unknown":
		return p.yellow
	default:
		return p.reset
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
	p := newPalette(color)

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
			fmt.Fprintf(w, "%s%s%s\n", p.bold, tid, p.reset)
		} else {
			sortedDeps := make([]string, len(deps))
			copy(sortedDeps, deps)
			sort.Strings(sortedDeps)
			fmt.Fprintf(w, "%s%s%s -> [%s]\n", p.bold, tid, p.reset, strings.Join(sortedDeps, ", "))
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
			fmt.Fprintf(w, "  %ssources:%s %s\n", p.dim, p.reset, strings.Join(labels, ", "))
		}

		if len(target.Outputs) > 0 {
			labels := make([]string, len(target.Outputs))
			for j, o := range target.Outputs {
				labels[j] = OutputLabel(o)
			}
			fmt.Fprintf(w, "  %soutputs:%s %s\n", p.dim, p.reset, strings.Join(labels, ", "))
		}
	}
}

// deriveBase returns the base branch for a target by looking at its first
// blocked_by dependency that has a Branch set. Falls back to "main".
func deriveBase(tid string, targets map[string]config.Target) string {
	t := targets[tid]
	for _, dep := range t.BlockedBy {
		if dt, ok := targets[dep]; ok && dt.Branch != "" {
			return dt.Branch
		}
	}
	return "main"
}

// PrintBranchDAG renders branch topology derived from targets with Branch set.
func PrintBranchDAG(w io.Writer, targets map[string]config.Target,
	adj map[string][]string, allTargets []string, color bool) {
	p := newPalette(color)

	// Filter to targets with a branch set
	var withBranch []string
	for _, tid := range allTargets {
		if targets[tid].Branch != "" {
			withBranch = append(withBranch, tid)
		}
	}

	if len(withBranch) == 0 {
		return
	}

	// Build child map: base branch → list of target IDs branching from it
	children := make(map[string][]string)
	for _, tid := range withBranch {
		base := deriveBase(tid, targets)
		children[base] = append(children[base], tid)
	}

	// Find roots: branches not themselves a target branch name, or "main"
	branchToTid := make(map[string]string)
	for _, tid := range withBranch {
		branchToTid[targets[tid].Branch] = tid
	}

	var roots []string
	for base := range children {
		if _, isBranch := branchToTid[base]; !isBranch {
			roots = append(roots, base)
		}
	}
	sort.Strings(roots)

	var printNode func(tid string, indent string, isLast bool)
	printNode = func(tid string, indent string, isLast bool) {
		t := targets[tid]
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		fmt.Fprintf(w, "%s%s%s%s%s\n", indent, connector, p.bold, tid, p.reset)

		base := deriveBase(tid, targets)
		prStr := ""
		if t.PR != "" {
			prStr = fmt.Sprintf("  PR: %s", t.PR)
		}
		fmt.Fprintf(w, "%s    %sbranch: %s (base: %s)%s%s\n", indent, p.dim, t.Branch, base, prStr, p.reset)

		childIndent := indent + "│   "
		if isLast {
			childIndent = indent + "    "
		}

		kids := children[t.Branch]
		for i, kid := range kids {
			printNode(kid, childIndent, i == len(kids)-1)
		}
	}

	for _, root := range roots {
		fmt.Fprintf(w, "%s%s%s\n", p.bold, root, p.reset)
		kids := children[root]
		for i, kid := range kids {
			printNode(kid, "", i == len(kids)-1)
		}
	}
}

// StaleEntry holds display data for a single stale target.
type StaleEntry struct {
	ID      string   `json:"id"`
	Status  string   `json:"status"`
	Outputs []string `json:"outputs"`
	Cause   string   `json:"cause"`
}

// PrintStaleTerminal renders stale targets to a terminal.
func PrintStaleTerminal(w io.Writer, entries []StaleEntry, color bool) {
	p := newPalette(color)

	for i, e := range entries {
		if i > 0 {
			fmt.Fprintln(w)
		}

		c := statusColor(e.Status, p)
		fmt.Fprintf(w, "%s%s%s  %s%s%s\n", p.bold, e.ID, p.reset, c, e.Status, p.reset)

		for _, out := range e.Outputs {
			fmt.Fprintf(w, "  %soutput:%s %s\n", p.dim, p.reset, out)
		}

		if e.Cause != "" {
			fmt.Fprintf(w, "  %scause:%s %s\n", p.dim, p.reset, e.Cause)
		}
	}
}
