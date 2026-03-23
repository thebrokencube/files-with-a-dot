package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/taxonomy"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runNew(args []string) int {
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("new")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	noRegister := fs.BoolLong("no-register", "Skip adding source entry to folio.yml")
	dryRun := fs.Bool('n', "dry-run", "Print what would be created, no side effects")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if len(fs.GetArgs()) < 2 {
		printNewUsage()
		return dendrik.ExitUserError
	}

	artifactType := fs.GetArgs()[0]
	topic := strings.ReplaceAll(fs.GetArgs()[1], " ", "-")

	// Handle vault: prefix — scaffold directly in vault directory (no folio.yml needed)
	if strings.HasPrefix(artifactType, "vault:") {
		return runNewVault(artifactType, topic, *dryRun)
	}

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	if artifactType == "round" {
		return runNewRound(topic, *folioPath, *dryRun)
	}

	// Deprecation check
	if artifactType == "note" {
		fmt.Fprintln(os.Stderr, pal.Errf("\"note\" was removed — use \"spike\" for exploratory content, or a specific type (domain, guide, retro)"))
		return dendrik.ExitUserError
	}

	// Validate type
	if !taxonomy.ValidTypes[artifactType] {
		fmt.Fprintln(os.Stderr, pal.Errf("unknown type %q", artifactType))
		fmt.Fprintf(os.Stderr, "  Valid types: %s\n", validTypeList())
		return dendrik.ExitUserError
	}

	// Resolve path
	folioDir := filepath.Dir(*folioPath)

	relPath := taxonomy.TypePath(artifactType, topic)
	if relPath == "" {
		fmt.Fprintln(os.Stderr, pal.Errf("cannot resolve path for type %q", artifactType))
		return dendrik.ExitUserError
	}

	colocated := false

	// Plan/brief: use existing work dir if one matches the topic
	if artifactType == "plan" || artifactType == "brief" {
		if workDir := findWorkDir(folioDir, topic); workDir != "" {
			rel, _ := filepath.Rel(folioDir, workDir)
			relPath = filepath.Join(rel, "README.md")
			colocated = true
		}
	}

	// Colocatable types (design, retro): colocate inside work dir
	if !colocated && isColocatable(artifactType) {
		workDir := findWorkDir(folioDir, topic)
		if workDir == "" && artifactType == "design" {
			// Design creates the work directory
			date := time.Now().Format("2006-01-02")
			workDir = filepath.Join(folioDir, "work", "active", date+"-"+topic)
		}
		if workDir != "" {
			rel, _ := filepath.Rel(folioDir, workDir)
			// Nested colocation: reference/<type>/YYYY-MM-DD-<topic>.md
			date := time.Now().Format("2006-01-02")
			relPath = filepath.Join(rel, "reference", artifactType, fmt.Sprintf("%s-%s.md", date, topic))
			colocated = true
		}
	}

	absPath := filepath.Join(folioDir, relPath)

	// Check file doesn't already exist
	if _, err := os.Stat(absPath); err == nil {
		fmt.Fprintln(os.Stderr, pal.Errf("file already exists: %s", relPath))
		return dendrik.ExitUserError
	}

	if *dryRun {
		fmt.Printf("Would create: %s\n", relPath)
		if colocated {
			fmt.Printf("  → colocated with %s/\n", filepath.Dir(relPath))
		}
		if !*noRegister {
			fmt.Printf("  Would add source entry to folio.yml\n")
		}
		return dendrik.ExitOK
	}

	// Validate folio.yml parses before modifying
	if _, err := config.Load(*folioPath); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("creating directory: %s", err))
		return dendrik.ExitUserError
	}

	// Write template
	tmpl := taxonomy.Template(artifactType, topic)
	if err := os.WriteFile(absPath, []byte(tmpl), 0644); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("writing file: %s", err))
		return dendrik.ExitUserError
	}

	// Register in folio.yml
	if !*noRegister {
		if err := appendNewSource(*folioPath, relPath); err != nil {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
			return dendrik.ExitUserError
		}
	}

	fmt.Println(pal.Successf("Created %s", relPath))
	if colocated {
		fmt.Printf("  → colocated with %s/\n", filepath.Dir(relPath))
	}
	if !*noRegister {
		fmt.Printf("  Added source entry to folio.yml\n")
	}
	return dendrik.ExitOK
}

func runNewRound(topic, folioPath string, dryRun bool) int {
	pal := dendrik.NewPalette(true)
	folioDir := filepath.Dir(folioPath)
	workDir := taxonomy.FindWorkDir(folioDir, topic)
	if workDir == "" {
		fmt.Fprintln(os.Stderr, pal.Errf("no work directory found for topic %q", topic))
		return dendrik.ExitUserError
	}
	if strings.Contains(workDir, string(filepath.Separator)+"work"+string(filepath.Separator)+"archive"+string(filepath.Separator)) {
		fmt.Fprintln(os.Stderr, pal.Errf("work directory for %q is archived — rounds are active work only", topic))
		return dendrik.ExitUserError
	}

	agentDir := filepath.Join(workDir, "agent-research")
	maxNum := 0
	matches, _ := filepath.Glob(filepath.Join(agentDir, "????-round"))
	for _, m := range matches {
		base := filepath.Base(m)
		prefix := base[:4]
		if n, err := strconv.Atoi(prefix); err == nil && n > maxNum {
			maxNum = n
		}
	}

	nextNum := maxNum + 1
	roundName := fmt.Sprintf("%04d-round", nextNum)
	roundDir := filepath.Join(agentDir, roundName)
	relPath, _ := filepath.Rel(folioDir, roundDir)

	if dryRun {
		fmt.Printf("Would create: %s\n", relPath)
		return dendrik.ExitOK
	}

	if err := os.MkdirAll(roundDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("creating directory: %s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Created %s", relPath))
	return dendrik.ExitOK
}

var validVaultLabels = map[string]bool{
	"research": true,
	"domain":   true,
	"guide":    true,
	"insight":  true,
}

func runNewVault(artifactType, topic string, dryRun bool) int {
	pal := dendrik.NewPalette(true)
	label := strings.TrimPrefix(artifactType, "vault:")
	if !validVaultLabels[label] {
		fmt.Fprintln(os.Stderr, pal.Errf("unknown vault label %q", label))
		fmt.Fprintf(os.Stderr, "  Valid labels: research, domain, guide, insight\n")
		return dendrik.ExitUserError
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("cannot determine home directory: %s", err))
		return dendrik.ExitUserError
	}

	today := time.Now().Format("2006-01-02")
	filename := today + "-" + topic + ".md"
	absPath := filepath.Join(home, ".folio", "vault", label, filename)

	if _, err := os.Stat(absPath); err == nil {
		fmt.Fprintln(os.Stderr, pal.Errf("file already exists: vault/%s/%s", label, filename))
		return dendrik.ExitUserError
	}

	if dryRun {
		fmt.Printf("Would create: vault/%s/%s\n", label, filename)
		fmt.Printf("  No folio.yml registration (vault has no manifest)\n")
		return dendrik.ExitOK
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("creating directory: %s", err))
		return dendrik.ExitUserError
	}

	// Use the reference template for the label type
	tmpl := taxonomy.Template(label, topic)
	if err := os.WriteFile(absPath, []byte(tmpl), 0644); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("writing file: %s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Created vault/%s/%s", label, filename))
	return dendrik.ExitOK
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
  Convention:  round (auto-incrementing agent-research dir under work dir)

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
