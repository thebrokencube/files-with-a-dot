package forest

import (
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
	LastPush time.Time `json:"last_push,omitempty"`
	FileHash string    `json:"file_hash,omitempty"`
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
	return &s, nil
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
func (s *State) IsStale(key string, fileMtime time.Time) bool {
	ns, ok := s.Nodes[key]
	if !ok {
		return true // never pushed
	}
	return fileMtime.After(ns.LastPush)
}

// RecordPush updates the state for a node after a successful push.
func (s *State) RecordPush(key string) {
	s.Nodes[key] = NodeState{
		LastPush: time.Now(),
	}
}
