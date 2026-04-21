package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const ConfigFile = ".pre-commit-config.yaml"

// ErrNotFound is returned when the config file does not exist.
// Callers can check with errors.Is(err, ErrNotFound).
var ErrNotFound = errors.New("config file not found")

// Config represents the top-level .pre-commit-config.yaml structure.
type Config struct {
	Repos    []Repo `yaml:"repos"`
	Exclude  string `yaml:"exclude"`
	FailFast bool   `yaml:"fail_fast"`
}

// Repo is a single entry under `repos:`.
// Use repo: local for hooks defined inline without a remote source.
type Repo struct {
	Repo  string `yaml:"repo"`
	Rev   string `yaml:"rev,omitempty"`
	Hooks []Hook `yaml:"hooks"`
}

// Hook configures a single hook within a repo.
type Hook struct {
	ID                    string   `yaml:"id"`
	Name                  string   `yaml:"name,omitempty"`
	Entry                 string   `yaml:"entry,omitempty"`    // command to run (system hooks)
	Language              string   `yaml:"language,omitempty"` // system, script, python, golang, node, ...
	Types                 []string `yaml:"types,omitempty"`    // file must match ALL
	TypesOr               []string `yaml:"types_or,omitempty"` // file must match ANY
	ExcludeTypes          []string `yaml:"exclude_types,omitempty"`
	Files                 string   `yaml:"files,omitempty"`   // regex filter
	Exclude               string   `yaml:"exclude,omitempty"` // regex exclude
	Args                  []string `yaml:"args,omitempty"`
	Stages                []string `yaml:"stages,omitempty"`
	AlwaysRun             bool     `yaml:"always_run,omitempty"`
	PassFilenames         *bool    `yaml:"pass_filenames,omitempty"`
	Verbose               bool     `yaml:"verbose,omitempty"`
	AdditionalDependencies []string `yaml:"additional_dependencies,omitempty"`
}

// ShouldPassFilenames returns true unless pass_filenames is explicitly false.
func (h *Hook) ShouldPassFilenames() bool {
	if h.PassFilenames == nil {
		return true
	}
	return *h.PassFilenames
}

// Load reads and parses the config file at path.
// Returns ErrNotFound (via errors.Is) when the file does not exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, path)
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return &cfg, nil
}
