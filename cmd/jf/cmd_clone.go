package main

import (
	"encoding/json"
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runClone(dir string, depth int, syncMode, repo, rootKey string) int {
	if syncMode != "" && syncMode != "push" && syncMode != "pull" && syncMode != "both" {
		fmt.Fprintf(os.Stderr, "✗ --sync must be 'push', 'pull', or 'both'\n")
		return dendrik.ExitUserError
	}

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	fmt.Printf("Fetching %s...\n", rootKey)
	root, err := fetchIssue(p, rootKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return dendrik.ExitExternalErr
	}

	tree, err := fetchTree(p, root, depth, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return dendrik.ExitExternalErr
	}

	total := countNodes(tree)
	fmt.Printf("Found %d total nodes\n\n", total)

	forestDir := filepath.Join(dir, ".jf")

	// Safety: refuse to overwrite existing forest
	if _, err := os.Stat(filepath.Join(forestDir, "forest.yml")); err == nil {
		fmt.Fprintf(os.Stderr, "✗ .jf/forest.yml already exists in %s\n  Remove .jf/ first or use a different --dir\n", dir)
		return dendrik.ExitUserError
	}

	if err := scaffoldTree(forestDir, tree, "", syncMode); err != nil {
		fmt.Fprintf(os.Stderr, "✗ scaffold error: %s\n", err)
		return dendrik.ExitUserError
	}

	if err := generateForestYAML(forestDir, root, syncMode, repo); err != nil {
		fmt.Fprintf(os.Stderr, "✗ forest.yml error: %s\n", err)
		return dendrik.ExitUserError
	}

	fmt.Printf("\nPulling descriptions...\n")
	pullCode := pullForest(forestDir, nil, "", false, false, true)

	fmt.Printf("\n✓ Forest ready at %s\n", forestDir)
	return pullCode
}

type cloneNode struct {
	Key      string
	Summary  string
	Type     string
	Children []*cloneNode
}

func fetchIssue(p *pipeline.Pipeline, key string) (*cloneNode, error) {
	out, err := p.View(key, "summary,issuetype", true)
	if err != nil {
		return nil, err
	}
	var issue struct {
		Fields struct {
			Summary   string `json:"summary"`
			IssueType struct {
				Name string `json:"name"`
			} `json:"issuetype"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(out, &issue); err != nil {
		return nil, fmt.Errorf("parse %s: %w", key, err)
	}
	return &cloneNode{
		Key:     key,
		Summary: issue.Fields.Summary,
		Type:    issue.Fields.IssueType.Name,
	}, nil
}

func fetchTree(p *pipeline.Pipeline, node *cloneNode, maxDepth, currentDepth int) (*cloneNode, error) {
	if maxDepth > 0 && currentDepth >= maxDepth {
		return node, nil
	}

	jql := fmt.Sprintf("parent = %s ORDER BY rank", node.Key)
	out, err := p.Search(jql, "summary,issuetype", 100, true)
	if err != nil {
		return node, nil // no children or search failed
	}

	children, err := parseSearchResults(out)
	if err != nil {
		return node, nil
	}

	for _, child := range children {
		fetched, err := fetchTree(p, child, maxDepth, currentDepth+1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ %s: %s\n", child.Key, err)
			continue
		}
		node.Children = append(node.Children, fetched)
	}

	return node, nil
}

func parseSearchResults(out []byte) ([]*cloneNode, error) {
	var issues []struct {
		Key    string `json:"key"`
		Fields struct {
			Summary   string `json:"summary"`
			IssueType struct {
				Name string `json:"name"`
			} `json:"issuetype"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, err
	}

	var nodes []*cloneNode
	for _, iss := range issues {
		nodes = append(nodes, &cloneNode{
			Key:     iss.Key,
			Summary: iss.Fields.Summary,
			Type:    iss.Fields.IssueType.Name,
		})
	}
	return nodes, nil
}

var (
	reBracketPrefix = regexp.MustCompile(`^\[.*?\]\s*`)
	reNonAlnum      = regexp.MustCompile(`[^a-z0-9]+`)
)

func slugify(summary string) string {
	s := reBracketPrefix.ReplaceAllString(summary, "")
	s = strings.ToLower(strings.TrimSpace(s))
	s = reNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "forest"
	}
	return s
}

func cloneFrontmatter(key, summary, syncMode string) string {
	escapedSummary := strings.ReplaceAll(summary, "\"", "\\\"")
	if syncMode == "" {
		return fmt.Sprintf("---\njira: %s\nlabel: \"%s\"\n---\n", key, escapedSummary)
	}
	return fmt.Sprintf("---\njira: %s\nlabel: \"%s\"\nsync: %s\n---\n", key, escapedSummary, syncMode)
}

func scaffoldTree(baseDir string, node *cloneNode, relPath, syncMode string) error {
	dirPath := filepath.Join(baseDir, relPath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(dirPath, "README.md")
	fm := cloneFrontmatter(node.Key, node.Summary, syncMode)
	if err := os.WriteFile(filePath, []byte(fm), 0644); err != nil {
		return err
	}

	fmt.Printf("  %s  <- %s\n", filepath.Join(relPath, "README.md"), node.Key)

	for _, child := range node.Children {
		if len(child.Children) > 0 {
			childDir := slugify(child.Summary)
			if err := scaffoldTree(baseDir, child, filepath.Join(relPath, childDir), syncMode); err != nil {
				return err
			}
		} else {
			leafFile := filepath.Join(dirPath, child.Key+".md")
			fm := cloneFrontmatter(child.Key, child.Summary, syncMode)
			if err := os.WriteFile(leafFile, []byte(fm), 0644); err != nil {
				return err
			}
			fmt.Printf("  %s  <- %s\n", filepath.Join(relPath, child.Key+".md"), child.Key)
		}
	}
	return nil
}

func generateForestYAML(forestDir string, root *cloneNode, syncMode, repo string) error {
	project := strings.Split(root.Key, "-")[0]
	var yml string
	if syncMode == "" {
		yml = fmt.Sprintf("schema: 1\ndefaults:\n  project: %s\n", project)
	} else {
		yml = fmt.Sprintf("schema: 1\ndefaults:\n  sync: %s\n  project: %s\n", syncMode, project)
	}
	if repo != "" {
		yml = strings.TrimRight(yml, "\n") + fmt.Sprintf("\n  repos:\n    - %s\n", repo)
	}
	return os.WriteFile(filepath.Join(forestDir, "forest.yml"), []byte(yml), 0644)
}

func countNodes(node *cloneNode) int {
	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}
