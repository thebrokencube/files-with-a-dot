package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
)

// PrintValidateJSON renders validation results as JSON.
func PrintValidateJSON(w io.Writer, r *validate.Result) {
	// Ensure non-nil slices for JSON output
	if r.Errors == nil {
		r.Errors = []string{}
	}
	if r.Warnings == nil {
		r.Warnings = []string{}
	}
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Fprintf(w, `{"valid":false,"errors":["json marshal error: %s"],"warnings":[]}`, err)
		fmt.Fprintln(w)
		return
	}
	fmt.Fprintln(w, string(data))
}

// PrintStatusJSON renders project status as JSON.
func PrintStatusJSON(w io.Writer, ps *status.ProjectStatus) {
	if ps.Sources == nil {
		ps.Sources = []status.SourceInfo{}
	}
	if ps.Targets == nil {
		ps.Targets = make(map[string]status.TargetStatus)
	}
	for tid, ts := range ps.Targets {
		if ts.Sources == nil {
			ts.Sources = []string{}
		}
		if ts.Outputs == nil {
			ts.Outputs = []status.OutputStatus{}
		}
		ps.Targets[tid] = ts
	}
	data, err := json.Marshal(ps)
	if err != nil {
		fmt.Fprintf(w, `{"error":"json marshal error: %s"}`, err)
		fmt.Fprintln(w)
		return
	}
	fmt.Fprintln(w, string(data))
}
