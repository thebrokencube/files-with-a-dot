package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/taxonomy"
)

func runNew(args []string) int {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	noRegister := fs.Bool("no-register", false, "Skip adding source entry to folio.yml")
	fs.Parse(args)

	if !resolveOrDie(folioPath) {
		return 1
	}

	if fs.NArg() < 2 {
		printNewUsage()
		return 1
	}

	artifactType := fs.Arg(0)
	topic := fs.Arg(1)

	// Validate type
	if !taxonomy.ValidTypes[artifactType] {
		fmt.Fprintln(os.Stderr, output.Errf("unknown type %q", artifactType))
		fmt.Fprintf(os.Stderr, "  Valid types: %s\n", validTypeList())
		return 1
	}

	// Resolve path
	relPath := taxonomy.TypePath(artifactType, topic)
	if relPath == "" {
		fmt.Fprintln(os.Stderr, output.Errf("cannot resolve path for type %q", artifactType))
		return 1
	}

	folioDir := filepath.Dir(*folioPath)
	absPath := filepath.Join(folioDir, relPath)

	// Check file doesn't already exist
	if _, err := os.Stat(absPath); err == nil {
		fmt.Fprintln(os.Stderr, output.Errf("file already exists: %s", relPath))
		return 1
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
	if !*noRegister {
		fmt.Printf("  Added source entry to folio.yml\n")
	}
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
	fmt.Fprintf(os.Stderr, `Usage: folio new <type> <topic> [--folio PATH] [--no-register]

Scaffold a typed artifact at the correct path.

Types:
  Reference: %s
  Work:      brief

Options:
  --folio PATH      Path or shortname (default: ./folio.yml)
  --no-register     Skip adding source entry to folio.yml
`, strings.Join(taxonomy.ReferenceTypes, ", "))
}
