package graph

import "sort"

// DetectCycle performs DFS cycle detection on an adjacency list.
// Returns the full cycle path (e.g., [a, b, c, a]) or nil if no cycle exists.
// Nodes are visited in sorted order for deterministic output.
func DetectCycle(adj map[string][]string) []string {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	// Collect and sort all nodes for deterministic iteration
	nodeSet := make(map[string]bool)
	for node, deps := range adj {
		nodeSet[node] = true
		for _, dep := range deps {
			nodeSet[dep] = true
		}
	}
	nodes := make([]string, 0, len(nodeSet))
	for n := range nodeSet {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)

	var path []string

	var dfs func(node string) []string
	dfs = func(node string) []string {
		visited[node] = true
		inStack[node] = true
		path = append(path, node)

		for _, dep := range adj[node] {
			if inStack[dep] {
				// Found back edge: extract cycle from path
				for i, n := range path {
					if n == dep {
						cycle := make([]string, len(path[i:]))
						copy(cycle, path[i:])
						cycle = append(cycle, dep)
						return cycle
					}
				}
			}
			if !visited[dep] {
				if cycle := dfs(dep); cycle != nil {
					return cycle
				}
			}
		}

		path = path[:len(path)-1]
		inStack[node] = false
		return nil
	}

	for _, node := range nodes {
		if !visited[node] {
			if cycle := dfs(node); cycle != nil {
				return cycle
			}
		}
	}

	return nil
}
