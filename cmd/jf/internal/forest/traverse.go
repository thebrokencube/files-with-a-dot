package forest

import (
	"fmt"
	"strings"
)

// PostOrder returns nodes in post-order DFS (children before parents).
// Used for push — child content exists before parent references it.
func PostOrder(roots []*Node) []*Node {
	var result []*Node
	for _, r := range roots {
		postOrderWalk(r, &result)
	}
	return result
}

func postOrderWalk(n *Node, result *[]*Node) {
	for _, c := range n.Children {
		postOrderWalk(c, result)
	}
	*result = append(*result, n)
}

// PreOrder returns nodes in pre-order DFS (parents before children).
// Used for creation — Jira requires parent to exist before child.
func PreOrder(roots []*Node) []*Node {
	var result []*Node
	for _, r := range roots {
		preOrderWalk(r, &result)
	}
	return result
}

func preOrderWalk(n *Node, result *[]*Node) {
	*result = append(*result, n)
	for _, c := range n.Children {
		preOrderWalk(c, result)
	}
}

// Subtree returns the node matching target and all its descendants.
// Matches by key (case-insensitive) or filename stem.
func Subtree(roots []*Node, target string) (*Node, error) {
	node := resolve(roots, target)
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", target)
	}
	return node, nil
}

// Resolve returns a single node matching target by key or filename stem.
func Resolve(roots []*Node, target string) (*Node, error) {
	node := resolve(roots, target)
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", target)
	}
	return node, nil
}

func resolve(roots []*Node, target string) *Node {
	var all []*Node
	collectAll(roots, &all)

	upper := strings.ToUpper(target)

	// Try key match first (case-insensitive)
	for _, n := range all {
		if strings.ToUpper(n.Key) == upper {
			return n
		}
	}

	// Try filename stem match
	for _, n := range all {
		stem := fileStem(n.File)
		if strings.EqualFold(stem, target) {
			return n
		}
	}

	// Try file path match
	for _, n := range all {
		if n.File == target {
			return n
		}
	}

	return nil
}

func collectAll(nodes []*Node, out *[]*Node) {
	for _, n := range nodes {
		*out = append(*out, n)
		collectAll(n.Children, out)
	}
}

func fileStem(path string) string {
	base := path
	if i := strings.LastIndex(path, "/"); i >= 0 {
		base = path[i+1:]
	}
	if i := strings.LastIndex(base, "."); i >= 0 {
		base = base[:i]
	}
	return base
}
