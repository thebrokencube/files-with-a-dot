package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/jira"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/touch"
)

func runJiraLint(args []string) int {
	fs := flag.NewFlagSet("jira lint", flag.ExitOnError)
	fs.Parse(args)

	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, output.Errf("no files specified"))
		fmt.Fprintf(os.Stderr, "Usage: folio jira lint <file...>\n")
		fmt.Fprintf(os.Stderr, "  Use - to read from stdin\n")
		return 1
	}

	hasIssues := false
	for _, file := range files {
		var input []byte
		var err error
		name := file

		if file == "-" {
			input, err = io.ReadAll(os.Stdin)
			name = "<stdin>"
		} else {
			input, err = os.ReadFile(file)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("read %s: %s", name, err))
			return 1
		}

		issues := jira.Lint(input, name)
		for _, iss := range issues {
			fmt.Fprintf(os.Stderr, "%s:%d: %s\n", name, iss.Line, iss.Message)
			hasIssues = true
		}
	}

	if hasIssues {
		return 1
	}
	return 0
}

func runJiraCompile(args []string) int {
	fs := flag.NewFlagSet("jira compile", flag.ExitOnError)
	id := fs.String("id", "", "Jira issue key (e.g., BEN-123)")
	source := fs.String("source", "", "Markdown source file (- for stdin)")
	outFile := fs.String("output", "", "Output file (default: stdout)")
	fs.Parse(args)

	if *id == "" || *source == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--id and --source are required"))
		fmt.Fprintf(os.Stderr, "Usage: folio jira compile --id KEY --source FILE [--output FILE]\n")
		return 1
	}

	input, err := readSource(*source)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	p := &jira.Pipeline{Run: jira.DefaultRunner}
	compiled, err := p.Compile(*id, input)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("compile: %s", err))
		return 1
	}

	if *outFile != "" {
		if err := os.WriteFile(*outFile, compiled, 0644); err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("write %s: %s", *outFile, err))
			return 1
		}
	} else {
		fmt.Println(string(compiled))
	}

	return 0
}

func runJiraPush(args []string) int {
	fs := flag.NewFlagSet("jira push", flag.ExitOnError)
	id := fs.String("id", "", "Jira issue key (e.g., BEN-123)")
	source := fs.String("source", "", "Markdown source file (- for stdin)")
	fs.Parse(args)

	if *id == "" || *source == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--id and --source are required"))
		fmt.Fprintf(os.Stderr, "Usage: folio jira push --id KEY --source FILE\n")
		return 1
	}

	input, err := readSource(*source)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	// Lint gate
	issues := jira.Lint(input, *source)
	if len(issues) > 0 {
		for _, iss := range issues {
			fmt.Fprintf(os.Stderr, "%s:%d: %s\n", *source, iss.Line, iss.Message)
		}
		fmt.Fprintln(os.Stderr, output.Errf("lint failed — fix issues before pushing"))
		return 1
	}

	p := &jira.Pipeline{Run: jira.DefaultRunner}

	compiled, err := p.Compile(*id, input)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("compile: %s", err))
		return 1
	}

	if err := p.Push(compiled); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("push: %s", err))
		return 1
	}

	fmt.Println(output.Successf("Pushed description for %s", *id))
	if touched, err := autoTouch(*source); err == nil && touched > 0 {
		fmt.Printf("  Auto-touched %d output(s)\n", touched)
	}
	return 0
}

func runJiraCreate(args []string) int {
	fs := flag.NewFlagSet("jira create", flag.ExitOnError)
	jsonFile := fs.String("json", "", "Creation JSON payload file")
	source := fs.String("source", "", "Markdown source file for description")
	fs.Parse(args)

	if *jsonFile == "" || *source == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--json and --source are required"))
		fmt.Fprintf(os.Stderr, "Usage: folio jira create --json FILE --source FILE\n")
		return 1
	}

	srcInput, err := readSource(*source)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	jsonPayload, err := readSource(*jsonFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("read %s: %s", *jsonFile, err))
		return 1
	}

	// Lint gate
	issues := jira.Lint(srcInput, *source)
	if len(issues) > 0 {
		for _, iss := range issues {
			fmt.Fprintf(os.Stderr, "%s:%d: %s\n", *source, iss.Line, iss.Message)
		}
		fmt.Fprintln(os.Stderr, output.Errf("lint failed — fix issues before creating"))
		return 1
	}

	p := &jira.Pipeline{Run: jira.DefaultRunner}

	// Create ticket (barebones, no description)
	key, err := p.Create(jsonPayload)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("create: %s", err))
		return 1
	}

	// Compile and push description using the new key
	compiled, err := p.Compile(key, srcInput)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("compile: %s", err))
		return 1
	}

	if err := p.Push(compiled); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("push description for %s: %s", key, err))
		return 1
	}

	fmt.Println(output.Successf("Created %s and pushed description", key))
	if touched, err := autoTouch(*source); err == nil && touched > 0 {
		fmt.Printf("  Auto-touched %d output(s)\n", touched)
	}
	return 0
}

func runJiraView(args []string) int {
	fs := flag.NewFlagSet("jira view", flag.ExitOnError)
	id := fs.String("id", "", "Jira issue key (e.g., BEN-123)")
	fields := fs.String("fields", "", "Comma-separated field list")
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Parse(args)

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
	fs := flag.NewFlagSet("jira search", flag.ExitOnError)
	jql := fs.String("jql", "", "JQL query string")
	fields := fs.String("fields", "", "Comma-separated field list")
	limit := fs.Int("limit", 50, "Max results")
	fs.Parse(args)

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
  lint       Validate markdown uses supported ADF subset
  compile    Convert markdown to acli-edit JSON
  push       Lint + convert + push description to Jira
  create     Create ticket + push description

Read commands:
  view       Fetch issue details
  search     Search issues by JQL

Run 'folio jira <command> --help' for details.
`)
}
