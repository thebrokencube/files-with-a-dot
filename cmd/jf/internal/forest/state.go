package forest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// State tracks mutable per-node data in .jf/state.json.
type State struct {
	Nodes map[string]NodeState `json:"nodes"`
}

// NodeState tracks per-node push/pull state.
type NodeState struct {
	LastSync   time.Time `json:"last_sync,omitempty"`
	Direction  string    `json:"direction,omitempty"`
	LastPush   time.Time `json:"last_push,omitempty"`
	LastPull   time.Time `json:"last_pull,omitempty"`
	LocalHash  string    `json:"local_hash,omitempty"`  // sha256 of content below frontmatter
	RemoteHash string    `json:"remote_hash,omitempty"` // sha256 of ADF JSON from Jira
}

// LoadState reads .jf/state.json from the forest directory.
// Returns empty State if the file doesn't exist or is corrupt.
func LoadState(forestDir string) (*State, error) {
	path := filepath.Join(forestDir, ".jf", "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Nodes: make(map[string]NodeState)}, nil
		}
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ corrupt state.json, starting fresh: %s\n", err)
		return &State{Nodes: make(map[string]NodeState)}, nil
	}
	if s.Nodes == nil {
		s.Nodes = make(map[string]NodeState)
	}
	migrateState(&s)
	return &s, nil
}

// migrateState upgrades NodeState entries from LastPush/LastPull to LastSync.
// Idempotent: already-migrated entries are left unchanged.
func migrateState(s *State) {
	for key, ns := range s.Nodes {
		if !ns.LastSync.IsZero() {
			continue // already migrated
		}
		if ns.LastPush.IsZero() && ns.LastPull.IsZero() {
			continue // nothing to migrate
		}
		if ns.LastPull.IsZero() || ns.LastPush.After(ns.LastPull) {
			ns.LastSync = ns.LastPush
			ns.Direction = "push"
		} else {
			ns.LastSync = ns.LastPull
			ns.Direction = "pull"
		}
		s.Nodes[key] = ns
	}
}

// SaveState writes .jf/state.json atomically via tempfile + rename.
// Creates .jf/ directory if needed.
func SaveState(forestDir string, state *State) error {
	stateDir := filepath.Join(forestDir, ".jf")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(stateDir, "state-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, filepath.Join(stateDir, "state.json"))
}

// IsStale returns true if the file has been modified since last push.
// Comparison uses time.After with nanosecond precision. On filesystems
// with coarser mtime granularity (e.g., HFS+ at 1s), a file modified
// within the same second as a push may appear clean.
func (s *State) IsStale(key string, fileMtime time.Time) bool {
	ns, ok := s.Nodes[key]
	if !ok {
		return true // never pushed
	}
	return fileMtime.After(ns.LastPush)
}

// RecordPush updates the state for a node after a successful push.
func (s *State) RecordPush(key string, localHash, remoteHash string) {
	ns := s.Nodes[key]
	ns.LastPush = time.Now()
	ns.LocalHash = localHash
	ns.RemoteHash = remoteHash
	s.Nodes[key] = ns
}

// RecordPull updates the state for a node after a successful pull.
func (s *State) RecordPull(key string, localHash, remoteHash string) {
	ns := s.Nodes[key]
	ns.LastPull = time.Now()
	ns.LocalHash = localHash
	ns.RemoteHash = remoteHash
	s.Nodes[key] = ns
}

// RecordSync updates the state for a node after a successful sync operation.
func (s *State) RecordSync(key, direction, localHash, remoteHash string) {
	ns := s.Nodes[key]
	ns.LastSync = time.Now()
	ns.Direction = direction
	ns.LocalHash = localHash
	ns.RemoteHash = remoteHash
	s.Nodes[key] = ns
}

// ComputeHash returns the sha256 hex digest of the given content.
func ComputeHash(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}

// ConflictStatus represents the state of a bidirectional node.
type ConflictStatus int

const (
	ConflictNone       ConflictStatus = iota
	ConflictLocalOnly                 // local changed, remote unchanged
	ConflictRemoteOnly                // remote changed, local unchanged
	ConflictBoth                      // both sides changed
)

// DetectConflict compares current local and remote content against stored state.
func (s *State) DetectConflict(key string, localContent, remoteADF []byte) ConflictStatus {
	ns, ok := s.Nodes[key]
	if !ok {
		return ConflictNone // never synced
	}

	localChanged := ComputeHash(localContent) != ns.LocalHash
	remoteChanged := ComputeHash(remoteADF) != ns.RemoteHash

	switch {
	case localChanged && remoteChanged:
		return ConflictBoth
	case localChanged:
		return ConflictLocalOnly
	case remoteChanged:
		return ConflictRemoteOnly
	default:
		return ConflictNone
	}
}

// IsPullStale returns true if the node has never been pulled.
func (s *State) IsPullStale(key string) bool {
	ns, ok := s.Nodes[key]
	if !ok {
		return true
	}
	return ns.LastPull.IsZero()
}
