package sync

import (
	"errors"
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

func TestStubKindsNotImplemented(t *testing.T) {
	for _, k := range []string{config.KindCode, config.KindDot} {
		s := config.Store{Kind: k}
		if _, err := For(s).Push("/tmp/x", s, "feat(x): y", PushOpts{}); !errors.Is(err, ErrNotImplemented) {
			t.Errorf("kind %q Push err = %v, want ErrNotImplemented", k, err)
		}
		if err := For(s).Pull("/tmp/x", s); !errors.Is(err, ErrNotImplemented) {
			t.Errorf("kind %q Pull err = %v, want ErrNotImplemented", k, err)
		}
	}
}
