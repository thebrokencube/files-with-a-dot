package graph

// PropagateStaleness performs transitive staleness propagation across the DAG.
// Takes a map of target ID → status (clean/stale/missing/unknown) and the
// adjacency list (target → upstream deps). Returns an updated status map
// and a map of target → upstream target that caused staleness.
func PropagateStaleness(statuses map[string]string, adj map[string][]string) (map[string]string, map[string]string) {
	result := make(map[string]string)
	for k, v := range statuses {
		result[k] = v
	}
	causedBy := make(map[string]string)

	// Convergence loop: keep propagating until no changes
	changed := true
	for changed {
		changed = false
		for tid, deps := range adj {
			if result[tid] == "clean" {
				for _, dep := range deps {
					depStatus := result[dep]
					if depStatus == "stale" || depStatus == "missing" || depStatus == "unknown" {
						result[tid] = "stale"
						causedBy[tid] = dep
						changed = true
						break
					}
				}
			}
		}
	}

	return result, causedBy
}
