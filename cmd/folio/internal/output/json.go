package output

import (
	"io"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// PrintValidateJSON renders validation results as JSON.
func PrintValidateJSON(w io.Writer, r *validate.Result) {
	if r.Errors == nil {
		r.Errors = []string{}
	}
	if r.Warnings == nil {
		r.Warnings = []string{}
	}
	dendrik.WriteResult(w, r)
}

// PrintBranchDAGJSON renders branch topology as JSON.
func PrintBranchDAGJSON(w io.Writer, bt *BranchTopology) {
	dendrik.WriteResult(w, bt)
}

// PrintStatusJSON renders project status as JSON.
func PrintStatusJSON(w io.Writer, ps *status.ProjectStatus) {
	// Per-target slices may still be nil (depends on target definition)
	for tid, ts := range ps.Targets {
		if ts.Sources == nil {
			ts.Sources = []string{}
		}
		if ts.Outputs == nil {
			ts.Outputs = []status.OutputStatus{}
		}
		ps.Targets[tid] = ts
	}
	dendrik.WriteResult(w, ps)
}
