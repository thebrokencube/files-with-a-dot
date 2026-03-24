package forest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FindForest walks up from startDir looking for forest.yml.
// Returns parsed Forest or nil if not found.
func FindForest(startDir string) (*Forest, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}

	for {
		path := filepath.Join(dir, "forest.yml")
		if _, err := os.Stat(path); err == nil {
			return ParseForestFile(path)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, nil
		}
		dir = parent
	}
}

// Discover scans the forest directory tree and builds the node hierarchy.
// A file is a node iff it has a jira: field in YAML frontmatter.
// README.md with jira: = parent for its directory's other jira: files.
// Directory without jira: README.md = pass-through (children attach to
// nearest ancestor that IS a node).
func Discover(forest *Forest) ([]*Node, error) {
	if forest == nil {
		return nil, fmt.Errorf("forest is nil")
	}

	type discovered struct {
		node     *Node
		dir      string
		isREADME bool
	}

	var nodes []discovered

	err := filepath.Walk(forest.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories (e.g., .jf/, .git/).
		// Hidden *files* (e.g., .secret.md) are intentionally discovered
		// if they have jira: frontmatter — only directories are skipped.
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}

		// Only process .md files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Skip forest.yml sibling files that aren't in subdirectories
		// (forest.yml itself is not .md so already skipped)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		fm, err := ParseFrontmatter(content)
		if err != nil {
			// Malformed frontmatter: skip with warning (not fatal)
			fmt.Fprintf(os.Stderr, "⚠ %s: %s\n", path, err)
			return nil
		}

		if fm == nil {
			return nil // No jira: field, not a node
		}

		rel, err := filepath.Rel(forest.Dir, path)
		if err != nil {
			return err
		}

		node := &Node{
			Key:   fm.Jira,
			Label: DeriveLabel(fm, content, rel),
			Type:  fm.Type,
			Sync:  fm.Sync,
			Order: fm.Order,
			File:  rel,
		}

		// Apply defaults where frontmatter is empty
		if node.Type == "" {
			node.Type = forest.Defaults.Type
		}
		if node.Sync == "" {
			node.Sync = forest.Defaults.Sync
		}
		if node.Sync == "" {
			node.Sync = "both"
		}

		isREADME := strings.ToUpper(filepath.Base(path)) == "README.MD"

		nodes = append(nodes, discovered{
			node:     node,
			dir:      filepath.Dir(rel),
			isREADME: isREADME,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Build directory → README node map (these are potential parents)
	dirParent := make(map[string]*Node)
	for _, d := range nodes {
		if d.isREADME {
			dirParent[d.dir] = d.node
		}
	}

	// Attach children to parents
	var roots []*Node
	for _, d := range nodes {
		if d.isREADME {
			// README nodes attach to parent directory's README
			parent := findAncestorParent(d.dir, dirParent)
			if parent != nil {
				d.node.Parent = parent
				parent.Children = append(parent.Children, d.node)
			} else {
				roots = append(roots, d.node)
			}
		} else {
			// Non-README nodes attach to their directory's README or nearest ancestor
			parent := dirParent[d.dir]
			if parent == nil {
				parent = findAncestorParent(d.dir, dirParent)
			}
			if parent != nil {
				d.node.Parent = parent
				parent.Children = append(parent.Children, d.node)
			} else {
				roots = append(roots, d.node)
			}
		}
	}

	// Sort children by order (then alphabetically by label)
	sortChildren(roots)

	return roots, nil
}

// findAncestorParent walks up directory components to find the nearest
// ancestor directory that has a README.md node.
func findAncestorParent(dir string, dirParent map[string]*Node) *Node {
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
		if node, ok := dirParent[dir]; ok {
			return node
		}
		if dir == "." {
			return nil
		}
	}
}

// sortChildren recursively sorts node children by order, then label.
func sortChildren(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Order != nodes[j].Order {
			return nodes[i].Order < nodes[j].Order
		}
		return nodes[i].Label < nodes[j].Label
	})
	for _, n := range nodes {
		if len(n.Children) > 0 {
			sortChildren(n.Children)
		}
	}
}
