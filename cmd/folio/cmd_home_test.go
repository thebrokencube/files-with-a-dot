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

func writeStores(t *testing.T, dir, yaml string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "stores.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
}
