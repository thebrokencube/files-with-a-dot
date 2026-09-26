// Package lease gives a store's declared main dev server one cooperative
// holder: `folio lease run|status|release` decide from a small lease file under
// the umbrella joined with the OS's own answer to who owns each declared port.
package lease

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"gopkg.in/yaml.v3"
)

// Process is one long-running command of a declared server and the TCP port it listens on.
type Process struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
	Run  string `yaml:"run"`
}

// Serve is a store's `serve:` block.
type Serve struct {
	URL         string    `yaml:"url"`
	Processes   []Process `yaml:"processes"`
	RequireEnv  []string  `yaml:"require_env"`
	Alternative string    `yaml:"alternative"`
}

// PortConflict is a port this store declares that another store's serve block also declares.
type PortConflict struct {
	Port  int
	Other string
}

// Decode reads and validates a store's serve block. declared=false means the
// store has none, so no lease command applies to it.
func Decode(store config.Store) (s Serve, declared bool, err error) {
	if store.Serve.Kind == 0 {
		return Serve{}, false, nil
	}
	raw, err := yaml.Marshal(&store.Serve)
	if err != nil {
		return Serve{}, true, fmt.Errorf("store %q serve: %w", store.Name, err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&s); err != nil {
		return Serve{}, true, fmt.Errorf("store %q serve: %w", store.Name, err)
	}
	if len(s.Processes) == 0 {
		return Serve{}, true, fmt.Errorf("store %q serve: declare at least one process", store.Name)
	}
	seen := map[int]bool{}
	for i, p := range s.Processes {
		if p.Port < 1 || p.Port > 65535 {
			return Serve{}, true, fmt.Errorf("store %q serve: processes[%d] port %d is out of range", store.Name, i, p.Port)
		}
		if strings.TrimSpace(p.Run) == "" {
			return Serve{}, true, fmt.Errorf("store %q serve: processes[%d] has no run command", store.Name, i)
		}
		if seen[p.Port] {
			return Serve{}, true, fmt.Errorf("store %q serve: port %d is declared twice", store.Name, p.Port)
		}
		seen[p.Port] = true
	}
	for _, k := range s.RequireEnv {
		if k == "" {
			return Serve{}, true, fmt.Errorf("store %q serve: require_env has an empty name", store.Name)
		}
	}
	return s, true, nil
}

// findConflict returns the first declared port another store's serve block also
// declares. Another store's block is read leniently: its own errors surface
// only when that store is leased.
func findConflict(reg *config.Registry, name string, s Serve) *PortConflict {
	for _, other := range reg.AllStores() {
		if other.Name == name || other.Serve.Kind == 0 {
			continue
		}
		var o Serve
		if err := other.Serve.Decode(&o); err != nil {
			continue
		}
		for _, p := range s.Processes {
			for _, op := range o.Processes {
				if p.Port == op.Port {
					return &PortConflict{Port: p.Port, Other: other.Name}
				}
			}
		}
	}
	return nil
}

func (s Serve) where() string {
	if s.URL != "" {
		return s.URL
	}
	ports := make([]string, len(s.Processes))
	for i, p := range s.Processes {
		ports[i] = ":" + strconv.Itoa(p.Port)
	}
	return strings.Join(ports, ", ")
}
