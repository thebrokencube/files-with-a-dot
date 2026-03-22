package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/taxonomy"
)

func runNew(args []string) int {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	noRegister := fs.Bool("no-register", false, "Skip adding source entry to folio.yml")
	dryRun := fs.Bool("dry-run", false, "Print what would be created, no side effects")
	parseFlags(fs, args)

	if !resolveOrDie(folioPath) {
		return 1
	}

	if fs.NArg() < 2 {
		printNewUsage()
		return 1
	}

	artifactType := fs.Arg(0)
	topic := fs.Arg(1)

	// Handle vault: prefix — scaffold directly in vault directory
	if strings.HasPrefix(artifactType, "vault:") {
		return runNewVault(artifactType, topic, *dryRun)
	}

	// Deprecation check
	if artifactType == "note" {
		fmt.Fprintln(os.Stderr, output.Errf("\"note\" was removed — use \"spike\" for exploratory content, or a specific type (domain, guide, retro)"))
		return 1
	}

	// Validate type
	if !taxonomy.ValidTypes[artifactType] {
		fmt.Fprintln(os.Stderr, output.Errf("unknown type %q", artifactType))
		fmt.Fprintf(os.Stderr, "  Valid types: %s\n", validTypeList())
		return 1
	}

	// Resolve path
	folioDir := filepath.Dir(*folioPath)

	relPath := taxonomy.TypePath(artifactType, topic)
	if relPath == "" {
		fmt.Fprintln(os.Stderr, output.Errf("cannot resolve path for type %q", artifactType))
		return 1
	}

	colocated := false
	if isColocatable(artifactType) {
		if workDir := findWorkDir(folioDir, topic); workDir != "" {
			rel, _ := filepath.Rel(folioDir, workDir)
			relPath = filepath.Join(rel, artifactType+".md")
			colocated = true
		}
	}

	absPath := filepath.Join(folioDir, relPath)

	// Check file doesn't already exist
	if _, err := os.Stat(absPath); err == nil {
		fmt.Fprintln(os.Stderr, output.Errf("file already exists: %s", relPath))
		return 1
	}

	if *dryRun {
		fmt.Printf("Would create: %s\n", relPath)
		if colocated {
			fmt.Printf("  → colocated with %s/\n", filepath.Dir(relPath))
		}
		if !*noRegister {
			fmt.Printf("  Would add source entry to folio.yml\n")
		}
		return 0
	}

	// Validate folio.yml parses before modifying
	if _, err := config.Load(*folioPath); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("creating directory: %s", err))
		return 1
	}

	// Write template
	tmpl := taxonomy.Template(artifactType, topic)
	if err := os.WriteFile(absPath, []byte(tmpl), 0644); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("writing file: %s", err))
		return 1
	}

	// Register in folio.yml
	if !*noRegister {
		if err := appendNewSource(*folioPath, relPath); err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("%s", err))
			return 1
		}
	}

	fmt.Println(output.Successf("Created %s", relPath))
	if colocated {
		fmt.Printf("  → colocated with %s/\n", filepath.Dir(relPath))
	}
	if !*noRegister {
		fmt.Printf("  Added source entry to folio.yml\n")
	}
	return 0
}

var validVaultLabels = map[string]bool{
	"research": true,
	"domain":   true,
	"guide":    true,
	"insight":  true,
}

func runNewVault(artifactType, topic string, dryRun bool) int {
	label := strings.TrimPrefix(artifactType, "vault:")
	if !validVaultLabels[label] {
		fmt.Fprintln(os.Stderr, output.Errf("unknown vault label %q", label))
		fmt.Fprintf(os.Stderr, "  Valid labels: research, domain, guide, insight\n")
		return 1
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("cannot determine home directory: %s", err))
		return 1
	}

	today := time.Now().Format("2006-01-02")
	filename := today + "-" + topic + ".md"
	absPath := filepath.Join(home, ".folio", "vault", label, filename)

	if _, err := os.Stat(absPath); err == nil {
		fmt.Fprintln(os.Stderr, output.Errf("file already exists: vault/%s/%s", label, filename))
		return 1
	}

	if dryRun {
		fmt.Printf("Would create: vault/%s/%s\n", label, filename)
		fmt.Printf("  No folio.yml registration (vault has no manifest)\n")
		return 0
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("creating directory: %s", err))
		return 1
	}

	// Use the reference template for the label type
	tmpl := taxonomy.Template(label, topic)
	if err := os.WriteFile(absPath, []byte(tmpl), 0644); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("writing file: %s", err))
		return 1
	}

	fmt.Println(output.Successf("Created vault/%s/%s", label, filename))
	return 0
}

// appendNewSource adds a path source entry to folio.yml.
func appendNewSource(folioYmlPath, relPath string) error {
	lines, _, insertIdx, indent, err := findSourcesInsertPoint(folioYmlPath)
	if err != nil {
		return err
	}

	newLines := []string{
		fmt.Sprintf("%s- path: %s", indent, relPath),
	}

	return insertLines(folioYmlPath, lines, insertIdx, newLines)
}

func validTypeList() string {
	types := make([]string, 0, len(taxonomy.ValidTypes))
	for t := range taxonomy.ValidTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return strings.Join(types, ", ")
}

func printNewUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio new <type> <topic> [--folio PATH] [--no-register] [--dry-run]

Scaffold a typed artifact at the correct path.

Types:
  Lifecycle:   spike, design, plan, retro
  Reference:   %s
  Vault:       vault:research, vault:domain, vault:guide, vault:insight
  Alias:       brief (-> plan)
  Dual-layer:  design, retro (colocate with work dir if one matches topic)

Options:
  --folio PATH      Path or shortname (default: ./folio.yml)
  --no-register     Skip adding source entry to folio.yml
  --dry-run         Print what would be created, no side effects
`, strings.Join(taxonomy.ReferenceTypes, ", "))
}

func isColocatable(t string) bool { return taxonomy.ColocatableTypes[t] }

func findWorkDir(folioDir, topic string) string {
	return taxonomy.FindWorkDir(folioDir, topic)
}
