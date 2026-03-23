package main

import (
	"bytes"
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runCreateMissing(args []string) int {
	fs := dendrik.NewFlagSet("create-missing")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")
	dryRun := fs.Bool('n', "dry-run", "Show what would be created without side effects")
	force := fs.Bool('f', "force", "Push as plain text if marklassian conversion fails")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	f, roots, code := loadForestOrFail(*dir, false)
	if code != 0 {
		return code
	}

	issues := forest.Validate(roots, f)
	hasErrors := false
	for _, iss := range issues {
		if iss.Level == "error" {
			fmt.Fprintln(os.Stderr, iss.String())
			hasErrors = true
		}
	}
	if hasErrors {
		return dendrik.ExitUserError
	}

	// Collect TBD nodes in pre-order (parents before children)
	ordered := forest.PreOrder(roots)
	var tbdNodes []*forest.Node
	for _, n := range ordered {
		if forest.IsTBD(n.Key) {
			tbdNodes = append(tbdNodes, n)
		}
	}

	if len(tbdNodes) == 0 {
		fmt.Println("No TBD nodes found.")
		return dendrik.ExitOK
	}

	if *dryRun {
		return dryRunCreate(tbdNodes, f)
	}

	return executeCreate(tbdNodes, f, *force)
}

func dryRunCreate(nodes []*forest.Node, f *forest.Forest) int {
	fmt.Printf("Would create %d ticket(s):\n\n", len(nodes))
	for i, n := range nodes {
		parentKey := "(root)"
		if n.Parent != nil && !forest.IsTBD(n.Parent.Key) {
			parentKey = n.Parent.Key
		}
		fmt.Printf("  %d. %s\n", i+1, n.Label)
		fmt.Printf("     Type: %s, Project: %s, Parent: %s\n", n.Type, f.Defaults.Project, parentKey)
		fmt.Printf("     File: %s\n", n.File)
	}
	return dendrik.ExitOK
}

func executeCreate(nodes []*forest.Node, f *forest.Forest, force bool) int {
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	// Track failed parents so we skip their children
	failedKeys := make(map[*forest.Node]bool)
	succeeded := 0
	failed := 0

	for _, n := range nodes {
		// Skip if parent failed (can't create child without parent key)
		if n.Parent != nil && failedKeys[n.Parent] {
			fmt.Fprintf(os.Stderr, "⊘ %s: skipped (parent creation failed)\n", n.Label)
			failedKeys[n] = true
			failed++
			continue
		}

		// Dedup check
		jql := fmt.Sprintf(`project = %s AND summary ~ %q`, f.Defaults.Project, n.Label)
		existingKey, err := dedupCheck(p, jql)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ %s: dedup check failed: %s\n", n.Label, err)
			// Continue with creation despite dedup failure
		}

		var newKey string
		if existingKey != "" {
			fmt.Fprintf(os.Stderr, "⚠ %s already exists for %q — using existing key\n", existingKey, n.Label)
			newKey = existingKey
		} else {
			payload := buildCreatePayload(n, f)
			newKey, err = p.Create(payload)
			if err != nil {
				fmt.Fprintf(os.Stderr, "✗ %s: creation failed: %s\n", n.Label, err)
				failedKeys[n] = true
				failed++
				continue
			}
		}

		// Update the node's key in memory (so children see the real key)
		n.Key = newKey

		// Rewrite frontmatter in source file
		filePath := filepath.Join(f.Dir, n.File)
		if err := rewriteFrontmatterKey(filePath, newKey); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ %s: frontmatter rewrite failed: %s\n", newKey, err)
			// Not fatal — ticket was created, just frontmatter not updated
		}

		// Compile and push description (skip if empty)
		source, err := os.ReadFile(filePath)
		if err == nil {
			stripped := pipeline.StripFrontmatter(source)
			if len(bytes.TrimSpace(stripped)) == 0 {
				fmt.Printf("✓ Created %s %q (%s) — description empty, push skipped\n", newKey, n.Label, n.File)
				succeeded++
				continue
			}
			compiled, compileErr := p.Compile(newKey, source, "")
			if compileErr != nil {
				if force {
					fmt.Fprintf(os.Stderr, "⚠ %s: conversion failed, pushing as plain text\n", newKey)
					compiled = buildPlainTextPayload(newKey, source)
				} else {
					fmt.Fprintf(os.Stderr, "⚠ %s: description push skipped (compile failed: %s)\n", newKey, compileErr)
					compiled = nil
				}
			}
			if compiled != nil {
				if pushErr := p.Push(compiled); pushErr != nil {
					fmt.Fprintf(os.Stderr, "⚠ %s: description push failed: %s\n", newKey, pushErr)
				}
			}
		}

		fmt.Printf("✓ Created %s %q (%s)\n", newKey, n.Label, n.File)
		succeeded++
	}

	fmt.Printf("\nCreated %d/%d tickets", succeeded, succeeded+failed)
	if failed > 0 {
		fmt.Printf(" (%d failed)", failed)
	}
	fmt.Println()

	if failed > 0 {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}

