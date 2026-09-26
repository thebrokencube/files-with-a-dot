package lease

import (
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

const leasePGID = 500

func baseFacts() Facts {
	return Facts{
		Store:    "app",
		Path:     "/code/app",
		Declared: true,
		Serve: Serve{
			URL: "http://localhost:3000",
			Processes: []Process{
				{Name: "web", Port: 3000, Run: "bin/web"},
				{Name: "assets", Port: 3036, Run: "bin/assets"},
			},
			RequireEnv:  []string{"APP_TOKEN"},
			Alternative: "use the running server",
		},
		Owners:     map[int][]Owner{},
		InCheckout: true,
		Digest:     "sha256:same",
	}
}

func liveLease(f *Facts) {
	f.Lease = &Record{Store: "app", Checkout: "/code/app", PGID: leasePGID, EnvDigest: "sha256:same", Started: "2026-01-02T03:04:05Z", Log: "/leases/app.log"}
	f.Alive = true
}

func inside(port int) func(*Facts) {
	return func(f *Facts) { f.Owners[port] = []Owner{{PID: leasePGID + port, PGID: leasePGID}} }
}

func TestDecide(t *testing.T) {
	cases := []struct {
		name     string
		setup    []func(*Facts)
		outcome  string
		code     int
		state    string
		contains []string
	}{
		{
			name:    "no lease, ports free, main checkout, env present",
			outcome: actionStart, code: dendrik.ExitOK, state: StateFree,
		},
		{
			name: "dead lease pgid is stale and ignored",
			setup: []func(*Facts){liveLease, func(f *Facts) {
				f.Alive = false
				f.Lease.EnvDigest = "sha256:other"
			}},
			outcome: actionStart, code: dendrik.ExitOK, state: StateFree,
		},
		{
			name:    "work-area cwd with nothing running is refused",
			setup:   []func(*Facts){func(f *Facts) { f.InCheckout = false }},
			outcome: outcomeRefused, code: dendrik.ExitUserError, state: StateFree,
			contains: []string{"/code/app"},
		},
		{
			name:    "missing require_env key is refused by name",
			setup:   []func(*Facts){func(f *Facts) { f.Missing = []string{"APP_TOKEN"} }},
			outcome: outcomeRefused, code: dendrik.ExitUserError, state: StateFree,
			contains: []string{"APP_TOKEN"},
		},
		{
			name:    "port held with no lease is an unleased owner",
			setup:   []func(*Facts){func(f *Facts) { f.Owners[3036] = []Owner{{PID: 77, PGID: 77}} }},
			outcome: outcomeHeld, code: dendrik.ExitConflict, state: StateHeld,
			contains: []string{"pid 77", "3036"},
		},
		{
			name: "port held outside a live lease pgid is never treated as the lease",
			setup: []func(*Facts){liveLease, inside(3000), func(f *Facts) {
				f.Owners[3036] = []Owner{{PID: 88, PGID: 88}}
			}},
			outcome: outcomeHeld, code: dendrik.ExitConflict, state: StateHeld,
			contains: []string{"pid 88"},
		},
		{
			name: "one listener outside the pgid among inside ones is held",
			setup: []func(*Facts){liveLease, inside(3036), func(f *Facts) {
				f.Owners[3000] = []Owner{{PID: 501, PGID: leasePGID}, {PID: 99, PGID: 99}}
			}},
			outcome: outcomeHeld, code: dendrik.ExitConflict, state: StateHeld,
			contains: []string{"pid 99"},
		},
		{
			name: "listener whose pgid cannot be read is held",
			setup: []func(*Facts){liveLease, inside(3036), func(f *Facts) {
				f.Owners[3000] = []Owner{{PID: 66, PGID: -1}}
			}},
			outcome: outcomeHeld, code: dendrik.ExitConflict, state: StateHeld,
			contains: []string{"pid 66"},
		},
		{
			name:    "booting: live pgid, some ports free",
			setup:   []func(*Facts){liveLease, inside(3000)},
			outcome: outcomeHeld, code: dendrik.ExitConflict, state: StateStarting,
			contains: []string{"starting", "do not relaunch"},
		},
		{
			name:    "booting: live pgid, no port bound yet",
			setup:   []func(*Facts){liveLease},
			outcome: outcomeHeld, code: dendrik.ExitConflict, state: StateStarting,
			contains: []string{"starting"},
		},
		{
			name:    "booting from a work area is still starting",
			setup:   []func(*Facts){liveLease, func(f *Facts) { f.InCheckout = false }},
			outcome: outcomeHeld, code: dendrik.ExitConflict, state: StateStarting,
			contains: []string{"starting"},
		},
		{
			name:    "running, main checkout, same env is a no-op",
			setup:   []func(*Facts){liveLease, inside(3000), inside(3036)},
			outcome: outcomeRunning, code: dendrik.ExitOK, state: StateRunning,
			contains: []string{"http://localhost:3000", "pgid 500", "2026-01-02T03:04:05Z"},
		},
		{
			name:    "running, main checkout, env digest differs",
			setup:   []func(*Facts){liveLease, inside(3000), inside(3036), func(f *Facts) { f.Digest = "sha256:new" }},
			outcome: outcomeHeld, code: dendrik.ExitConflict, state: StateRunning,
			contains: []string{"env differs", "folio lease release app"},
		},
		{
			name: "running, main checkout, missing key is refused without suggesting release",
			setup: []func(*Facts){liveLease, inside(3000), inside(3036), func(f *Facts) {
				f.Missing = []string{"APP_TOKEN"}
				f.Digest = "sha256:empty"
			}},
			outcome: outcomeRefused, code: dendrik.ExitUserError, state: StateRunning,
			contains: []string{"APP_TOKEN", "keep using the running server"},
		},
		{
			name:    "running, work-area cwd names holder and alternative",
			setup:   []func(*Facts){liveLease, inside(3000), inside(3036), func(f *Facts) { f.InCheckout = false; f.Digest = "sha256:new" }},
			outcome: outcomeHeld, code: dendrik.ExitConflict, state: StateRunning,
			contains: []string{"/code/app", "pgid 500", "use the running server"},
		},
		{
			name:    "undeclared store is refused",
			setup:   []func(*Facts){func(f *Facts) { *f = Facts{Store: "app"} }},
			outcome: outcomeRefused, code: dendrik.ExitUserError,
			contains: []string{"no serve block"},
		},
		{
			name: "duplicate port across serve blocks names both stores",
			setup: []func(*Facts){liveLease, inside(3000), inside(3036), func(f *Facts) {
				f.Conflict = &PortConflict{Port: 3036, Other: "example-web"}
			}},
			outcome: outcomeRefused, code: dendrik.ExitUserError,
			contains: []string{"3036", `"app"`, `"example-web"`},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := baseFacts()
			for _, fn := range tc.setup {
				fn(&f)
			}
			v := decide(f)
			if v.Outcome != tc.outcome || v.Code != tc.code {
				t.Fatalf("decide = (%s, %d), want (%s, %d): %s", v.Outcome, v.Code, tc.outcome, tc.code, v.Message)
			}
			for _, want := range tc.contains {
				if !strings.Contains(v.Message, want) {
					t.Errorf("message %q missing %q", v.Message, want)
				}
			}
			if tc.state != "" {
				if got := stateOf(f).State; got != tc.state {
					t.Errorf("state = %s, want %s", got, tc.state)
				}
			}
		})
	}
}

