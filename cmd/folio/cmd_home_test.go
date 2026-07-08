package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// No stores.yml ⇒ implicit registry ⇒ resolveHomeOrFail returns FOLIO_HOME
// unchanged (legacy single-home), byte-for-byte.
func TestResolveHomeLegacyNoRegistry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("FOLIO_HOME", home)
	t.Chdir(home)

	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		t.Fatalf("code = %d, want OK", code)
	}
	if dir != home {
		t.Errorf("dir = %q, want umbrella %q (legacy)", dir, home)
	}
}

// A stores.yml with a default: redirects to the default store's root.
func TestResolveHomeDefaultStore(t *testing.T) {
	umbrella := t.TempDir()
	work := filepath.Join(umbrella, "thebrokencube-folio")
	if err := os.MkdirAll(work, 0755); err != nil {
		t.Fatal(err)
	}
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n  work: { path: "+work+", kind: folio }\n")
	t.Setenv("FOLIO_HOME", umbrella)
	t.Chdir(umbrella) // not inside any store → default wins

	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		t.Fatalf("code = %d, want OK", code)
	}
	if dir != work {
		t.Errorf("dir = %q, want default store %q", dir, work)
	}
}

// cwd inside a registered store overrides the default.
func TestResolveHomeCwdOverridesDefault(t *testing.T) {
	umbrella := t.TempDir()
	work := filepath.Join(umbrella, "work")
	vault := filepath.Join(umbrella, "vault-store")
	sub := filepath.Join(vault, "active", "proj")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(work, 0755); err != nil {
		t.Fatal(err)
	}
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n"+
		"  work:  { path: "+work+", kind: folio }\n"+
		"  vault: { path: "+vault+", kind: folio }\n")
	t.Setenv("FOLIO_HOME", umbrella)
	t.Chdir(sub)

	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		t.Fatalf("code = %d, want OK", code)
	}
	if dir != vault {
		t.Errorf("dir = %q, want cwd store %q", dir, vault)
	}
}

// A default: naming an unregistered store is a hard error.
func TestResolveHomeUnknownDefaultErrors(t *testing.T) {
	umbrella := t.TempDir()
	writeStores(t, umbrella, "schema: 2\ndefault: ghost\nstores:\n  work: { path: /tmp/w, kind: folio }\n")
	t.Setenv("FOLIO_HOME", umbrella)
	t.Chdir(umbrella)

	if _, code := resolveHomeOrFail(); code == dendrik.ExitOK {
		t.Fatal("unknown default store must fail, got OK")
	}
}

// An explicit <store> positional resolves to that store's root, overriding the
// active/default resolution.
func TestResolveSyncTargetExplicitStore(t *testing.T) {
	umbrella := t.TempDir()
	work := filepath.Join(umbrella, "work")
	adr := filepath.Join(umbrella, "adr")
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n"+
		"  work: { path: "+work+", kind: folio }\n"+
		"  adr:  { path: "+adr+", kind: external }\n")
	t.Setenv("FOLIO_HOME", umbrella)
	t.Chdir(umbrella)

	dir, store, code := resolveSyncTarget("adr")
	if code != dendrik.ExitOK {
		t.Fatalf("code = %d, want OK", code)
	}
	if dir != adr || store.Name != "adr" || !store.IsExternal() {
		t.Errorf("got (dir=%q store=%q external=%v), want adr/external", dir, store.Name, store.IsExternal())
	}
}

func TestResolveSyncTargetUnknownStore(t *testing.T) {
	umbrella := t.TempDir()
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n  work: { path: /tmp/w, kind: folio }\n")
	t.Setenv("FOLIO_HOME", umbrella)
	t.Chdir(umbrella)

	if _, _, code := resolveSyncTarget("ghost"); code == dendrik.ExitOK {
		t.Fatal("unknown <store> positional must fail, got OK")
	}
}

// Pushing an external store is rejected (read-only) before any repo work.
func TestHomePushRejectsExternalStore(t *testing.T) {
	umbrella := t.TempDir()
	adr := filepath.Join(umbrella, "adr")
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n"+
		"  work: { path: "+filepath.Join(umbrella, "work")+", kind: folio }\n"+
		"  adr:  { path: "+adr+", kind: external }\n")
	t.Setenv("FOLIO_HOME", umbrella)
	t.Chdir(umbrella)

	if code := runHomePush([]string{"adr", "-m", "feat(x): y"}); code == dendrik.ExitOK {
		t.Fatal("push to external store must be rejected, got OK")
	}
}

// Push requires -m; the positional is reserved for <store>, never the message.
func TestHomePushRequiresMessage(t *testing.T) {
	umbrella := t.TempDir()
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n  work: { path: /tmp/w, kind: folio }\n")
	t.Setenv("FOLIO_HOME", umbrella)
	t.Chdir(umbrella)

	if code := runHomePush([]string{"work"}); code == dendrik.ExitOK {
		t.Fatal("push without -m must fail, got OK")
	}
}

func writeStores(t *testing.T, dir, yaml string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "stores.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
}

// The router handles help/empty-group at the workspace group level before any
// home resolution or jj requirement — so a bare '--help' never falls through to
// running create (the bug that spawned this).
func TestWorkspaceDispatchHelp(t *testing.T) {
	t.Run("no args prints usage and errors", func(t *testing.T) {
		if code := buildRoot().Execute([]string{"home", "workspace"}); code != dendrik.ExitUserError {
			t.Fatalf("code = %d, want ExitUserError", code)
		}
	})
	t.Run("--help returns ExitOK", func(t *testing.T) {
		if code := buildRoot().Execute([]string{"home", "workspace", "--help"}); code != dendrik.ExitOK {
			t.Fatalf("code = %d, want ExitOK", code)
		}
	})
	t.Run("help returns ExitOK", func(t *testing.T) {
		if code := buildRoot().Execute([]string{"home", "workspace", "help"}); code != dendrik.ExitOK {
			t.Fatalf("code = %d, want ExitOK", code)
		}
	})
}

// The router parses flags and enforces arity before Run, so bad flags and stray
// positionals are rejected (not silently ignored) without needing a jj repo —
// the leaf's home/jj resolution is never reached.
func TestWorkspaceCreateStrictArgs(t *testing.T) {
	t.Run("--help short-circuits to ExitOK", func(t *testing.T) {
		if code := buildRoot().Execute([]string{"home", "workspace", "create", "--help"}); code != dendrik.ExitOK {
			t.Fatalf("code = %d, want ExitOK", code)
		}
	})
	t.Run("unknown flag is rejected", func(t *testing.T) {
		if code := buildRoot().Execute([]string{"home", "workspace", "create", "--bogus"}); code != dendrik.ExitUserError {
			t.Fatalf("code = %d, want ExitUserError", code)
		}
	})
	t.Run("stray positional is rejected", func(t *testing.T) {
		if code := buildRoot().Execute([]string{"home", "workspace", "create", "oops"}); code != dendrik.ExitUserError {
			t.Fatalf("code = %d, want ExitUserError", code)
		}
	})
}
