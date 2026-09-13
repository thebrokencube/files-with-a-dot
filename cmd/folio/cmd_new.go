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
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runNew(folioPath string, noRegister, dryRun bool, artifactType, topicRaw string) int {
	pal := dendrik.NewPalette(true)
	topic := strings.ReplaceAll(topicRaw, " ", "-")

	// Vault artifacts are store-only and do not require a project manifest.
	if strings.HasPrefix(artifactType, "vault:") {
		return runNewVault(artifactType, topic, dryRun)
	}

	ctx, err := resolveContext(folioPath, contextMutation)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}
	folioPath = ctx.FolioPath

	if artifactType == "round" {
		return runNewRound(topic, folioPath, dryRun)
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
	folioDir := filepath.Dir(folioPath)

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

	// Colocatable types (design, retro, sketch): colocate inside work dir
	if !colocated && isColocatable(artifactType) {
		workDir := findWorkDir(folioDir, topic)
		if workDir == "" && (artifactType == "design" || artifactType == "sketch") {
			// design & sketch create the work dir when none exists
			date := time.Now().Format("2006-01-02")
			workDir = filepath.Join(folioDir, "work", "active", date+"-"+topic)
		}
		if workDir != "" {
			rel, _ := filepath.Rel(folioDir, workDir)
			if artifactType == "sketch" {
				// sketch is a fixed-name HTML page, not a dated .md
				relPath = filepath.Join(rel, "reference", "sketch", "index.html")
			} else {
				// Nested colocation: reference/<type>/YYYY-MM-DD-<topic>.md
				date := time.Now().Format("2006-01-02")
				relPath = filepath.Join(rel, "reference", artifactType, fmt.Sprintf("%s-%s.md", date, topic))
			}
			colocated = true
		}
	}

	absPath := filepath.Join(folioDir, relPath)

	// Check file doesn't already exist
	if _, err := os.Stat(absPath); err == nil {
		fmt.Fprintln(os.Stderr, pal.Errf("file already exists: %s", relPath))
		return dendrik.ExitUserError
	}

	if dryRun {
		fmt.Printf("Would create: %s\n", relPath)
		if colocated {
			fmt.Printf("  → colocated with %s/\n", filepath.Dir(relPath))
		}
		if !noRegister {
			fmt.Printf("  Would add source entry to folio.yml\n")
		}
		return dendrik.ExitOK
	}

	rawBefore, err := os.ReadFile(folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("reading folio.yml: %s", err))
		return dendrik.ExitUserError
	}
	beforeFolio, err := config.Parse(rawBefore)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}
	before := validate.Validate(beforeFolio, ctx, validate.Mutation)

	createdDirs := missingParentDirs(filepath.Dir(absPath), folioDir)
	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("creating directory: %s", err))
		return dendrik.ExitUserError
	}

	// Write template
	tmpl := taxonomy.Template(artifactType, topic)
	if artifactType == "sketch" {
		tmpl = taxonomy.SketchTemplate(topic)
	}
	if err := os.WriteFile(absPath, []byte(tmpl), 0644); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("writing file: %s", err))
		return dendrik.ExitUserError
	}

	// Register in folio.yml
	if !noRegister {
		if err := appendNewSource(folioPath, relPath); err != nil {
			rollbackNewRegistration(folioPath, rawBefore, absPath, createdDirs)
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
			return dendrik.ExitUserError
		}
		projected, loadErr := config.Load(folioPath)
		if loadErr != nil {
			rollbackNewRegistration(folioPath, rawBefore, absPath, createdDirs)
			fmt.Fprintln(os.Stderr, pal.Errf("validation failed — artifact not created: %s", loadErr))
			return dendrik.ExitUserError
		}
		after := validate.Validate(projected, ctx, validate.Mutation)
		delta := validate.Delta(before, after, nil)
		if !delta.Valid {
			rollbackNewRegistration(folioPath, rawBefore, absPath, createdDirs)
			fmt.Fprintln(os.Stderr, pal.Errf("validation failed — artifact not created:"))
			for _, validationErr := range delta.Errors {
				fmt.Fprintf(os.Stderr, "  - %s\n", validationErr)
			}
			return dendrik.ExitUserError
		}
	}

	fmt.Println(pal.Successf("Created %s", relPath))
	if colocated {
		fmt.Printf("  → colocated with %s/\n", filepath.Dir(relPath))
	}
	if !noRegister {
		fmt.Printf("  Added source entry to folio.yml\n")
	}
	return dendrik.ExitOK
}
func rollbackNewRegistration(folioPath string, original []byte, artifactPath string, createdDirs []string) {
	_ = os.WriteFile(folioPath, original, 0644)
	_ = os.Remove(artifactPath)
	for _, dir := range createdDirs {
		if err := os.Remove(dir); err != nil {
			break
		}
	}
}

func missingParentDirs(path, stop string) []string {
	var dirs []string
	for dir := path; dir != stop && dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		if _, err := os.Stat(dir); err == nil {
			break
		}
		dirs = append(dirs, dir)
	}
	return dirs
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

	workRoot, err := resolveFolioRootForInit()
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("cannot determine Folio work root: %s", err))
		return dendrik.ExitUserError
	}
	ctx := config.Context{WorkRoot: workRoot}

	today := time.Now().Format("2006-01-02")
	filename := today + "-" + topic + ".md"
	absPath, err := ctx.ResolvePath("vault:" + filepath.Join(label, filename))
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("cannot resolve vault path: %s", err))
		return dendrik.ExitUserError
	}

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

func isColocatable(t string) bool { return taxonomy.ColocatableTypes[t] }

func findWorkDir(folioDir, topic string) string {
	return taxonomy.FindWorkDir(folioDir, topic)
}
