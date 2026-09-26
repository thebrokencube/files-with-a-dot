package lease

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func loadRegistry(t *testing.T, stores string) *config.Registry {
	t.Helper()
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "stores.yml"), []byte("schema: 3\nstores:\n"+stores), 0o644); err != nil {
		t.Fatal(err)
	}
	reg, err := config.LoadRegistryFrom(home)
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func TestDecode(t *testing.T) {
	reg := loadRegistry(t, `
  plain: { path: /code/plain, kind: code }
  app:   { path: /code/app, kind: code, serve: { url: "http://localhost:3000", processes: [{ name: web, port: 3000, run: bin/web }], require_env: [APP_TOKEN] } }
  none:  { path: /code/none, kind: code, serve: { processes: [] } }
  twice: { path: /code/twice, kind: code, serve: { processes: [{ port: 3000, run: a }, { port: 3000, run: b }] } }
  norun: { path: /code/norun, kind: code, serve: { processes: [{ port: 3000 }] } }
  typo:  { path: /code/typo, kind: code, serve: { processes: [{ port: 3000, run: a }], require-env: [APP_TOKEN] } }
`)
	cases := []struct {
		store    string
		declared bool
		errPart  string
	}{
		{"plain", false, ""},
		{"app", true, ""},
		{"none", true, "at least one process"},
		{"twice", true, "declared twice"},
		{"norun", true, "no run command"},
		{"typo", true, "require-env"},
	}
	for _, tc := range cases {
		t.Run(tc.store, func(t *testing.T) {
			store, _ := reg.Lookup(tc.store)
			s, declared, err := Decode(store)
			if declared != tc.declared {
				t.Fatalf("declared = %v, want %v", declared, tc.declared)
			}
			if tc.errPart == "" {
				if err != nil {
					t.Fatalf("Decode: %v", err)
				}
				if tc.declared && (len(s.Processes) != 1 || s.RequireEnv[0] != "APP_TOKEN") {
					t.Fatalf("decoded %+v", s)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.errPart) {
				t.Fatalf("err = %v, want it to mention %q", err, tc.errPart)
			}
		})
	}
}

func TestFindConflict(t *testing.T) {
	reg := loadRegistry(t, `
  app:   { path: /code/app, kind: code, serve: { processes: [{ port: 3000, run: a }, { port: 3036, run: b }] } }
  web:   { path: /code/web, kind: code, serve: { processes: [{ port: 3036, run: c }] } }
  other: { path: /code/other, kind: code, serve: { processes: [{ port: 4000, run: d }] } }
  plain: { path: /code/plain, kind: code }
`)
	for _, tc := range []struct {
		store string
		want  *PortConflict
	}{
		{"app", &PortConflict{Port: 3036, Other: "web"}},
		{"web", &PortConflict{Port: 3036, Other: "app"}},
		{"other", nil},
	} {
		store, _ := reg.Lookup(tc.store)
		s, _, err := Decode(store)
		if err != nil {
			t.Fatal(err)
		}
		got := findConflict(reg, tc.store, s)
		if (got == nil) != (tc.want == nil) || (got != nil && *got != *tc.want) {
			t.Errorf("%s: findConflict = %+v, want %+v", tc.store, got, tc.want)
		}
	}
}
