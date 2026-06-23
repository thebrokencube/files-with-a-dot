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

// MarshalEnvelope is the single source of envelope JSON formatting: indented
// two spaces, with a trailing newline. Every path that emits a ResultEnvelope
// as JSON goes through here, so the encoding cannot drift between paths.
func MarshalEnvelope(env ResultEnvelope) ([]byte, error) {
	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// WriteResult marshals data as indented JSON to w.
func WriteResult(w io.Writer, data any) error {
	return WriteResultWithExit(w, data, 0)
}

// WriteResultWithExit marshals data as indented JSON to w, carrying a
// non-zero exit code in the envelope's exit_code field. Use for exit-as-signal
// commands that succeed structurally but flag a condition via exit status
// (e.g. `folio stale` emits the stale set AND exit_code: 1). Per the contract,
// exit_code is omitted when code == 0.
func WriteResultWithExit(w io.Writer, data any, code int) error {
	env := ResultEnvelope{Data: data}
	if code != 0 {
		env.ExitCode = &code
	}
	return writeEnvelope(w, env)
}

// WriteError marshals a structured error as indented JSON to w.
func WriteError(w io.Writer, msg, detail string) error {
	return writeEnvelope(w, ResultEnvelope{Error: msg, Detail: detail})
}

func writeEnvelope(w io.Writer, env ResultEnvelope) error {
	b, err := MarshalEnvelope(env)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}
