package hooks

// Registry maps hook IDs to their built-in definitions.
// When a config references one of these IDs without an `entry:` field,
// the Go implementation is used directly — no subprocess required.
var Registry = map[string]BuiltinDef{
	"trailing-whitespace": {
		Name:          "Trim Trailing Whitespace",
		Fn:            builtinTrailingWhitespace,
		DefaultTypes:  []string{"text"},
		PassFilenames: true,
	},
	"end-of-file-fixer": {
		Name:          "Fix End of Files",
		Fn:            builtinEndOfFileFixer,
		DefaultTypes:  []string{"text"},
		PassFilenames: true,
	},
	"check-yaml": {
		Name:          "Check Yaml",
		Fn:            builtinCheckYAML,
		DefaultTypes:  []string{"yaml"},
		PassFilenames: true,
	},
	"check-json": {
		Name:          "Check JSON",
		Fn:            builtinCheckJSON,
		DefaultTypes:  []string{"json"},
		PassFilenames: true,
	},
	"check-toml": {
		Name:          "Check Toml",
		Fn:            builtinCheckTOML,
		DefaultTypes:  []string{"toml"},
		PassFilenames: true,
	},
	"check-merge-conflict": {
		Name:          "Check for merge conflicts",
		Fn:            builtinCheckMergeConflict,
		DefaultTypes:  []string{"text"},
		PassFilenames: true,
	},
	"detect-private-key": {
		Name:          "Detect Private Key",
		Fn:            builtinDetectPrivateKey,
		DefaultTypes:  []string{},
		PassFilenames: true,
	},
	"check-added-large-files": {
		Name:          "Check for added large files",
		Fn:            builtinCheckAddedLargeFiles,
		DefaultTypes:  []string{},
		PassFilenames: true,
	},
	"mixed-line-ending": {
		Name:          "Mixed line ending",
		Fn:            builtinMixedLineEnding,
		DefaultTypes:  []string{"text"},
		PassFilenames: true,
	},
	"check-case-conflict": {
		Name:          "Check for case conflicts",
		Fn:            builtinCheckCaseConflict,
		DefaultTypes:  []string{},
		PassFilenames: true,
	},
	"check-symlinks": {
		Name:          "Check for broken symlinks",
		Fn:            builtinCheckSymlinks,
		DefaultTypes:  []string{},
		PassFilenames: true,
	},
	"no-commit-to-branch": {
		Name:          "Don't commit to branch",
		Fn:            builtinNoCommitToBranch,
		DefaultTypes:  []string{},
		PassFilenames: false, // branch check, not file check
	},
	"check-executables-have-shebangs": {
		Name:          "Check that executables have shebangs",
		Fn:            builtinCheckExecutablesHaveShebangs,
		DefaultTypes:  []string{"text", "executable"},
		PassFilenames: true,
	},
}
