package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/engine"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
	"gopkg.in/yaml.v3"
)

// testNode defines one node in the safe-sync test forest.
// Expectations use string constants (not engine types) because this ships
// before the engine package exists.
type testNode struct {
	Name       string // unique name, used as filename stem
	Sync       string // "push", "pull", "both"
	LocalBody  string // content below frontmatter (empty = frontmatter-only)
	NeedTicket bool   // false for TBD and ZZZTESTNOEXIST nodes
	SeedBase   bool   // pre-seed state.json baseline at setup --seed-baselines
	WantKind   string // "push", "pull", "skip", "blocked"
	WantBlock  string // "empty", "remote-unknown", "first-push", "first-pull", "overwrite", "conflict", ""
}

var testNodes = []testNode{
	{Name: "empty-push", Sync: "push", LocalBody: "", NeedTicket: true,
		WantKind: "blocked", WantBlock: "empty"},
	{Name: "first-push-safe", Sync: "push", LocalBody: "Substantive content for push test.", NeedTicket: true,
		WantKind: "push"},
	{Name: "first-push-conflict", Sync: "push", LocalBody: "Local content here.", NeedTicket: true,
		WantKind: "blocked", WantBlock: "first-push"},
	{Name: "first-pull-safe", Sync: "pull", LocalBody: "", NeedTicket: true,
		WantKind: "pull"},
	{Name: "first-pull-conflict", Sync: "pull", LocalBody: "Local content that would be overwritten.", NeedTicket: true,
		WantKind: "blocked", WantBlock: "first-pull"},
	{Name: "local-changed", Sync: "push", LocalBody: "Modified local content.", NeedTicket: true,
		SeedBase: true, WantKind: "push"},
	{Name: "overwrite-blocked", Sync: "push", LocalBody: "Original local content.", NeedTicket: true,
		SeedBase: true, WantKind: "blocked", WantBlock: "overwrite"},
	{Name: "conflict", Sync: "both", LocalBody: "Modified local side.", NeedTicket: true,
		SeedBase: true, WantKind: "blocked", WantBlock: "conflict"},
	{Name: "unchanged", Sync: "both", LocalBody: "Stable content.", NeedTicket: true,
		SeedBase: true, WantKind: "skip"},
	{Name: "both-local-only", Sync: "both", LocalBody: "Modified local only.", NeedTicket: true,
		SeedBase: true, WantKind: "push"},
	{Name: "both-remote-only", Sync: "both", LocalBody: "Original local content.", NeedTicket: true,
		SeedBase: true, WantKind: "pull"},
	{Name: "tbd-skip", Sync: "push", LocalBody: "Irrelevant.", NeedTicket: false,
		WantKind: "skip"},
	{Name: "remote-err", Sync: "push", LocalBody: "Substantive content.", NeedTicket: false,
		WantKind: "blocked", WantBlock: "remote-unknown"},
}

// testConfig is persisted as .test-config.yml alongside the test forest.
type testConfig struct {
	EpicKey string            `yaml:"epic_key"`
	Dir     string            `yaml:"dir"`
	Keys    map[string]string `yaml:"keys"` // node name -> Jira key
}

// runTest dispatches `jf test` subcommands.
//
// This is a CLI command invoked manually (e.g., `jf test setup`, `jf test run 2`).
// It is NOT executed by `go test ./...`. The Go unit tests in cmd_testharness_test.go
// only test the harness's local file-generation logic against temp directories —
// they never call Jira. Use `jf test run` with human oversight when you need to
// validate against real Jira tickets.
func runTest(args []string) int {
	if len(args) == 0 {
		printTestUsage()
		return dendrik.ExitUserError
	}

	switch args[0] {
	case "setup":
		return runTestSetup(args[1:])
	case "run":
		return runTestRun(args[1:])
	case "reset":
		return runTestReset(args[1:])
	case "teardown":
		return runTestTeardown(args[1:])
	case "--help", "-h", "help":
		printTestUsage()
		return dendrik.ExitOK
	default:
		fmt.Fprintf(os.Stderr, "Unknown test subcommand: %s\n", args[0])
		printTestUsage()
		return dendrik.ExitUserError
	}
}

