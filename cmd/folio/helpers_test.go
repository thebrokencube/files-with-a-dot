package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/list"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func canonicalMainTest(t *testing.T, path string) string {
	t.Helper()
	got, err := config.CanonicalPath(path)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// setupTestHome creates a temporary FOLIO_HOME with the given active folio.yml files.
func setupTestHome(t *testing.T, paths ...string) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "active"), 0755)
	os.MkdirAll(filepath.Join(dir, "archive"), 0755)
	for _, p := range paths {
		full := filepath.Join(dir, "active", p, "folio.yml")
		os.MkdirAll(filepath.Dir(full), 0755)
		os.WriteFile(full, []byte("schema: 1\nproject: test\nsources: []\ntargets: {}\nobservations: []\n"), 0644)
	}
	return dir
}

func TestSplitStoreTarget(t *testing.T) {
	cases := []struct {
		in          string
		store, proj string
		ok          bool
	}{
		{"work:my-project", "work", "my-project", true},
		{"vault:reference/x", "vault", "reference/x", true},
		{"ben/foo", "", "", false},       // no colon — home shortname
		{"reference/a:b", "", "", false}, // colon after slash — not a store ref
		{"work:", "", "", false},         // empty project
		{":foo", "", "", false},          // empty store
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			s, p, ok := splitStoreTarget(c.in)
			if ok != c.ok || s != c.store || p != c.proj {
				t.Errorf("splitStoreTarget(%q) = (%q,%q,%v), want (%q,%q,%v)", c.in, s, p, ok, c.store, c.proj, c.ok)
			}
		})
	}
}

func TestResolveInStore(t *testing.T) {
	root := t.TempDir()
	mk := func(section, p string) {
		full := filepath.Join(root, section, p, "folio.yml")
		os.MkdirAll(filepath.Dir(full), 0755)
		os.WriteFile(full, []byte("schema: 1\nproject: t\nsources: []\ntargets: {}\nobservations: []\n"), 0644)
	}
	mk("active", "my-project")
	mk("archive", "2026-01-01-old-thing")

	got, err := resolveInStore(root, "work", "my-project")
	if err != nil {
		t.Fatalf("active match: %v", err)
	}
	if want := canonicalMainTest(t, filepath.Join(root, "active", "my-project", "folio.yml")); got != want {
		t.Errorf("active = %q, want %q", got, want)
	}

	if _, err := resolveInStore(root, "work", "2026-01-01-old-thing"); err == nil {
		t.Error("archived shortname unexpectedly resolved")
	}

	if _, err := resolveInStore(root, "work", "nope"); err == nil {
		t.Error("expected error for unknown project")
	}
}

// Same shortname in active/ and archive/ resolves to the active project.
func TestResolveInStoreActiveWinsCollision(t *testing.T) {
	root := t.TempDir()
	mk := func(section, p string) {
		full := filepath.Join(root, section, p, "folio.yml")
		os.MkdirAll(filepath.Dir(full), 0755)
		os.WriteFile(full, []byte("schema: 1\nproject: t\nsources: []\ntargets: {}\nobservations: []\n"), 0644)
	}
	mk("active", "dup")
	mk("archive", "dup")

	got, err := resolveInStore(root, "work", "dup")
	if err != nil {
		t.Fatalf("collision: %v", err)
	}
	if want := canonicalMainTest(t, filepath.Join(root, "active", "dup", "folio.yml")); got != want {
		t.Errorf("collision resolved to %q, want active %q", got, want)
	}
}

