package lint

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// dirHasGoSource reports whether dir contains a non-test .go file.
func dirHasGoSource(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			return true
		}
	}
	return false
}

// GatherToolData is the imperative adapter for the lint core: it reads a tool
// directory (and the surrounding repo) off disk and returns the I/O-free
// ToolData bundle that Run and the per-layer checks operate on. This is the one
// I/O boundary of the lint verb — keep it thin and side-effect-only.
func GatherToolData(toolDir string) (*ToolData, error) {
	toolName := filepath.Base(toolDir)

	// Find repo root by walking up to find go.work
	repoRoot := toolDir
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "go.work")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			return nil, fmt.Errorf("could not find go.work in parent directories of %s", toolDir)
		}
		repoRoot = parent
	}

	data := &ToolData{
		ToolDir:  toolDir,
		ToolName: toolName,
		RepoRoot: repoRoot,
	}

	// Read files (nil if missing — that's fine, linters check for it)
	data.GoMod, _ = os.ReadFile(filepath.Join(toolDir, "go.mod"))
	data.GoWork, _ = os.ReadFile(filepath.Join(repoRoot, "go.work"))
	data.Makefile, _ = os.ReadFile(filepath.Join(toolDir, "Makefile"))
	data.SymlinkMap, _ = os.ReadFile(filepath.Join(repoRoot, "symlink_map.txt"))

	// README
	readmePath := filepath.Join(toolDir, "README.md")
	if content, err := os.ReadFile(readmePath); err == nil {
		data.HasREADME = true
		data.READMEBytes = content
	}

	// CLAUDE.md
	if _, err := os.Stat(filepath.Join(toolDir, "CLAUDE.md")); err == nil {
		data.HasCLAUDEMD = true
	}

	// docs/ directory
	docsDir := filepath.Join(toolDir, "docs")
	if docsEntries, err := os.ReadDir(docsDir); err == nil {
		for _, e := range docsEntries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				data.DocsFiles = append(data.DocsFiles, e.Name())
			}
		}
	}

	// Skill layer
	data.SkillDir = filepath.Join(toolDir, "skill")
	data.SkillMD, _ = os.ReadFile(filepath.Join(data.SkillDir, "SKILL.md"))

	data.RefContents = map[string][]byte{}
	refsDir := filepath.Join(data.SkillDir, "references")
	if entries, err := os.ReadDir(refsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				data.RefFiles = append(data.RefFiles, e.Name())
				if content, err := os.ReadFile(filepath.Join(refsDir, e.Name())); err == nil {
					data.RefContents[e.Name()] = content
				}
			}
		}
	}

	// Go files — parse all .go files in tool dir
	fset := token.NewFileSet()
	entries, err := os.ReadDir(toolDir)
	if err != nil {
		return nil, fmt.Errorf("reading tool directory: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		path := filepath.Join(toolDir, e.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		gf := GoFileData{
			Path:    e.Name(),
			Content: content,
		}
		astFile, parseErr := parser.ParseFile(fset, path, content, parser.AllErrors)
		if parseErr != nil {
			gf.Err = parseErr
		}
		gf.AST = astFile
		data.GoFiles = append(data.GoFiles, gf)
	}

	// Verb cores: pkg/dendrik subdirs with Go source. Shared sub-packages land
	// here harmlessly — core-in-pkg only checks verbs that have a cmd_<verb>.go.
	pkgDendrikDir := filepath.Join(repoRoot, "pkg", "dendrik")
	if subEntries, err := os.ReadDir(pkgDendrikDir); err == nil {
		for _, e := range subEntries {
			if !e.IsDir() {
				continue
			}
			if dirHasGoSource(filepath.Join(pkgDendrikDir, e.Name())) {
				data.PkgVerbCores = append(data.PkgVerbCores, e.Name())
			}
		}
	}

	// cmd/*/ directories with go.mod (for go-work-sync)
	cmdDir := filepath.Join(repoRoot, "cmd")
	if cmdEntries, err := os.ReadDir(cmdDir); err == nil {
		for _, e := range cmdEntries {
			if !e.IsDir() {
				continue
			}
			if _, err := os.Stat(filepath.Join(cmdDir, e.Name(), "go.mod")); err == nil {
				data.CmdDirs = append(data.CmdDirs, e.Name())
			}
		}
	}

	return data, nil
}