func printTestUsage() {
	fmt.Fprintf(os.Stderr, `Usage: jf test <subcommand> [flags]

Developer-only test harness for safe-sync validation.
Operates on a dedicated test forest with 13 nodes covering all plan rules.

IMPORTANT: 'jf test run' hits real Jira (read-only by default, mutations with
--execute). It is NOT part of 'go test ./...' and requires human oversight.

Subcommands:
  setup                 Generate test forest (Phase 1: local files)
  setup --seed-baselines Fetch remote hashes and seed state.json (Phase 2: Jira read)
  run [track]           Validate plan output against expected actions (Jira read)
                        Results are always appended to .test-report.md
  run --execute         Full mutation round-trip (Jira write — after Track 3 only)
  reset                 Restore test forest to baseline (Jira read for re-seeding)
  teardown              Remove test forest and print Jira cleanup instructions
`)
}

// runTestSetup generates the test forest structure or seeds baselines.
func runTestSetup(args []string) int {
	fs := dendrik.NewFlagSet("test setup")
	epic := fs.StringLong("epic", "", "Epic key (e.g., BEN-123)")
	dir := fs.StringLong("dir", "", "Directory for test forest")
	keys := fs.StringLong("keys", "", "Key mapping: name=KEY,name=KEY,...")
	seedBaselines := fs.BoolLong("seed-baselines", "Phase 2: fetch remote hashes and seed state.json")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *seedBaselines {
		return runTestSeedBaselines(*dir)
	}

	if *epic == "" || *dir == "" || *keys == "" {
		fmt.Fprintln(os.Stderr, "✗ --epic, --dir, and --keys are required for setup")
		return dendrik.ExitUserError
	}

	keyMap, err := parseKeyMap(*keys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ --keys: %s\n", err)
		return dendrik.ExitUserError
	}

	return generateTestForest(*epic, *dir, keyMap)
}

// parseKeyMap parses "name=KEY,name=KEY,..." into a map.
func parseKeyMap(raw string) (map[string]string, error) {
	result := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key pair: %q (expected name=KEY)", pair)
		}
		name := strings.TrimSpace(parts[0])
		key := strings.TrimSpace(parts[1])
		if name == "" || key == "" {
			return nil, fmt.Errorf("empty name or key in pair: %q", pair)
		}
		result[name] = key
	}
	return result, nil
}

// generateTestForest creates the test forest directory with forest.yml,
// .md files, and .test-config.yml.
func generateTestForest(epic, dir string, keyMap map[string]string) int {
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "✗ mkdir %s: %s\n", dir, err)
		return dendrik.ExitUserError
	}

	// Generate forest.yml
	forestYML := "schema: 1\n\ndefaults:\n  sync: push\n  type: Story\n  project: BEN\n"
	if err := os.WriteFile(filepath.Join(dir, "forest.yml"), []byte(forestYML), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "✗ write forest.yml: %s\n", err)
		return dendrik.ExitUserError
	}

	// Generate .md files for each test node
	for _, n := range testNodes {
		key := resolveTestNodeKey(n, keyMap)
		content := buildTestNodeFile(n, key)
		path := filepath.Join(dir, n.Name+".md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "✗ write %s: %s\n", path, err)
			return dendrik.ExitUserError
		}
	}

	// Validate generated forest
	f, err := forest.FindForest(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ forest validation failed: %s\n", err)
		return dendrik.ExitUserError
	}
	if f == nil {
		fmt.Fprintln(os.Stderr, "✗ forest.yml not found after generation")
		return dendrik.ExitUserError
	}
	roots, err := forest.Discover(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ forest discovery failed: %s\n", err)
		return dendrik.ExitUserError
	}
	fmt.Printf("✓ Generated test forest: %d nodes discovered\n", len(forest.Flatten(roots)))

	// Write .test-config.yml
	cfg := testConfig{
		EpicKey: epic,
		Dir:     dir,
		Keys:    keyMap,
	}
	if code := writeTestConfig(dir, &cfg); code != dendrik.ExitOK {
		return code
	}

	fmt.Printf("✓ Config written to %s\n", filepath.Join(dir, ".test-config.yml"))
	fmt.Println()
	fmt.Println("Phase 1 complete. Next steps:")
	fmt.Println("  1. Create Jira tickets via MCP (see skill/references/testing.md)")
	fmt.Println("  2. Run: jf test setup --seed-baselines --dir", dir)
	return dendrik.ExitOK
}

// resolveTestNodeKey returns the Jira key for a test node.
// TBD node gets "TBD", remote-err node gets a nonexistent key,
// others come from the key map.
func resolveTestNodeKey(n testNode, keyMap map[string]string) string {
	switch n.Name {
	case "tbd-skip":
		return "TBD"
	case "remote-err":
		return "ZZZTESTNOEXIST-99999"
	default:
		if key, ok := keyMap[n.Name]; ok {
			return key
		}
		return "TBD"
	}
}

