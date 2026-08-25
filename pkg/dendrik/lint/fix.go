package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ApplyFixes applies the single mechanically-correct edit for each auto-fixable
// finding in results, re-deriving the edit from data (never from the prose
// Remediation). It returns the CheckIDs it fixed. Imperative adapter — mutates
// files under data.RepoRoot. Only the go.work `use` entries and the symlink_map
// skill line are fixable (both recomputed from ToolData); everything else is a
// no-op (see the switch default).
func ApplyFixes(data *ToolData, results []Result) ([]string, error) {
	var fixed []string
	done := map[string]bool{}
	for _, r := range results {
		if done[r.CheckID] {
			continue
		}
		var did bool
		var err error
		switch r.CheckID {
		case "go-work-sync", "go-mod-linked":
			did, err = fixGoWorkEntries(data)
		case "symlink-entries":
			did, err = fixSymlinkEntries(data)
		default:
			continue // not auto-fixable — reported, never mutated
		}
		if err != nil {
			return fixed, err
		}
		done[r.CheckID] = true
		if did {
			fixed = append(fixed, r.CheckID)
		}
	}
	return fixed, nil
}

// fixGoWorkEntries rewrites go.work's ./cmd/* use entries to exactly match the
// cmd/*/ directories that have a go.mod (data.CmdDirs), preserving every other
// line (e.g. ./pkg/dendrik). Resolves both go-work-sync and the go.work half of
// go-mod-linked. Idempotent.
func fixGoWorkEntries(data *ToolData) (bool, error) {
	if data.GoWork == nil {
		return false, nil // a missing go.work isn't mechanically fixable
	}
	desired := append([]string(nil), data.CmdDirs...)
	sort.Strings(desired)

	lines := strings.Split(string(data.GoWork), "\n")
	kept := make([]string, 0, len(lines))
	useIdx := -1
	for _, line := range lines {
		if goWorkUsePattern.MatchString(line) {
			continue // drop existing ./cmd/* entries; re-added below
		}
		kept = append(kept, line)
		if strings.TrimSpace(line) == "use (" {
			useIdx = len(kept) - 1
		}
	}
	if useIdx < 0 {
		return false, fmt.Errorf("go.work has no `use (` block to fix")
	}

	entries := make([]string, len(desired))
	for i, name := range desired {
		entries[i] = "\t./cmd/" + name
	}
	out := make([]string, 0, len(kept)+len(entries))
	out = append(out, kept[:useIdx+1]...)
	out = append(out, entries...)
	out = append(out, kept[useIdx+1:]...)

	newContent := strings.Join(out, "\n")
	if newContent == string(data.GoWork) {
		return false, nil
	}
	return true, os.WriteFile(filepath.Join(data.RepoRoot, "go.work"), []byte(newContent), 0o644)
}

// fixSymlinkEntries appends the missing skill symlink line for this tool to
// symlink_map.txt. Idempotent.
func fixSymlinkEntries(data *ToolData) (bool, error) {
	if data.SymlinkMap == nil {
		return false, nil
	}
	skillPath := "plugins/" + data.ToolName + "/skills/" + data.ToolName
	destPath := "$HOME/.claude/skills/" + data.ToolName
	for _, line := range strings.Split(string(data.SymlinkMap), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) == 2 && parts[0] == skillPath && parts[1] == destPath {
			return false, nil
		}
	}
	content := string(data.SymlinkMap)
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += skillPath + ":" + destPath + "\n"
	return true, os.WriteFile(filepath.Join(data.RepoRoot, "symlink_map.txt"), []byte(content), 0o644)
}