// dedupCheck searches Jira for an existing ticket matching the JQL.
// Returns the key if found, empty string if not.
func dedupCheck(p *pipeline.Pipeline, jql string) (string, error) {
	out, err := p.Search(jql, "summary", 1, false)
	if err != nil {
		return "", err
	}

	// acli search returns lines like "BEN-123  Summary text"
	// Look for a Jira key in the output
	key := pipeline.ExtractJiraKey(out)
	return key, nil
}

// buildCreatePayload builds the JSON payload for acli create.
func buildCreatePayload(n *forest.Node, f *forest.Forest) []byte {
	payload := fmt.Sprintf(`{"project":%q,"type":%q,"summary":%q`, f.Defaults.Project, n.Type, n.Label)

	// Add parent link if parent has a real key
	if n.Parent != nil && !forest.IsTBD(n.Parent.Key) {
		payload += fmt.Sprintf(`,"parent":{"key":%q}`, n.Parent.Key)
	}

	payload += "}"
	return []byte(payload)
}

// rewriteFrontmatterKey does a line-level replacement of jira: TBD with the new key.
// No YAML roundtrip — preserves formatting, comments, and content below frontmatter.
func rewriteFrontmatterKey(filePath, newKey string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	rewritten, changed := rewriteTBDLine(content, newKey)
	if !changed {
		return fmt.Errorf("no jira: TBD line found in frontmatter")
	}

	return os.WriteFile(filePath, rewritten, 0644)
}

// rewriteTBDLine replaces the first jira: TBD (or jira: "TBD") line in YAML frontmatter.
// Returns the modified content and whether a replacement was made.
func rewriteTBDLine(content []byte, newKey string) ([]byte, bool) {
	lines := bytes.Split(content, []byte("\n"))
	if len(lines) == 0 || strings.TrimSpace(string(lines[0])) != "---" {
		return content, false
	}

	// First pass: find closing fence to confirm valid frontmatter
	closingIdx := -1
	for i := 1; i < len(lines) && i < forest.MaxFrontmatterLines; i++ {
		if strings.TrimSpace(string(lines[i])) == "---" {
			closingIdx = i
			break
		}
	}
	if closingIdx < 0 {
		return content, false
	}

	// Second pass: find and replace jira: TBD within frontmatter bounds
	for i := 1; i < closingIdx; i++ {
		line := string(lines[i])
		trimmed := strings.TrimSpace(line)

		if isTBDLine(trimmed) {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			lines[i] = []byte(indent + "jira: " + newKey)
			return bytes.Join(lines, []byte("\n")), true
		}
	}

	return content, false
}

// isTBDLine checks if a trimmed line is a jira: TBD declaration.
func isTBDLine(trimmed string) bool {
	lower := strings.ToLower(trimmed)
	if !strings.HasPrefix(lower, "jira:") {
		return false
	}
	value := strings.TrimSpace(trimmed[5:])
	return strings.EqualFold(value, "TBD") ||
		strings.EqualFold(value, `"TBD"`) ||
		strings.EqualFold(value, `'TBD'`)
}