// buildTestNodeFile generates the markdown content for a test node.
func buildTestNodeFile(n testNode, key string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("jira: %s\n", key))
	sb.WriteString(fmt.Sprintf("sync: %s\n", n.Sync))
	sb.WriteString("---\n")
	if n.LocalBody != "" {
		sb.WriteString("\n")
		sb.WriteString(n.LocalBody)
		sb.WriteString("\n")
	}
	return sb.String()
}

// writeTestConfig writes .test-config.yml to the test forest directory.
func writeTestConfig(dir string, cfg *testConfig) int {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ marshal config: %s\n", err)
		return dendrik.ExitUserError
	}
	header := "# jf test harness config — generated by 'jf test setup'\n" +
		"# This file is local to the test forest, NOT ~/.jf.yml\n"
	path := filepath.Join(dir, ".test-config.yml")
	if err := os.WriteFile(path, append([]byte(header), data...), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "✗ write config: %s\n", err)
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}

// loadTestConfig reads .test-config.yml from the test forest directory.
func loadTestConfig(dir string) (*testConfig, error) {
	path := filepath.Join(dir, ".test-config.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w (run 'jf test setup' first)", path, err)
	}
	var cfg testConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &cfg, nil
}

// runTestSeedBaselines is Phase 2: fetch remote hashes from Jira and seed state.json.
// Requires Jira auth and network connectivity.
func runTestSeedBaselines(dir string) int {
	if dir == "" {
		fmt.Fprintln(os.Stderr, "✗ --dir is required for --seed-baselines")
		return dendrik.ExitUserError
	}

	cfg, err := loadTestConfig(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return dendrik.ExitUserError
	}

	state := &forest.State{Nodes: make(map[string]forest.NodeState)}

	for _, n := range testNodes {
		if !n.SeedBase {
			continue
		}

		key, ok := cfg.Keys[n.Name]
		if !ok {
			fmt.Fprintf(os.Stderr, "⚠ %s: no key in config, skipping baseline\n", n.Name)
			continue
		}

		// Compute current local hash from .md file
		mdPath := filepath.Join(dir, n.Name+".md")
		content, err := os.ReadFile(mdPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: read local file: %s\n", n.Name, err)
			return dendrik.ExitUserError
		}
		localContent := pipeline.StripFrontmatter(content)
		currentLocalHash := pipeline.ComputeLocalHash(localContent)

		// Fetch current remote hash from Jira
		remoteHash, err := fetchRemoteHash(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: fetch remote hash: %s\n", n.Name, err)
			return dendrik.ExitExternalErr
		}

		// Seed baseline hashes per the baseline hash strategy:
		// "changed" side gets a synthetic stale hash, "unchanged" gets the real hash.
		baseLocalH, baseRemoteH := computeBaselineHashes(n, currentLocalHash, remoteHash)

		state.Nodes[key] = forest.NodeState{
			LocalHash:  baseLocalH,
			RemoteHash: baseRemoteH,
		}
		fmt.Printf("✓ %s (%s): seeded baseline (local=%s, remote=%s)\n",
			n.Name, key, abbrevHash(baseLocalH), abbrevHash(baseRemoteH))
	}

	if err := forest.SaveState(dir, state); err != nil {
		fmt.Fprintf(os.Stderr, "✗ save state.json: %s\n", err)
		return dendrik.ExitUserError
	}

	fmt.Printf("✓ state.json written with %d baseline entries\n", len(state.Nodes))
	return dendrik.ExitOK
}

// computeBaselineHashes returns the baseline local and remote hashes for a node.
// Nodes needing "changed" on a side get a synthetic stale hash; nodes needing
// "unchanged" get the current real hash.
func computeBaselineHashes(n testNode, currentLocalHash, currentRemoteHash string) (baseLocal, baseRemote string) {
	const staleLocal = "stale-baseline-local"
	const staleRemote = "stale-baseline-remote"

	switch n.Name {
	case "local-changed":
		// local changed, remote unchanged
		return staleLocal, currentRemoteHash
	case "overwrite-blocked":
		// local unchanged, remote changed
		return currentLocalHash, staleRemote
	case "conflict":
		// both changed
		return staleLocal, staleRemote
	case "unchanged":
		// neither changed
		return currentLocalHash, currentRemoteHash
	case "both-local-only":
		// local changed, remote unchanged
		return staleLocal, currentRemoteHash
	case "both-remote-only":
		// local unchanged, remote changed
		return currentLocalHash, staleRemote
	default:
		return currentLocalHash, currentRemoteHash
	}
}

