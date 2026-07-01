package sync

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestForDispatch(t *testing.T) {
	cases := []struct {
		kind     string
		wantType string
		readOnly bool
		canPush  bool
	}{
		{config.KindFolio, "sync.kbStrategy", false, true},
		{"", "sync.kbStrategy", false, true}, // legacy single-home (zero-value Store)
		{config.KindExternal, "sync.externalStrategy", true, false},
		{config.KindCode, "sync.stubStrategy", false, true},
		{config.KindDot, "sync.stubStrategy", false, true},
		{"bogus-unknown", "sync.externalStrategy", true, false}, // safe fallback
	}
	for _, c := range cases {
		s := config.Store{Kind: c.kind}
		got := For(s)
		if got.ReadOnly() != c.readOnly {
			t.Errorf("kind %q: ReadOnly()=%v, want %v", c.kind, got.ReadOnly(), c.readOnly)
		}
		if CanPush(s) != c.canPush {
			t.Errorf("kind %q: CanPush()=%v, want %v", c.kind, CanPush(s), c.canPush)
		}
	}
}

func TestExternalPushRefused(t *testing.T) {
	s := config.Store{Kind: config.KindExternal}
	_, err := For(s).Push("/tmp/whatever", s, "docs(x): y", PushOpts{})
	if !errors.Is(err, ErrReadOnly) {
		t.Fatalf("external Push err = %v, want ErrReadOnly", err)
	}
}

// dot Push (P4) uses the same delegate mechanic as code: it refuses on a shared
// branch and, on a feature branch, emits an ErrDelegate to /commit.
func TestDotPushRefusesMainEmitsDelegate(t *testing.T) {
	dir := gitInit(t) // on main
	s := config.Store{Name: "dotfiles", Kind: config.KindDot, Path: dir, DefaultBranch: "main"}

	if _, err := For(s).Push(dir, s, "chore(x): y", PushOpts{}); err == nil {
		t.Fatal("expected refusal on main for dot push")
	}
	cmd := exec.Command("git", "checkout", "-q", "-b", "feature/dots")
	cmd.Dir = dir
	if out, cerr := cmd.CombinedOutput(); cerr != nil {
		t.Fatalf("checkout: %v\n%s", cerr, out)
	}
	var del *ErrDelegate
	if _, err := For(s).Push(dir, s, "chore(x): y", PushOpts{}); !errors.As(err, &del) {
		t.Fatalf("expected ErrDelegate for dot push on feature branch, got %v", err)
	}
}
