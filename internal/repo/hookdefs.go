package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// HookDef is one entry in a remote repo's .pre-commit-hooks.yaml.
// Fields mirror the pre-commit spec.
type HookDef struct {
	ID                      string   `yaml:"id"`
	Name                    string   `yaml:"name"`
	Language                string   `yaml:"language"`
	Entry                   string   `yaml:"entry"`
	Args                    []string `yaml:"args"`
	Types                   []string `yaml:"types"`
	TypesOr                 []string `yaml:"types_or"`
	ExcludeTypes            []string `yaml:"exclude_types"`
	Files                   string   `yaml:"files"`
	Exclude                 string   `yaml:"exclude"`
	Stages                  []string `yaml:"stages"`
	PassFilenames           *bool    `yaml:"pass_filenames"`
	RequireSerial           bool     `yaml:"require_serial"`
	AdditionalDependencies  []string `yaml:"additional_dependencies"`
	Verbose                 bool     `yaml:"verbose"`
	AlwaysRun               bool     `yaml:"always_run"`
}

// LoadHookDefs reads .pre-commit-hooks.yaml from localRepoPath and returns a
// map keyed by hook ID.
func LoadHookDefs(localRepoPath string) (map[string]*HookDef, error) {
	p := filepath.Join(localRepoPath, ".pre-commit-hooks.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", p, err)
	}

	var defs []HookDef
	if err := yaml.Unmarshal(data, &defs); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", p, err)
	}

	out := make(map[string]*HookDef, len(defs))
	for i := range defs {
		out[defs[i].ID] = &defs[i]
	}
	return out, nil
}
