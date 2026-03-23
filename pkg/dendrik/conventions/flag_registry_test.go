package conventions

import "testing"

func TestCollisions(t *testing.T) {
	collisions := Collisions()

	// -f is the known cross-CLI collision (force in jf, folio in folio)
	if _, ok := collisions['f']; !ok {
		t.Error("expected -f to be a cross-CLI collision")
	}

	// -j should NOT be a collision (it's global, same meaning everywhere)
	if _, ok := collisions['j']; ok {
		t.Error("-j should not be a collision (global flag)")
	}
}

func TestIsGlobalFlag(t *testing.T) {
	tests := []struct {
		flag byte
		want bool
	}{
		{'h', true},
		{'j', true},
		{'n', true},
		{'f', false},
		{'d', false},
		{'z', false},
	}
	for _, tt := range tests {
		t.Run(string(tt.flag), func(t *testing.T) {
			if got := IsGlobalFlag(tt.flag); got != tt.want {
				t.Errorf("IsGlobalFlag(%q) = %v, want %v", tt.flag, got, tt.want)
			}
		})
	}
}

func TestGlobalFlagsNotInCLIFlags(t *testing.T) {
	// Global flags should not appear in CLIFlags with a different long name
	for _, global := range GlobalFlags {
		for _, cli := range CLIFlags {
			if cli.Short == global.Short && cli.Long != global.Long {
				t.Errorf("CLI flag -%c/--%s (%s) conflicts with global flag -%c/--%s",
					cli.Short, cli.Long, cli.CLI, global.Short, global.Long)
			}
		}
	}
}