// fetchRemoteHash fetches the Jira description ADF and returns its sha256 hash.
func fetchRemoteHash(key string) (string, error) {
	p := newTestPipeline()
	viewJSON, err := p.View(key, "description", true)
	if err != nil {
		return "", err
	}

	adf, err := pipeline.ExtractDescriptionADF(viewJSON)
	if err != nil {
		return "", fmt.Errorf("extract ADF: %w", err)
	}

	if adf == nil || string(adf) == "null" {
		return forest.ComputeHash(nil), nil
	}
	return forest.ComputeHash(adf), nil
}

// abbrevHash returns the first 8 chars of a hash for display.
func abbrevHash(h string) string {
	if len(h) > 8 {
		return h[:8]
	}
	return h
}

// reportEntry captures one per-node validation result for the report.
type reportEntry struct {
	Node     string
	Expected string
	Got      string
	Pass     bool
}

func formatAction(kind, block string) string {
	if block != "" {
		return kind + "(" + block + ")"
	}
	return kind
}

// runTestRun validates plan output against expected actions.
// Hits Jira (read-only) to build the plan. Requires human oversight.
// Always appends results to .test-report.md in the test forest directory.
func runTestRun(args []string) int {
	fs := dendrik.NewFlagSet("test run")
	dir := fs.StringLong("dir", defaultTestDir(), "Test forest directory")
	resolve := fs.String('r', "resolve", "", "Conflict resolution: local|remote")
	_ = fs.BoolLong("execute", "Full mutation round-trip (after Track 3)")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *resolve != "" && *resolve != "local" && *resolve != "remote" {
		fmt.Fprintln(os.Stderr, "✗ --resolve must be 'local' or 'remote'")
		return dendrik.ExitUserError
	}

	cfg, err := loadTestConfig(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return dendrik.ExitUserError
	}

	track := ""
	remaining := fs.GetArgs()
	if len(remaining) > 0 {
		track = remaining[0]
	}

	fmt.Println("jf test run — safe-sync validation")
	fmt.Println("WARNING: This command makes Jira API calls. Use with human oversight.")
	fmt.Println()

	fmt.Printf("Track filter: %s\n", displayTrack(track))
	fmt.Printf("Test forest: %s\n", *dir)
	fmt.Printf("Nodes: %d defined, %d need tickets\n", len(testNodes), countTicketNodes())
	fmt.Println()

	var entries []reportEntry
	runResolve := *resolve
	exitCode := dendrik.ExitOK

	// Track 0 validation: forest structure
	if track == "" || track == "0" {
		fmt.Println("── Track 0: Structure ──")
		code := validateTrack0(cfg)
		if code != dendrik.ExitOK {
			exitCode = code
		}
	}

	// Tracks 2-3: engine validation (requires Jira read)
	if exitCode == dendrik.ExitOK && (track == "2" || track == "3" || track == "") {
		fmt.Println()
		fmt.Printf("── Track %s: Engine Plan ──\n", displayTrack(track))
		var code int
		entries, code = validateEnginePlan(cfg, runResolve)
		if code != dendrik.ExitOK {
			exitCode = code
		}
	}

	// Track 4: grep checks
	if exitCode == dendrik.ExitOK && (track == "4" || track == "") {
		fmt.Println()
		fmt.Println("── Track 4: Grep checks ──")
		code := validateGrepChecks()
		if code != dendrik.ExitOK {
			exitCode = code
		}
	}

	// Always write report
	if err := appendReport(cfg.Dir, track, runResolve, entries, exitCode); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ report write failed: %s\n", err)
	}

	if exitCode == dendrik.ExitOK {
		fmt.Println()
		fmt.Println("✓ All checks passed")
	}
	return exitCode
}

