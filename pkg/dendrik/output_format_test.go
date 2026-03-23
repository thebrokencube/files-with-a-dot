package dendrik

import (
	"encoding/json"
	"testing"
)

func TestOutputIsJSON(t *testing.T) {
	tests := []struct {
		mode string
		want bool
	}{
		{"json", true},
		{"human", false},
		{"plain", false},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			o := Output{Mode: tt.mode}
			if got := o.IsJSON(); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputResult(t *testing.T) {
	o := Output{Mode: "json"}

	t.Run("marshals envelope", func(t *testing.T) {
		b, err := o.Result(map[string]string{"k": "v"})
		if err != nil {
			t.Fatalf("Result error: %v", err)
		}
		var env ResultEnvelope
		if err := json.Unmarshal(b, &env); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		m, ok := env.Data.(map[string]any)
		if !ok {
			t.Fatalf("Data is not a map: %T", env.Data)
		}
		if m["k"] != "v" {
			t.Fatalf("got %v, want v", m["k"])
		}
	})

	t.Run("trailing newline", func(t *testing.T) {
		b, err := o.Result("hello")
		if err != nil {
			t.Fatalf("Result error: %v", err)
		}
		if b[len(b)-1] != '\n' {
			t.Fatal("missing trailing newline")
		}
	})

	t.Run("no exit_code when nil", func(t *testing.T) {
		b, err := o.Result("data")
		if err != nil {
			t.Fatalf("Result error: %v", err)
		}
		var raw map[string]any
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if _, exists := raw["exit_code"]; exists {
			t.Fatal("exit_code should be omitted when nil")
		}
	})
}

func TestOutputMustResult(t *testing.T) {
	o := Output{Mode: "json"}

	t.Run("succeeds for marshalable data", func(t *testing.T) {
		b := o.MustResult(map[string]string{"a": "b"})
		if len(b) == 0 {
			t.Fatal("empty result")
		}
	})

	t.Run("panics for unmarshalable data", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		o.MustResult(make(chan int))
	})
}

func TestOutputError(t *testing.T) {
	t.Run("json mode", func(t *testing.T) {
		o := Output{Mode: "json"}
		got := o.Error("bad input", "field X missing")
		var env ResultEnvelope
		if err := json.Unmarshal([]byte(got), &env); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if env.Error != "bad input" {
			t.Fatalf("got %q, want %q", env.Error, "bad input")
		}
		if env.Detail != "field X missing" {
			t.Fatalf("got %q, want %q", env.Detail, "field X missing")
		}
	})

	t.Run("human mode with detail", func(t *testing.T) {
		o := Output{Mode: "human", Pal: NewPalette(false)}
		got := o.Error("bad input", "field X missing")
		want := "Error: bad input\n  field X missing"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("human mode without detail", func(t *testing.T) {
		o := Output{Mode: "human", Pal: NewPalette(false)}
		got := o.Error("bad input", "")
		want := "Error: bad input"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

func TestOutputSuccess(t *testing.T) {
	t.Run("json mode returns empty", func(t *testing.T) {
		o := Output{Mode: "json"}
		if got := o.Success("done %d", 1); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("human mode returns formatted", func(t *testing.T) {
		o := Output{Mode: "human", Pal: NewPalette(false)}
		got := o.Success("done %d", 1)
		want := "✓ done 1"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

func TestResultEnvelopeExitCode(t *testing.T) {
	t.Run("omitted when nil", func(t *testing.T) {
		env := ResultEnvelope{Data: "ok"}
		b, _ := json.Marshal(env)
		var raw map[string]any
		json.Unmarshal(b, &raw)
		if _, exists := raw["exit_code"]; exists {
			t.Fatal("exit_code should be omitted when nil")
		}
	})

	t.Run("present when set", func(t *testing.T) {
		code := 1
		env := ResultEnvelope{Data: "stale", ExitCode: &code}
		b, _ := json.Marshal(env)
		var raw map[string]any
		json.Unmarshal(b, &raw)
		got, exists := raw["exit_code"]
		if !exists {
			t.Fatal("exit_code should be present")
		}
		if got.(float64) != 1 {
			t.Fatalf("got %v, want 1", got)
		}
	})
}
