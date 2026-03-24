package dendrik

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNewFlagSet(t *testing.T) {
	fs := NewFlagSet("test")
	if fs == nil {
		t.Fatal("NewFlagSet returned nil")
	}
}

func TestParse(t *testing.T) {
	t.Run("basic flags", func(t *testing.T) {
		fs := NewFlagSet("test")
		name := fs.StringLong("name", "default", "a name")
		if err := Parse(fs, []string{"--name", "hello"}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if *name != "hello" {
			t.Fatalf("got %q, want %q", *name, "hello")
		}
	})

	t.Run("short flags", func(t *testing.T) {
		fs := NewFlagSet("test")
		name := fs.String('n', "name", "default", "a name")
		if err := Parse(fs, []string{"-n", "hello"}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if *name != "hello" {
			t.Fatalf("got %q, want %q", *name, "hello")
		}
	})

	t.Run("post-positional flags", func(t *testing.T) {
		fs := NewFlagSet("test")
		token := fs.StringLong("token", "", "a token")
		if err := Parse(fs, []string{"KEY", "FILE", "--token", "abc123"}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if *token != "abc123" {
			t.Fatalf("token: got %q, want %q", *token, "abc123")
		}
		args := fs.GetArgs()
		if len(args) != 2 || args[0] != "KEY" || args[1] != "FILE" {
			t.Fatalf("positional args: got %v, want [KEY FILE]", args)
		}
	})

	t.Run("unknown flag returns error", func(t *testing.T) {
		fs := NewFlagSet("test")
		if err := Parse(fs, []string{"--bogus"}); err == nil {
			t.Fatal("expected error for unknown flag")
		}
	})
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name string
		code int
		want int
	}{
		{"OK", ExitOK, 0},
		{"UserError", ExitUserError, 1},
		{"ExternalErr", ExitExternalErr, 2},
		{"Conflict", ExitConflict, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Fatalf("got %d, want %d", tt.code, tt.want)
			}
		})
	}
}

func TestWriteResult(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}
	if err := WriteResult(&buf, data); err != nil {
		t.Fatalf("WriteResult error: %v", err)
	}

	var env ResultEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	m, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("Data is not a map: %T", env.Data)
	}
	if m["key"] != "value" {
		t.Fatalf("got %v, want %v", m["key"], "value")
	}
	if env.Error != "" {
		t.Fatalf("unexpected error field: %q", env.Error)
	}
}

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteError(&buf, "something broke", "details here"); err != nil {
		t.Fatalf("WriteError error: %v", err)
	}

	var env ResultEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if env.Error != "something broke" {
		t.Fatalf("got %q, want %q", env.Error, "something broke")
	}
	if env.Detail != "details here" {
		t.Fatalf("got %q, want %q", env.Detail, "details here")
	}
}

func TestOutputMode(t *testing.T) {
	tests := []struct {
		name  string
		json  bool
		plain bool
		want  string
	}{
		{"json flag wins", true, false, "json"},
		{"plain flag wins", false, true, "plain"},
		{"json over plain", true, true, "json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OutputMode(tt.json, tt.plain)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorEnabled(t *testing.T) {
	t.Run("noColorFlag wins", func(t *testing.T) {
		if ColorEnabled(true) {
			t.Fatal("expected false when noColorFlag is true")
		}
	})

	t.Run("NO_COLOR env wins", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		if ColorEnabled(false) {
			t.Fatal("expected false when NO_COLOR is set")
		}
	})
}

func TestPalette(t *testing.T) {
	t.Run("color enabled", func(t *testing.T) {
		p := NewPalette(true)
		if p.Red == "" {
			t.Fatal("expected non-empty Red")
		}
		if p.Reset == "" {
			t.Fatal("expected non-empty Reset")
		}
	})

	t.Run("color disabled", func(t *testing.T) {
		p := NewPalette(false)
		if p.Red != "" {
			t.Fatal("expected empty Red")
		}
		if p.Reset != "" {
			t.Fatal("expected empty Reset")
		}
	})

	t.Run("Errf", func(t *testing.T) {
		p := NewPalette(false)
		got := p.Errf("test %d", 42)
		if got != "Error: test 42" {
			t.Fatalf("got %q, want %q", got, "Error: test 42")
		}
	})

	t.Run("Successf", func(t *testing.T) {
		p := NewPalette(false)
		got := p.Successf("done %s", "now")
		if got != "✓ done now" {
			t.Fatalf("got %q, want %q", got, "✓ done now")
		}
	})
}