// validateEnginePlan runs engine Read → Plan on the test forest and compares
// each action against the expected WantKind/WantBlock from testNodes.
// Returns per-node report entries alongside the exit code.
func validateEnginePlan(cfg *testConfig, resolve string) ([]reportEntry, int) {
	f, roots, err := loadForest(cfg.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ loadForest: %s\n", err)
		return nil, dendrik.ExitUserError
	}

	state, stateErr := forest.LoadState(f.Dir)
	if stateErr != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	all := forest.Flatten(roots)
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	fmt.Printf("Reading %d nodes (Jira API calls)...\n", len(all))
	readings, err := engine.Read(all, p, state, f.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ engine.Read: %s\n", err)
		return nil, dendrik.ExitExternalErr
	}

	opts := engine.PlanOpts{Direction: "both", Resolve: resolve}
	plan := engine.Plan(readings, opts)

	fmt.Printf("Plan generated: %d actions\n\n", len(plan))

	// Build expected map: filename stem -> testNode
	expected := make(map[string]testNode)
	for _, tn := range testNodes {
		expected[tn.Name] = tn
	}

	var entries []reportEntry
	passed := 0
	failed := 0
	for _, a := range plan {
		// Extract stem from file (e.g., "empty-push.md" -> "empty-push")
		stem := strings.TrimSuffix(a.Node.File, ".md")

		tn, ok := expected[stem]
		if !ok {
			fmt.Fprintf(os.Stderr, "  ? %s: no expected definition\n", a.Node.File)
			entries = append(entries, reportEntry{Node: stem, Expected: "?", Got: "unknown", Pass: false})
			failed++
			continue
		}

		wantKind := tn.WantKind
		wantBlock := tn.WantBlock

		// --resolve changes conflict expectation
		if resolve == "local" && tn.Name == "conflict" {
			wantKind = "push"
			wantBlock = ""
		} else if resolve == "remote" && tn.Name == "conflict" {
			wantKind = "pull"
			wantBlock = ""
		}

		gotKind := a.Kind.String()
		gotBlock := a.Block.String()

		kindMatch := gotKind == wantKind
		blockMatch := gotBlock == wantBlock
		pass := kindMatch && blockMatch

		entries = append(entries, reportEntry{
			Node:     stem,
			Expected: formatAction(wantKind, wantBlock),
			Got:      formatAction(gotKind, gotBlock),
			Pass:     pass,
		})

		if pass {
			fmt.Printf("  ✓ %-25s %s", stem, gotKind)
			if gotBlock != "" {
				fmt.Printf("(%s)", gotBlock)
			}
			fmt.Println()
			passed++
		} else {
			fmt.Fprintf(os.Stderr, "  ✗ %-25s got %s", stem, gotKind)
			if gotBlock != "" {
				fmt.Fprintf(os.Stderr, "(%s)", gotBlock)
			}
			fmt.Fprintf(os.Stderr, ", want %s", wantKind)
			if wantBlock != "" {
				fmt.Fprintf(os.Stderr, "(%s)", wantBlock)
			}
			fmt.Fprintln(os.Stderr)
			failed++
		}
	}

	fmt.Printf("\nResults: %d passed, %d failed (of %d total)\n", passed, failed, len(plan))

	if failed > 0 {
		return entries, dendrik.ExitUserError
	}
	return entries, dendrik.ExitOK
}

// validateGrepChecks verifies no p.Push calls remain in cmd_*.go files.
func validateGrepChecks() int {
	// Check for p.Push in cmd_*.go files
	entries, err := os.ReadDir(".")
	if err != nil {
		// Try from the jf directory
		entries, err = os.ReadDir(filepath.Join(os.Getenv("HOME"), ".dotfiles", "cmd", "jf"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ cannot read cmd directory: %s\n", err)
			return dendrik.ExitUserError
		}
	}

	violations := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "cmd_") || !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		content, err := os.ReadFile(name)
		if err != nil {
			continue
		}
		if strings.Contains(string(content), "p.Push(") {
			fmt.Fprintf(os.Stderr, "  ✗ %s: contains p.Push() call\n", name)
			violations++
		}
	}

	if violations > 0 {
		fmt.Fprintf(os.Stderr, "✗ %d cmd_*.go file(s) still have p.Push() calls\n", violations)
		return dendrik.ExitUserError
	}

	fmt.Println("  ✓ No p.Push() calls in cmd_*.go files")
	return dendrik.ExitOK
}

// validateTrack0 checks that the test forest is parseable and has the expected nodes.
func validateTrack0(cfg *testConfig) int {
	f, err := forest.FindForest(cfg.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ FindForest: %s\n", err)
		return dendrik.ExitUserError
	}
	if f == nil {
		fmt.Fprintln(os.Stderr, "✗ forest.yml not found")
		return dendrik.ExitUserError
	}

	roots, err := forest.Discover(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Discover: %s\n", err)
		return dendrik.ExitUserError
	}

	all := forest.Flatten(roots)
	fmt.Printf("✓ Forest parsed: %d nodes\n", len(all))

	if len(all) != len(testNodes) {
		fmt.Fprintf(os.Stderr, "✗ Expected %d nodes, found %d\n", len(testNodes), len(all))
		return dendrik.ExitUserError
	}

	fmt.Println("✓ Node count matches")
	return dendrik.ExitOK
}

