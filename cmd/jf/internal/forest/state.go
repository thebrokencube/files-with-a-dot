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

	// Mutability cache — content-addressed by local hash.
	MutableClean bool   `json:"mutable_clean,omitempty"`
	MutableHash  string `json:"mutable_hash,omitempty"`
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

// IsStale returns true if the file has been modified since last sync.
// Checks LastSync first (engine pipeline), falls back to LastPush
// for nodes synced before the engine migration.
func (s *State) IsStale(key string, fileMtime time.Time) bool {
	ns, ok := s.Nodes[key]
	if !ok {
		return true // never synced
	}
	ref := ns.LastSync
	if ref.IsZero() {
		ref = ns.LastPush
	}
	if ref.IsZero() {
		return true
	}
	return fileMtime.After(ref)
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

// MutabilityCache returns cached mutability for a node if the hash matches.
func (s *State) MutabilityCache(key, localHash string) (clean bool, found bool) {
	ns, ok := s.Nodes[key]
	if !ok || ns.MutableHash == "" || ns.MutableHash != localHash {
		return false, false
	}
	return ns.MutableClean, true
}

// SetMutability caches a mutability result keyed by content hash.
func (s *State) SetMutability(key, localHash string, clean bool) {
	ns := s.Nodes[key]
	ns.MutableClean = clean
	ns.MutableHash = localHash
	s.Nodes[key] = ns
}

// IsPullStale returns true if the node has never been pulled.
func (s *State) IsPullStale(key string) bool {
	ns, ok := s.Nodes[key]
	if !ok {
		return true
	}
	return ns.LastPull.IsZero()
}
