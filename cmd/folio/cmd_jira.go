package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/jira"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/touch"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runJiraCompile(args []string) int {
	fs := dendrik.NewFlagSet("jira compile")
	id := fs.StringLong("id", "", "Jira issue key (e.g., BEN-123)")
	source := fs.StringLong("source", "", "Markdown source file (- for stdin)")
	_ = fs.StringLong("output", "", "Deprecated: output file")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *id == "" || *source == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--id and --source are required"))
		fmt.Fprintf(os.Stderr, "Usage: folio jira compile --id KEY --source FILE\n")
		return 1
	}

	fmt.Fprintln(os.Stderr, "⚠ folio jira compile is deprecated — use: jf push <KEY> <FILE>")

	// Delegate to jf push (which compiles + pushes; compile-only not available)
	return runJf("push", *id, *source)
}

func runJiraPush(args []string) int {
	fs := dendrik.NewFlagSet("jira push")
	id := fs.StringLong("id", "", "Jira issue key (e.g., BEN-123)")
	source := fs.StringLong("source", "", "Markdown source file (- for stdin)")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *id == "" || *source == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--id and --source are required"))
		fmt.Fprintf(os.Stderr, "Usage: folio jira push --id KEY --source FILE\n")
		return 1
	}

	code := runJf("push", *id, *source)
	if code == 0 {
		if touched, err := autoTouch(*source); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ autoTouch: %s\n", err)
		} else if touched > 0 {
			fmt.Printf("  Auto-touched %d output(s)\n", touched)
		}
	}
	return code
}

func runJiraCreate(args []string) int {
	fs := dendrik.NewFlagSet("jira create")
	jsonFile := fs.StringLong("json", "", "Creation JSON payload file")
	source := fs.StringLong("source", "", "Markdown source file for description")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *jsonFile == "" || *source == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--json and --source are required"))
		fmt.Fprintf(os.Stderr, "Usage: folio jira create --json FILE --source FILE\n")
		return 1
	}

	jsonPayload, err := readSource(*jsonFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("read %s: %s", *jsonFile, err))
		return 1
	}

	// Create ticket via acli (folio still owns this for now)
	p := &jira.Pipeline{Run: jira.DefaultRunner}
	key, err := p.Create(jsonPayload)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("create: %s", err))
		return 1
	}

	// Delegate compile+push to jf
	code := runJf("push", key, *source)
	if code != 0 {
		fmt.Fprintln(os.Stderr, output.Errf("push description for %s failed (ticket created)", key))
		return 1
	}

	fmt.Println(output.Successf("Created %s and pushed description", key))
	if touched, err := autoTouch(*source); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ autoTouch: %s\n", err)
	} else if touched > 0 {
		fmt.Printf("  Auto-touched %d output(s)\n", touched)
	}
	return 0
}

func runJiraView(args []string) int {
	fs := dendrik.NewFlagSet("jira view")
	id := fs.StringLong("id", "", "Jira issue key (e.g., BEN-123)")
	fields := fs.StringLong("fields", "", "Comma-separated field list")
	jsonOut := fs.BoolLong("json", "JSON output")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *id == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--id is required"))
		fmt.Fprintf(os.Stderr, "Usage: folio jira view --id KEY [--fields F] [--json]\n")
		return 1
	}

	p := &jira.Pipeline{Run: jira.DefaultRunner}
	out, err := p.View(*id, *fields, *jsonOut)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Print(string(out))
	return 0
}

func runJiraSearch(args []string) int {
	fs := dendrik.NewFlagSet("jira search")
	jql := fs.StringLong("jql", "", "JQL query string")
	fields := fs.StringLong("fields", "", "Comma-separated field list")
	limit := fs.IntLong("limit", 50, "Max results")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *jql == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--jql is required"))
		fmt.Fprintf(os.Stderr, "Usage: folio jira search --jql QUERY [--fields F] [--limit N]\n")
		return 1
	}

	p := &jira.Pipeline{Run: jira.DefaultRunner}
	out, err := p.Search(*jql, *fields, *limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Print(string(out))
	return 0
}

// runJf delegates to the jf binary with stdout/stderr passthrough.
func runJf(args ...string) int {
	cmd := exec.Command("jf", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "jf: %s\n", err)
		return 2
	}
	return 0
}

// autoTouch finds the folio.yml containing sourcePath, locates the target
// whose tree node references it, and touches that target's outputs.
func autoTouch(sourcePath string) (int, error) {
	abs, err := filepath.Abs(sourcePath)
	if err != nil {
		return 0, err
	}

	// Walk up to find folio.yml
	dir := filepath.Dir(abs)
	var folioPath string
	for {
		candidate := filepath.Join(dir, "folio.yml")
		if _, err := os.Stat(candidate); err == nil {
			folioPath = candidate
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return 0, fmt.Errorf("no folio.yml found above %s", sourcePath)
		}
		dir = parent
	}

	folioDir := filepath.Dir(folioPath)
	f, err := config.Load(folioPath)
	if err != nil {
		return 0, err
	}

	// Make sourcePath relative to folioDir for matching
	rel, err := filepath.Rel(folioDir, abs)
	if err != nil {
		return 0, err
	}

	// Find the target with a tree node referencing this source
	for _, tid := range sortedTargetKeys(f.Targets) {
		target := f.Targets[tid]
		if target.Tree != nil && treeNodeMatches(&target.Tree.Root, rel) {
			return touch.Target(folioDir, &target)
		}
	}

	return 0, fmt.Errorf("no tree target references %s", rel)
}

// treeNodeMatches returns true if any node in the tree has a File matching relPath.
func treeNodeMatches(node *config.TreeNode, relPath string) bool {
	if node.File != "" && node.File == relPath {
		return true
	}
	for i := range node.Children {
		if treeNodeMatches(&node.Children[i], relPath) {
			return true
		}
	}
	return false
}

// sortedTargetKeys returns target IDs in deterministic order.
func sortedTargetKeys(targets map[string]config.Target) []string {
	keys := make([]string, 0, len(targets))
	for k := range targets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func readSource(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func printJiraUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio jira <command> [flags]

Write commands:
  compile    Deprecated — use: jf push <KEY> <FILE>
  push       Compile + push description to Jira (delegates to jf)
  create     Create ticket + push description (delegates to jf for push)

Read commands:
  view       Fetch issue details
  search     Search issues by JQL

Run 'folio jira <command> --help' for details.
`)
}
