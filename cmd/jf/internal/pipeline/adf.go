package pipeline

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"jf/internal/forest"
	"os"
	"os/exec"
	"regexp"
	"sync"
)

//go:embed md2adf.bundle.mjs
var md2adfScript []byte

var (
	reFrontmatter = regexp.MustCompile(`^---\s*$`)

	scriptOnce sync.Once
	scriptFile string
	scriptErr  error
)

// StripFrontmatter removes YAML frontmatter from input if present.
// Only strips when: line 0 is exactly "---", a closing "---" appears
// within the first MaxFrontmatterLines, and at least one line between
// fences contains ":".
func StripFrontmatter(input []byte) []byte {
	lines := bytes.SplitN(input, []byte("\n"), -1)
	if len(lines) == 0 || !reFrontmatter.Match(lines[0]) {
		return input
	}

	limit := len(lines)
	if limit > forest.MaxFrontmatterLines {
		limit = forest.MaxFrontmatterLines
	}

	for i := 1; i < limit; i++ {
		if reFrontmatter.Match(lines[i]) {
			hasColon := false
			for j := 1; j < i; j++ {
				if bytes.Contains(lines[j], []byte(":")) {
					hasColon = true
					break
				}
			}
			if hasColon {
				return bytes.Join(lines[i+1:], []byte("\n"))
			}
			return input
		}
	}

	return input
}

// ensureScript writes the embedded md2adf bundle to a temp file once.
func ensureScript() (string, error) {
	scriptOnce.Do(func() {
		var f *os.File
		f, scriptErr = os.CreateTemp("", "jf-md2adf-*.mjs")
		if scriptErr != nil {
			return
		}
		if _, scriptErr = f.Write(md2adfScript); scriptErr != nil {
			f.Close()
			return
		}
		f.Close()
		scriptFile = f.Name()
	})
	if scriptErr != nil {
		return "", fmt.Errorf("write md2adf script: %w", scriptErr)
	}
	return scriptFile, nil
}

// CompileMarkdown converts markdown bytes to ADF JSON via marklassian.
// Strips frontmatter first. Returns the raw ADF document.
func CompileMarkdown(source []byte) (json.RawMessage, error) {
	stripped := StripFrontmatter(source)

	script, err := ensureScript()
	if err != nil {
		return nil, err
	}

	mdFile, err := os.CreateTemp("", "jf-md2adf-*.md")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(mdFile.Name())
	if _, err := mdFile.Write(stripped); err != nil {
		mdFile.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	mdFile.Close()

	cmd := exec.Command("node", script, mdFile.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("md2adf: %w\n%s", err, stderr.String())
	}

	var raw json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return nil, fmt.Errorf("parse md2adf output: %w", err)
	}

	return raw, nil
}
