package graph

import "sort"

// StatusRank is the sole ordering for local freshness states.
func StatusRank(status string) int {
	switch status {
	case "clean", "final":
		return 0
	case "unknown":
		return 1
	case "stale":
		return 2
	case "missing":
		return 3
	default:
		return -1
	}
}

// PropagateStaleness propagates local status through target dependencies.
func PropagateStaleness(statuses map[string]string, adj map[string][]string, terminal map[string]bool) (map[string]string, map[string]string) {
	result := make(map[string]string, len(statuses))
	for target, status := range statuses {
		result[target] = status
	}
	causedBy := make(map[string]string)
	targets := make([]string, 0, len(adj))
	for target := range adj {
		targets = append(targets, target)
	}
	sort.Strings(targets)

	changed := true
	for changed {
		changed = false
		for _, target := range targets {
			if terminal[target] {
				continue
			}
			current, ok := result[target]
			if !ok {
				continue
			}
			for _, dependency := range adj[target] {
				candidate := propagatedStatus(result[dependency])
				if StatusRank(candidate) > StatusRank(current) {
					result[target] = candidate
					causedBy[target] = dependency
					current = candidate
					changed = true
				}
			}
		}
	}
	return result, causedBy
}

func propagatedStatus(status string) string {
	if status == "missing" {
		return "stale"
	}
	return status
}