func TestIsFilePath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"./folio.yml", true},
		{"/abs/path.yml", true},
		{"~/foo.yaml", true},
		{"../relative.yml", true},
		{"ben/state-retirement-mandates", false},
		{"state-retirement-mandates", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isFilePath(tt.input)
			if got != tt.want {
				t.Errorf("isFilePath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestActiveShortnames(t *testing.T) {
	entries := []list.Entry{
		{Section: "active", Path: "ben/state-retirement-mandates"},
		{Section: "active", Path: "ben/onboarding-redesign"},
		{Section: "archive", Path: "2024-01-01_old-state-thing"},
	}

	t.Run("returns active only", func(t *testing.T) {
		got := activeShortnames(entries)
		if len(got) != 2 {
			t.Fatalf("expected 2 active paths, got %d: %v", len(got), got)
		}
	})

	t.Run("sorted alphabetically", func(t *testing.T) {
		got := activeShortnames(entries)
		if got[0] != "ben/onboarding-redesign" || got[1] != "ben/state-retirement-mandates" {
			t.Errorf("expected sorted order, got %v", got)
		}
	})

	t.Run("empty when no active", func(t *testing.T) {
		got := activeShortnames([]list.Entry{{Section: "archive", Path: "old"}})
		if len(got) != 0 {
			t.Fatalf("expected 0, got %d", len(got))
		}
	})
}

func TestResolveFolioPath(t *testing.T) {
	t.Run("file path passthrough", func(t *testing.T) {
		got, err := resolveFolioPath("./folio.yml")
		if err != nil {
			t.Fatal(err)
		}
		if got != canonicalMainTest(t, "folio.yml") {
			t.Errorf("got %q, want canonical folio.yml", got)
		}
	})

	t.Run("shortname exact match", func(t *testing.T) {
		homeDir := setupTestHome(t, "ben/state-retirement-mandates", "ben/onboarding")
		t.Setenv("FOLIO_HOME", homeDir)

		got, err := resolveFolioPath("ben/state-retirement-mandates")
		if err != nil {
			t.Fatal(err)
		}
		want := canonicalMainTest(t, filepath.Join(homeDir, "active", "ben/state-retirement-mandates", "folio.yml"))
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("shortname final component match", func(t *testing.T) {
		homeDir := setupTestHome(t, "ben/state-retirement-mandates", "ben/onboarding")
		t.Setenv("FOLIO_HOME", homeDir)

		got, err := resolveFolioPath("state-retirement-mandates")
		if err != nil {
			t.Fatal(err)
		}
		want := canonicalMainTest(t, filepath.Join(homeDir, "active", "ben/state-retirement-mandates", "folio.yml"))
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("miss lists active projects", func(t *testing.T) {
		homeDir := setupTestHome(t, "ben/state-retirement-mandates", "ben/state-tax-engine")
		t.Setenv("FOLIO_HOME", homeDir)

		_, err := resolveFolioPath("nonexistent")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "active projects") {
			t.Errorf("expected 'active projects' in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "ben/state-retirement-mandates") {
			t.Errorf("expected active projects listed in error, got: %s", err)
		}
	})

	t.Run("miss with no active projects", func(t *testing.T) {
		homeDir := setupTestHome(t) // no active projects
		t.Setenv("FOLIO_HOME", homeDir)

		_, err := resolveFolioPath("zzzzz")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "no active projects") {
			t.Errorf("expected 'no active projects' in error, got: %s", err)
		}
	})

	t.Run("FOLIO_HOME missing", func(t *testing.T) {
		t.Setenv("FOLIO_HOME", "/nonexistent/path/that/does/not/exist")

		_, err := resolveFolioPath("anything")
		if err == nil {
			t.Fatal("expected error when FOLIO_HOME is invalid")
		}
	})
}

func TestResolveContextLegacyFOLIOHomeOnly(t *testing.T) {
	homeDir := setupTestHome(t, "team/project")
	outside := t.TempDir()
	t.Setenv("FOLIO_HOME", homeDir)
	t.Setenv("FOLIO_UMBRELLA", "")
	t.Chdir(outside)

	project := filepath.Join(homeDir, "active", "team", "project", "folio.yml")
	ctx, err := resolveContext(project, contextReadOnly)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Store.Name != "" {
		t.Fatalf("legacy context selected store %q", ctx.Store.Name)
	}
	if ctx.WorkRoot != canonicalMainTest(t, homeDir) {
		t.Fatalf("work root = %q, want isolated home %q", ctx.WorkRoot, homeDir)
	}
	if ctx.FolioPath != canonicalMainTest(t, project) {
		t.Fatalf("folio path = %q, want %q", ctx.FolioPath, project)
	}
}
func TestResolveSyncTargetExplicitContentRootWins(t *testing.T) {
	umbrella := t.TempDir()
	work := filepath.Join(umbrella, "work")
	other := filepath.Join(umbrella, "other")
	if err := os.MkdirAll(work, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(other, 0755); err != nil {
		t.Fatal(err)
	}
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n"+
		"  work:  { path: "+work+", kind: folio }\n"+
		"  other: { path: "+other+", kind: folio }\n")
	t.Setenv("FOLIO_UMBRELLA", umbrella)
	t.Setenv("FOLIO_HOME", other)
	t.Chdir(umbrella)

	got, store, code := resolveSyncTarget("")
	if code != 0 {
		t.Fatalf("resolveSyncTarget code = %d", code)
	}
	if got != canonicalMainTest(t, other) || store.Name != "other" {
		t.Fatalf("sync target = (%q, %q), want explicit other root", got, store.Name)
	}
}

func TestResolveContextStoreQualifiedReferenceWins(t *testing.T) {
	umbrella := t.TempDir()
	work := filepath.Join(umbrella, "work")
	other := filepath.Join(umbrella, "other")
	for _, root := range []string{work, other} {
		full := filepath.Join(root, "active", "project", "folio.yml")
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("schema: 1\nproject: test\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n"+
		"  work:  { path: "+work+", kind: folio }\n"+
		"  other: { path: "+other+", kind: folio }\n")
	t.Setenv("FOLIO_UMBRELLA", umbrella)
	t.Setenv("FOLIO_HOME", work)
	t.Chdir(filepath.Join(work, "active", "project"))

	ctx, err := resolveContext("other:project", contextReadOnly)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Store.Name != "other" || ctx.WorkRoot != canonicalMainTest(t, other) {
		t.Fatalf("context = %+v, want other store root", ctx)
	}
	if _, err := resolveContext("other:project", contextMutation); err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("mutation should reject mismatched workspace, got %v", err)
	}
}

func TestResolveContextExplicitProjectNeverUsesUnrelatedWorkRoot(t *testing.T) {
	umbrella := t.TempDir()
	work := filepath.Join(umbrella, "work")
	other := filepath.Join(umbrella, "other")
	outside := filepath.Join(t.TempDir(), "standalone")
	registered := filepath.Join(other, "active", "project", "folio.yml")
	unregistered := filepath.Join(outside, "folio.yml")
	for _, file := range []string{registered, unregistered} {
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("schema: 1\nproject: test\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n"+
		"  work:  { path: "+work+", kind: folio }\n"+
		"  other: { path: "+other+", kind: folio }\n")
	t.Setenv("FOLIO_UMBRELLA", umbrella)
	t.Setenv("FOLIO_HOME", work)
	t.Chdir(umbrella)

	registeredCtx, err := resolveContext(registered, contextReadOnly)
	if err != nil {
		t.Fatal(err)
	}
	if registeredCtx.Store.Name != "other" || registeredCtx.WorkRoot != canonicalMainTest(t, other) {
		t.Fatalf("registered context = %+v, want owning store", registeredCtx)
	}
	unregisteredCtx, err := resolveContext(unregistered, contextReadOnly)
	if err != nil {
		t.Fatal(err)
	}
	if unregisteredCtx.Store.Name != "" || unregisteredCtx.WorkRoot != canonicalMainTest(t, outside) {
		t.Fatalf("unregistered context = %+v, want standalone root", unregisteredCtx)
	}
}

func TestWorkspaceCreateFromCodeStoreCWDUsesDefaultFolioStore(t *testing.T) {
	umbrella := t.TempDir()
	code := filepath.Join(umbrella, "code")
	folio := filepath.Join(umbrella, "folio")
	if err := os.MkdirAll(filepath.Join(code, ".jj"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(folio, ".jj"), 0755); err != nil {
		t.Fatal(err)
	}
	writeStores(t, umbrella, "schema: 2\ndefault: folio\nstores:\n"+
		"  code:  { path: "+code+", kind: code }\n"+
		"  folio: { path: "+folio+", kind: folio }\n")
	t.Setenv("FOLIO_UMBRELLA", umbrella)
	t.Setenv("FOLIO_HOME", "")
	t.Chdir(code)

	got, statusCode := resolveWorkspaceHome(dendrik.NewPalette(false))
	if statusCode != 0 {
		t.Fatalf("resolveWorkspaceHome code = %d", statusCode)
	}
	if got != canonicalMainTest(t, folio) {
		t.Fatalf("workspace root = %q, want default Folio store %q", got, folio)
	}
}

func TestResolveContextSecondaryWorkspaceDiscoversUmbrella(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not on PATH")
	}
	umbrella := t.TempDir()
	storeRoot := filepath.Join(umbrella, "folio")
	project := filepath.Join(storeRoot, "active", "project", "folio.yml")
	if err := os.MkdirAll(filepath.Dir(project), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(project, []byte("schema: 1\nproject: test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(storeRoot, 0755); err != nil {
		t.Fatal(err)
	}
	jjInit := exec.Command("jj", "git", "init", "--colocate")
	jjInit.Dir = storeRoot
	if output, err := jjInit.CombinedOutput(); err != nil {
		t.Fatalf("jj git init: %v\n%s", err, output)
	}
	writeStores(t, umbrella, "schema: 2\ndefault: folio\nstores:\n  folio: { path: "+storeRoot+", kind: folio }\n")

	workspaceParent := t.TempDir()
	workspace := filepath.Join(workspaceParent, "secondary")
	jjAdd := exec.Command("jj", "workspace", "add", workspace, "-r", "@")
	jjAdd.Dir = storeRoot
	if output, err := jjAdd.CombinedOutput(); err != nil {
		t.Fatalf("jj workspace add: %v\n%s", err, output)
	}
	t.Chdir(workspace)
	t.Setenv("FOLIO_HOME", "")
	t.Setenv("FOLIO_UMBRELLA", "")

	ctx, err := resolveContext("project", contextReadOnly)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Umbrella != canonicalMainTest(t, umbrella) {
		t.Fatalf("umbrella = %q, want %q", ctx.Umbrella, umbrella)
	}
	if ctx.Store.Name != "folio" || ctx.WorkRoot != canonicalMainTest(t, workspace) {
		t.Fatalf("context = %+v, want secondary workspace owned by folio store", ctx)
	}
}

func TestWorkspaceCreateFromCodeStoreCWDRefusesMissingDefault(t *testing.T) {
	umbrella := t.TempDir()
	code := filepath.Join(umbrella, "code")
	if err := os.MkdirAll(filepath.Join(code, ".jj"), 0755); err != nil {
		t.Fatal(err)
	}
	writeStores(t, umbrella, "schema: 2\ndefault: code\nstores:\n  code: { path: "+code+", kind: code }\n")
	t.Setenv("FOLIO_UMBRELLA", umbrella)
	t.Setenv("FOLIO_HOME", "")
	t.Chdir(code)

	if _, statusCode := resolveWorkspaceHome(dendrik.NewPalette(false)); statusCode == 0 {
		t.Fatal("workspace creation from code store must refuse without a Folio default")
	}
}
func TestUmbrellaValuedFOLIOHomeWarnsDuringCompatibilityRelease(t *testing.T) {
	umbrella := t.TempDir()
	work := filepath.Join(umbrella, "work")
	if err := os.MkdirAll(filepath.Join(work, "active", "project"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, "active", "project", "folio.yml"), []byte("schema: 1\nproject: test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n  work: { path: "+work+", kind: folio }\n")
	t.Setenv("FOLIO_UMBRELLA", umbrella)
	t.Setenv("FOLIO_HOME", umbrella)
	t.Chdir(umbrella)

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	ctx, resolveErr := resolveContext("project", contextReadOnly)
	_ = w.Close()
	os.Stderr = oldStderr
	if resolveErr != nil {
		t.Fatal(resolveErr)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "registry umbrella") {
		t.Fatalf("warning = %q, want compatibility warning", string(data))
	}
	if ctx.Store.Name != "work" || ctx.WorkRoot != canonicalMainTest(t, work) {
		t.Fatalf("context = %+v, want default Folio store", ctx)
	}
}
