package dendrik

import (
	"encoding/json"
	"io"
)

// ResultEnvelope is the standard JSON output wrapper for dendrik commands.
type ResultEnvelope struct {
	Data     any    `json:"data"`
	Error    string `json:"error,omitempty"`
	Detail   string `json:"detail,omitempty"`
	ExitCode *int   `json:"exit_code,omitempty"`
}

// WriteResult marshals data as indented JSON to w.
func WriteResult(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(ResultEnvelope{Data: data})
}

// WriteError marshals a structured error to w.
func WriteError(w io.Writer, msg, detail string) error {
	return json.NewEncoder(w).Encode(ResultEnvelope{Error: msg, Detail: detail})
}
