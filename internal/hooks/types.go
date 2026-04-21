package hooks

// HookContext carries runtime information available to built-in hooks.
type HookContext struct {
	RepoPath string
}

// Result is the outcome of a single hook execution.
type Result struct {
	Pass     bool
	Output   string
	Modified []string // files modified by a fixer hook
}

// BuiltinFn is the function signature every built-in hook must implement.
type BuiltinFn func(ctx *HookContext, files []string, args []string) *Result

// BuiltinDef describes a built-in hook entry in the registry.
type BuiltinDef struct {
	// Name is the human-readable label shown in output.
	Name string

	// Fn is the Go implementation.
	Fn BuiltinFn

	// DefaultTypes are the file types this hook runs on when the config
	// does not specify types. Empty means "all files".
	DefaultTypes []string

	// PassFilenames indicates whether the hook receives file paths.
	// Set to false for hooks that only inspect repository state (e.g. branch name).
	PassFilenames bool
}
