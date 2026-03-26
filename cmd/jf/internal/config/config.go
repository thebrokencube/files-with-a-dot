package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the ~/.jf.yml configuration file.
type Config struct {
	Site        string                       `yaml:"site,omitempty"`
	CloudID     string                       `yaml:"cloud_id,omitempty"`
	Projects    map[string]map[string]any    `yaml:"projects,omitempty"`
	ParkingLots map[string]ParkingLotConfig  `yaml:"parking_lots,omitempty"`
}

// ParkingLotConfig holds per-project parking lot settings.
type ParkingLotConfig struct {
	Epic   string `yaml:"epic"`
	Status string `yaml:"status"`
}

// Path returns the default config file path (~/.jf.yml).
func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jf.yml")
}

// Load reads the config from ~/.jf.yml.
// Returns empty config if file doesn't exist.
func Load() (*Config, error) {
	return LoadFrom(Path())
}

// LoadFrom reads the config from a specific path.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to ~/.jf.yml.
func (c *Config) Save() error {
	return c.SaveTo(Path())
}

// SaveTo writes the config to a specific path.
func (c *Config) SaveTo(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// BrowseURL returns the Jira browse URL for an issue key.
// Returns empty string if site is not configured.
func (c *Config) BrowseURL(key string) string {
	if c.Site == "" {
		return ""
	}
	return "https://" + c.Site + ".atlassian.net/browse/" + key
}
