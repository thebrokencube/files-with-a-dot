package dendrik

import (
	"encoding/json"
	"fmt"
)

// Output is an inert formatter that returns bytes/strings based on output mode.
// Parallel to Palette — value type, never writes to io.Writer.
type Output struct {
	Mode string  // "json" | "plain" | "human"
	Pal  Palette // exported for color access (out.Pal.Red, out.Pal.Errf)
}

// NewOutput creates an Output from flag values.
// Color is auto-disabled in non-human modes.
func NewOutput(jsonFlag, plainFlag, noColorFlag bool) Output {
	mode := OutputMode(jsonFlag, plainFlag)
	color := mode == "human" && ColorEnabled(noColorFlag)
	return Output{Mode: mode, Pal: NewPalette(color)}
}

// IsJSON returns true if output mode is "json".
func (o Output) IsJSON() bool {
	return o.Mode == "json"
}

// Result marshals data into a ResultEnvelope JSON with trailing newline.
func (o Output) Result(data any) ([]byte, error) {
	env := ResultEnvelope{Data: data}
	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// MustResult marshals data into a ResultEnvelope JSON, panicking on failure.
// Use for known-marshalable structs (template.Must pattern).
func (o Output) MustResult(data any) []byte {
	b, err := o.Result(data)
	if err != nil {
		panic(fmt.Sprintf("dendrik.Output.MustResult: %v", err))
	}
	return b
}

// Error returns a formatted error string.
// JSON mode: ResultEnvelope JSON bytes as string. Human mode: colored "Error: msg".
func (o Output) Error(msg, detail string) string {
	if o.IsJSON() {
		env := ResultEnvelope{Error: msg, Detail: detail}
		b, err := json.MarshalIndent(env, "", "  ")
		if err != nil {
			return fmt.Sprintf(`{"error":%q}`, msg)
		}
		return string(b) + "\n"
	}
	if detail != "" {
		return o.Pal.Errf("%s\n  %s", msg, detail)
	}
	return o.Pal.Errf("%s", msg)
}

// Success returns a formatted success string.
// JSON mode: empty string (success is conveyed via Result). Human mode: colored checkmark.
func (o Output) Success(format string, a ...any) string {
	if o.IsJSON() {
		return ""
	}
	return o.Pal.Successf(format, a...)
}