func TestDecideRelease(t *testing.T) {
	outside := func(port int) func(*Facts) {
		return func(f *Facts) { f.Owners[port] = append(f.Owners[port], Owner{PID: 77, PGID: 77}) }
	}
	cases := []struct {
		name     string
		setup    []func(*Facts)
		outcome  string
		code     int
		contains string
	}{
		{"undeclared", []func(*Facts){func(f *Facts) { *f = Facts{Store: "app"} }}, outcomeRefused, dendrik.ExitUserError, "no serve block"},
		{"work-area cwd", []func(*Facts){liveLease, inside(3000), inside(3036), func(f *Facts) { f.InCheckout = false }}, outcomeRefused, dendrik.ExitUserError, "/code/app"},
		{"no lease", []func(*Facts){outside(3000)}, actionNone, dendrik.ExitOK, "nothing to release"},
		{"stale lease is only removed", []func(*Facts){liveLease, inside(3000), func(f *Facts) { f.Alive = false }}, actionRemove, dendrik.ExitOK, ""},
		{"running lease is stopped", []func(*Facts){liveLease, inside(3000), inside(3036)}, actionStop, dendrik.ExitOK, ""},
		{"partly bound lease is stopped", []func(*Facts){liveLease, inside(3000)}, actionStop, dendrik.ExitOK, ""},
		{"inside owner beside an outside one is stopped", []func(*Facts){liveLease, inside(3000), outside(3036)}, actionStop, dendrik.ExitOK, ""},
		{"booting lease with nothing bound is kept", []func(*Facts){liveLease}, outcomeHeld, dendrik.ExitConflict, "-500"},
		{"live lease with ports owned only outside is kept", []func(*Facts){liveLease, outside(3000)}, outcomeHeld, dendrik.ExitConflict, "starting"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := baseFacts()
			for _, fn := range tc.setup {
				fn(&f)
			}
			v := decideRelease(f)
			if v.Outcome != tc.outcome || v.Code != tc.code {
				t.Fatalf("decideRelease = (%s, %d), want (%s, %d): %s", v.Outcome, v.Code, tc.outcome, tc.code, v.Message)
			}
			if !strings.Contains(v.Message, tc.contains) {
				t.Errorf("message %q missing %q", v.Message, tc.contains)
			}
		})
	}
}
