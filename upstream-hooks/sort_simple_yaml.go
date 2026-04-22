package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// RunSortSimpleYAML is the Go equivalent of sort_simple_yaml.py.
// Sorts top-level keys of simple YAML files (root must be a plain mapping).
// Returns 0 if no files changed, 1 if any file was sorted.
func RunSortSimpleYAML(args []string) int {
	fs := flag.NewFlagSet("sort-simple-yaml", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	retv := 0
	for _, f := range fs.Args() {
		if sortSimpleYAML(f) {
			fmt.Printf("Sorting %s\n", f)
			retv = 1
		}
	}
	return retv
}

// sortSimpleYAML sorts the top-level keys of the YAML mapping in filename.
// Returns true if the file was modified.
func sortSimpleYAML(filename string) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return false
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return false
	}
	mapping := root.Content[0]
	if mapping.Kind != yaml.MappingNode || len(mapping.Content)%2 != 0 {
		return false
	}

	// Pair up key+value nodes.
	type kv struct{ key, val *yaml.Node }
	pairs := make([]kv, 0, len(mapping.Content)/2)
	for i := 0; i < len(mapping.Content); i += 2 {
		pairs = append(pairs, kv{mapping.Content[i], mapping.Content[i+1]})
	}

	sorted := make([]kv, len(pairs))
	copy(sorted, pairs)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].key.Value) < strings.ToLower(sorted[j].key.Value)
	})

	// Check if already sorted.
	alreadySorted := true
	for i, p := range pairs {
		if p.key.Value != sorted[i].key.Value {
			alreadySorted = false
			break
		}
	}
	if alreadySorted {
		return false
	}

	flat := make([]*yaml.Node, 0, len(mapping.Content))
	for _, p := range sorted {
		flat = append(flat, p.key, p.val)
	}
	mapping.Content = flat

	out, err := yaml.Marshal(&root)
	if err != nil {
		return false
	}
	_ = os.WriteFile(filename, out, 0o644)
	return true
}