func displayTrack(track string) string {
	if track == "" {
		return "(all)"
	}
	return track
}

func countTicketNodes() int {
	count := 0
	for _, n := range testNodes {
		if n.NeedTicket {
			count++
		}
	}
	return count
}

func defaultTestDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".jf", "test", "safe-sync")
}

// appendReport appends a timestamped validation result to .test-report.md.
func appendReport(dir, track, resolve string, entries []reportEntry, exitCode int) error {
	path := filepath.Join(dir, ".test-report.md")

	var sb strings.Builder

	// Header
	ts := time.Now().Format("2006-01-02T15:04:05")
	outcome := "PASS"
	if exitCode != dendrik.ExitOK {
		outcome = "FAIL"
	}
	sb.WriteString(fmt.Sprintf("## %s — Track %s — %s\n\n", ts, displayTrack(track), outcome))

	if resolve != "" {
		sb.WriteString(fmt.Sprintf("Resolve: `%s`\n\n", resolve))
	}

	// Per-node table (only if engine plan ran)
	if len(entries) > 0 {
		passed := 0
		failed := 0
		for _, e := range entries {
			if e.Pass {
				passed++
			} else {
				failed++
			}
		}
		sb.WriteString(fmt.Sprintf("%d passed, %d failed (%d total)\n\n", passed, failed, len(entries)))

		sb.WriteString("| Node | Expected | Got | |\n")
		sb.WriteString("|------|----------|-----|---|\n")
		for _, e := range entries {
			mark := "✓"
			if !e.Pass {
				mark = "✗"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", e.Node, e.Expected, e.Got, mark))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(sb.String())
	if err == nil {
		fmt.Printf("Report appended to %s\n", path)
	}
	return err
}

// runTestReset restores the test forest to its baseline state.
// Regenerates .md files and re-seeds baselines (requires Jira auth).
func runTestReset(args []string) int {
	fs := dendrik.NewFlagSet("test reset")
	dir := fs.StringLong("dir", defaultTestDir(), "Test forest directory")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	cfg, err := loadTestConfig(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return dendrik.ExitUserError
	}

	// Regenerate .md files from node definitions
	for _, n := range testNodes {
		key := resolveTestNodeKey(n, cfg.Keys)
		content := buildTestNodeFile(n, key)
		path := filepath.Join(*dir, n.Name+".md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "✗ write %s: %s\n", path, err)
			return dendrik.ExitUserError
		}
	}
	fmt.Printf("✓ Regenerated %d .md files\n", len(testNodes))

	// Re-seed baselines (requires Jira auth)
	fmt.Println("Re-seeding baselines (Jira read-only)...")
	return runTestSeedBaselines(*dir)
}

// runTestTeardown removes the test forest and prints Jira cleanup instructions.
func runTestTeardown(args []string) int {
	fs := dendrik.NewFlagSet("test teardown")
	dir := fs.StringLong("dir", defaultTestDir(), "Test forest directory")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	cfg, err := loadTestConfig(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Could not load config (continuing with teardown): %s\n", err)
		cfg = nil
	}

	if err := os.RemoveAll(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "✗ remove %s: %s\n", *dir, err)
		return dendrik.ExitUserError
	}
	fmt.Printf("✓ Removed %s\n", *dir)

	if cfg != nil && len(cfg.Keys) > 0 {
		var keys []string
		for _, key := range cfg.Keys {
			keys = append(keys, key)
		}
		fmt.Println()
		fmt.Println("Jira test tickets still exist. Run this to clean up:")
		fmt.Printf("  jf park %s\n", strings.Join(keys, " "))
	}

	return dendrik.ExitOK
}

// newTestPipeline creates a pipeline for test harness Jira operations.
// Separated for testability — Go unit tests can avoid this path entirely.
func newTestPipeline() testPipelineViewer {
	return &livePipeline{}
}

// testPipelineViewer abstracts the Jira View call for testing.
type testPipelineViewer interface {
	View(key, field string, raw bool) ([]byte, error)
}

// livePipeline calls the real Jira API via the pipeline package.
type livePipeline struct{}

func (lp *livePipeline) View(key, field string, raw bool) ([]byte, error) {
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}
	return p.View(key, field, raw)
}
