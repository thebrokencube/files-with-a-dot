// Package conventions provides machine-readable convention data for dendrik CLIs.
//
// This package is the source of truth for cross-CLI conventions. It is consumed
// by the dendrik lint contract (see contract.go) for validation
// and dendrik new (Track 5) for scaffolding.
package conventions

//dendrik:block flag-registry
//dendrik:kind component
//dendrik:layer bridge
//dendrik:status shipped
//dendrik:definition shared CLI flag vocabulary (short/long/meaning) + collision detection
//dendrik:intent flags mean the same across every dendrik CLI; collisions caught at build

// FlagEntry describes a short flag's meaning within a specific CLI.
type FlagEntry struct {
	Short    byte   // Single-character flag (e.g., 'f')
	Long     string // Long form (e.g., "force")
	CLI      string // Which CLI uses this (e.g., "jf", "folio")
	Commands string // Which commands use it ("all", "push,sync", etc.)
	Meaning  string // Human-readable description
}

// GlobalFlags are short flags with fixed meaning across all dendrik CLIs.
// These letters must not be reused for CLI-specific purposes.
var GlobalFlags = []FlagEntry{
	{'h', "help", "*", "all", "Show help"},
	{'j', "json", "*", "read commands", "Machine-readable JSON output"},
	{'n', "dry-run", "*", "write commands", "Preview without side effects"},
}

// CLIFlags are short flags specific to individual CLIs.
// The same letter may have different meanings in different CLIs.
var CLIFlags = []FlagEntry{
	// jf flags
	{'c', "check", "jf", "setup", "Non-interactive check only"},
	{'d', "dir", "jf", "most commands", "Directory to scan/work in"},
	{'f', "force", "jf", "push,sync,create-missing", "Overwrite on conflict or conversion failure"},
	{'l', "limit", "jf", "search", "Maximum results"},
	{'o', "output", "jf", "schema", "Emit output schemas instead of input"},
	{'p', "project", "jf", "search,init", "Jira project key"},
	{'r', "resolve", "jf", "sync", "Conflict resolution strategy"},
	{'s', "subtree", "jf", "push", "Push node and all descendants"},
	{'t', "type", "jf", "search", "Filter by issue type"},
	{'v', "verbose", "jf", "tree", "Show sync direction and file paths"},

	// folio flags
	{'a', "all", "folio", "home push", "Stage all changes"},
	{'b', "branches", "folio", "dag", "Show branch topology"},
	{'f', "folio", "folio", "most commands", "Path or shortname to folio.yml"},
	{'i', "id", "folio", "jira compile,push,view", "Jira issue key"},
	{'l', "limit", "folio", "jira search", "Maximum results"},
	{'m', "materialize", "folio", "gather", "Create reference file stub"},
	{'m', "message", "folio", "home push", "Commit message"},
	{'q', "jql", "folio", "jira search", "JQL query string"},
	{'r', "read", "folio", "gather", "Read and summarize URL"},
	{'s', "source", "folio", "jira compile,push,create", "Markdown source file"},
	{'s', "status", "folio", "dag", "Show staleness overlay"},
	{'s', "scope", "folio", "observe list", "Filter by scope"},
	{'t', "type", "folio", "gather,observe list", "Reference or observation type"},
}

// Collisions returns flag letters that have different meanings across CLIs.
func Collisions() map[byte][]FlagEntry {
	byLetter := map[byte][]FlagEntry{}
	for _, f := range CLIFlags {
		byLetter[f.Short] = append(byLetter[f.Short], f)
	}

	collisions := map[byte][]FlagEntry{}
	for letter, entries := range byLetter {
		// Check if any entries span different CLIs with different long names
		seen := map[string]string{} // cli -> long
		for _, e := range entries {
			if prev, ok := seen[e.CLI]; ok && prev != e.Long {
				// Same CLI, different long — this is intra-CLI reuse (ok if different commands)
				continue
			}
			seen[e.CLI] = e.Long
		}
		// If multiple CLIs use this letter with different long names, it's a collision
		longs := map[string]bool{}
		for _, l := range seen {
			longs[l] = true
		}
		if len(longs) > 1 {
			collisions[letter] = entries
		}
	}

	return collisions
}

// IsGlobalFlag returns true if the given short flag letter is globally reserved.
func IsGlobalFlag(short byte) bool {
	for _, f := range GlobalFlags {
		if f.Short == short {
			return true
		}
	}
	return false
}
