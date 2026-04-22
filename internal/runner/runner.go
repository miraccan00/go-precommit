package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/miraccan00/go-precommit/internal/config"
	"github.com/miraccan00/go-precommit/internal/hooks"
	"github.com/miraccan00/go-precommit/internal/lang"
	"github.com/miraccan00/go-precommit/internal/repo"
)

const lineWidth = 76

// Runner orchestrates hook execution for a single `run` invocation.
type Runner struct {
	cfg     *config.Config
	verbose bool
	ctx     *hooks.HookContext
	repoMgr *repo.Manager
	envBase string // base dir for language environments
}

// New creates a Runner. If the repo manager cannot be initialised (e.g. no
// home dir), remote repos will be skipped with a warning.
func New(cfg *config.Config, verbose bool) *Runner {
	mgr, _ := repo.NewManager() // nil on error — handled in runOne

	base, _ := os.UserCacheDir()
	envBase := filepath.Join(base, "go-precommit", "envs")
	_ = os.MkdirAll(envBase, 0o755)

	return &Runner{
		cfg:     cfg,
		verbose: verbose,
		ctx:     &hooks.HookContext{RepoPath: "."},
		repoMgr: mgr,
		envBase: envBase,
	}
}

// Run executes all configured hooks against files.
// hookFilter, when non-empty, restricts execution to the listed hook IDs.
// Returns true only when every executed hook passes.
func (r *Runner) Run(files []string, hookFilter []string, stage string) bool {
	allPassed := true

	for _, repoCfg := range r.cfg.Repos {
		for _, hookCfg := range repoCfg.Hooks {
			if len(hookFilter) > 0 && !sliceContains(hookFilter, hookCfg.ID) {
				continue
			}
			if !matchesStage(hookCfg.Stages, stage) {
				continue
			}

			passed := r.runOne(repoCfg.Repo, repoCfg.Rev, &hookCfg, files)
			if !passed {
				allPassed = false
				if r.cfg.FailFast {
					return false
				}
			}
		}
	}

	return allPassed
}

// ─── effective hook ───────────────────────────────────────────────────────────

// effectiveHook holds the fully-resolved hook configuration after merging the
// remote hook definition (from .pre-commit-hooks.yaml) with the local config
// overrides.
type effectiveHook struct {
	id             string
	name           string
	entry          string
	language       string
	types          []string
	typesOr        []string
	excludeTypes   []string
	files          string
	exclude        string
	args           []string
	alwaysRun      bool
	passFilenames  bool
	additionalDeps []string
	repoLocalPath  string // only set for remote repos
}

// mergeHook produces an effectiveHook from an optional remote HookDef and a
// local config Hook. Config values win over hook def values; additional_dependencies
// are additive.
func mergeHook(def *repo.HookDef, cfg *config.Hook) effectiveHook {
	e := effectiveHook{
		id:            cfg.ID,
		passFilenames: true, // default
	}

	// Apply hook def first (remote repo's .pre-commit-hooks.yaml).
	if def != nil {
		e.name = def.Name
		e.entry = def.Entry
		e.language = def.Language
		e.types = def.Types
		e.typesOr = def.TypesOr
		e.excludeTypes = def.ExcludeTypes
		e.files = def.Files
		e.exclude = def.Exclude
		e.args = def.Args
		e.alwaysRun = def.AlwaysRun
		e.additionalDeps = def.AdditionalDependencies
		if def.PassFilenames != nil {
			e.passFilenames = *def.PassFilenames
		}
	}

	// Config overrides (local .pre-commit-config.yaml values win).
	if cfg.Name != "" {
		e.name = cfg.Name
	}
	if cfg.Entry != "" {
		e.entry = cfg.Entry
	}
	if cfg.Language != "" {
		e.language = cfg.Language
	}
	if len(cfg.Types) > 0 {
		e.types = cfg.Types
	}
	if len(cfg.TypesOr) > 0 {
		e.typesOr = cfg.TypesOr
	}
	if len(cfg.ExcludeTypes) > 0 {
		e.excludeTypes = cfg.ExcludeTypes
	}
	if cfg.Files != "" {
		e.files = cfg.Files
	}
	if cfg.Exclude != "" {
		e.exclude = cfg.Exclude
	}
	if len(cfg.Args) > 0 {
		e.args = cfg.Args // replace, not append
	}
	if cfg.AlwaysRun {
		e.alwaysRun = true
	}
	if cfg.PassFilenames != nil {
		e.passFilenames = *cfg.PassFilenames
	}
	// additional_dependencies from config are additive.
	e.additionalDeps = append(e.additionalDeps, cfg.AdditionalDependencies...)

	// Fallback name.
	if e.name == "" {
		e.name = cfg.ID
	}

	return e
}

// ─── runOne ───────────────────────────────────────────────────────────────────

func (r *Runner) runOne(repoURL, repoRev string, hookCfg *config.Hook, files []string) bool {
	isLocal := repoURL == "local" || repoURL == ""

	var def *repo.HookDef
	var localRepoPath string

	if !isLocal {
		// Remote repo — clone / fetch from cache.
		if r.repoMgr == nil {
			printLine(hookCfg.ID, "Skipped", color.New(color.FgYellow).SprintFunc())
			fmt.Fprintf(os.Stderr, "  warning: repo manager unavailable, skipping remote hook %q\n", hookCfg.ID)
			return true
		}

		var err error
		localRepoPath, err = r.repoMgr.LocalPath(repoURL, repoRev)
		if err != nil {
			printLine(hookCfg.ID, "Error", color.New(color.FgRed).SprintFunc())
			fmt.Fprintf(os.Stderr, "  error cloning %s: %v\n", repoURL, err)
			return false
		}

		defs, err := repo.LoadHookDefs(localRepoPath)
		if err != nil {
			printLine(hookCfg.ID, "Error", color.New(color.FgRed).SprintFunc())
			fmt.Fprintf(os.Stderr, "  error loading hook defs from %s: %v\n", repoURL, err)
			return false
		}

		d, ok := defs[hookCfg.ID]
		if !ok {
			printLine(hookCfg.ID, "Error", color.New(color.FgRed).SprintFunc())
			fmt.Fprintf(os.Stderr, "  hook %q not found in %s\n", hookCfg.ID, repoURL)
			return false
		}
		def = d
	}

	eff := mergeHook(def, hookCfg)
	eff.repoLocalPath = localRepoPath

	// ── Go built-in fast path ────────────────────────────────────────────────
	// Built-ins are always preferred when available. The entry: field in
	// .pre-commit-config.yaml exists so the Python pre-commit tool knows how
	// to invoke go-precommit; it is not an override for go-precommit's own
	// execution (using it would cause infinite recursion).
	builtinDef, isBuiltin := hooks.Registry[eff.id]
	if isBuiltin {
		return r.runBuiltin(builtinDef, eff, files)
	}

	// ── External command (system / language env) ─────────────────────────────
	// Reached only when:
	//   • hook ID has no Go builtin, OR
	//   • the config explicitly overrides entry: to use a specific command.
	return r.runExternal(eff, files)
}

// ─── built-in execution ───────────────────────────────────────────────────────

func (r *Runner) runBuiltin(def hooks.BuiltinDef, eff effectiveHook, files []string) bool {
	effectiveTypes := eff.types
	if len(effectiveTypes) == 0 {
		effectiveTypes = def.DefaultTypes
	}

	passFilenames := eff.passFilenames
	// If the config didn't explicitly set pass_filenames, use the builtin default.
	// (eff already contains the merged value, so we honour it as-is.)

	filtered := r.filterFiles(files, effectiveTypes, eff.typesOr, eff.excludeTypes, eff.files, eff.exclude)
	if len(filtered) == 0 && !eff.alwaysRun && passFilenames {
		printLine(eff.name, "Skipped", color.New(color.FgYellow).SprintFunc())
		return true
	}

	filesToPass := filtered
	if !passFilenames {
		filesToPass = nil
	}

	result := def.Fn(r.ctx, filesToPass, eff.args)
	return r.report(eff.id, eff.name, result)
}

// ─── external command execution ───────────────────────────────────────────────

func (r *Runner) runExternal(eff effectiveHook, files []string) bool {
	if eff.entry == "" {
		printLine(eff.name, "Skipped", color.New(color.FgYellow).SprintFunc())
		fmt.Fprintf(os.Stderr, "  warning: no entry for hook %q\n", eff.id)
		return true
	}

	// Set up language environment if needed.
	binDir, err := r.ensureEnv(eff)
	if err != nil {
		printLine(eff.name, "Error", color.New(color.FgRed).SprintFunc())
		fmt.Fprintf(os.Stderr, "  env setup failed for %q: %v\n", eff.id, err)
		return false
	}

	filtered := r.filterFiles(files, eff.types, eff.typesOr, eff.excludeTypes, eff.files, eff.exclude)
	if len(filtered) == 0 && !eff.alwaysRun && eff.passFilenames {
		printLine(eff.name, "Skipped", color.New(color.FgYellow).SprintFunc())
		return true
	}

	result := r.execCommand(eff.entry, filtered, eff.args, eff.passFilenames, binDir)
	return r.report(eff.id, eff.name, result)
}

// ensureEnv sets up the language environment and returns its bin directory.
// For system/script languages it returns "" (use PATH as-is).
func (r *Runner) ensureEnv(eff effectiveHook) (string, error) {
	if eff.repoLocalPath == "" || eff.language == "" ||
		eff.language == "system" || eff.language == "script" {
		return "", nil
	}

	key := lang.EnvCacheKey(eff.repoLocalPath, eff.language, eff.additionalDeps)
	envDir := filepath.Join(r.envBase, key)

	if err := lang.Setup(eff.language, eff.repoLocalPath, envDir, eff.additionalDeps); err != nil {
		return "", err
	}

	return lang.BinDir(eff.language, envDir), nil
}

// execCommand runs an entry command, optionally prepending binDir to PATH.
func (r *Runner) execCommand(entry string, files, args []string, passFilenames bool, binDir string) *hooks.Result {
	parts := strings.Fields(entry)
	if len(parts) == 0 {
		return &hooks.Result{Pass: false, Output: "empty entry command"}
	}

	cmdArgs := make([]string, 0, len(parts[1:])+len(args)+len(files))
	cmdArgs = append(cmdArgs, parts[1:]...)
	cmdArgs = append(cmdArgs, args...)
	if passFilenames {
		cmdArgs = append(cmdArgs, files...)
	}

	// Resolve the executable. exec.Command uses the current process PATH for
	// lookup — it does NOT use cmd.Env. So when binDir is set, we explicitly
	// look for the binary there first; fall back to system PATH otherwise.
	executable := parts[0]
	if binDir != "" {
		candidate := filepath.Join(binDir, parts[0])
		if _, err := os.Stat(candidate); err == nil {
			executable = candidate
		}
	}

	cmd := exec.Command(executable, cmdArgs...)

	// Still expose binDir via PATH so that the subprocess itself (e.g. python
	// calling another script) can find sibling executables.
	if binDir != "" {
		newPath := binDir + string(os.PathListSeparator) + os.Getenv("PATH")
		cmd.Env = append(os.Environ(), "PATH="+newPath)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return &hooks.Result{Pass: false, Output: strings.TrimSpace(string(out))}
	}
	return &hooks.Result{Pass: true, Output: strings.TrimSpace(string(out))}
}

// ─── output ───────────────────────────────────────────────────────────────────

func (r *Runner) report(id, name string, result *hooks.Result) bool {
	if result.Pass {
		printLine(name, "Passed", color.New(color.FgGreen).SprintFunc())
		if r.verbose && result.Output != "" {
			fmt.Println(indent(result.Output))
		}
		return true
	}

	printLine(name, "Failed", color.New(color.FgRed).SprintFunc())
	fmt.Printf("- hook id: %s\n", id)

	if len(result.Modified) > 0 {
		fmt.Println("- files were modified by this hook")
		for _, f := range result.Modified {
			fmt.Printf("  %s\n", f)
		}
	}
	if result.Output != "" {
		fmt.Println()
		fmt.Println(result.Output)
		fmt.Println()
	}
	return false
}

func printLine(name, status string, statusColor func(...interface{}) string) {
	const dots = ".................................................................."
	if len(name) > 50 {
		name = name[:47] + "..."
	}
	dotCount := lineWidth - len(name) - len(status)
	if dotCount < 3 {
		dotCount = 3
	}
	if dotCount > len(dots) {
		dotCount = len(dots)
	}
	fmt.Printf("%s%s%s\n", name, dots[:dotCount], statusColor(status))
}

// ─── file filtering ───────────────────────────────────────────────────────────

// filterFiles returns the subset of files that pass all active filters.
//
//   types     — file must have ALL listed types (AND)
//   typesOr   — file must have at least ONE listed type (OR)
//   excludeTypes — file must NOT have ANY listed type
//   filesRx   — file path must match regex
//   excludeRx — file path must NOT match regex
func (r *Runner) filterFiles(files, types, typesOr, excludeTypes []string, filesRx, excludeRx string) []string {
	var fRe, eRe *regexp.Regexp
	if filesRx != "" {
		fRe, _ = regexp.Compile(filesRx)
	}
	if excludeRx != "" {
		eRe, _ = regexp.Compile(excludeRx)
	}

	var result []string
	for _, f := range files {
		// Global config exclude.
		if r.cfg.Exclude != "" {
			if re, err := regexp.Compile(r.cfg.Exclude); err == nil && re.MatchString(f) {
				continue
			}
		}

		if fRe != nil && !fRe.MatchString(f) {
			continue
		}
		if eRe != nil && eRe.MatchString(f) {
			continue
		}

		detected := hooks.DetectFileTypes(f)
		detectedSet := make(map[string]bool, len(detected))
		for _, t := range detected {
			detectedSet[t] = true
		}

		// types: ALL must match.
		if len(types) > 0 {
			allMatch := true
			for _, t := range types {
				if !detectedSet[t] {
					allMatch = false
					break
				}
			}
			if !allMatch {
				continue
			}
		}

		// types_or: at least ONE must match.
		if len(typesOr) > 0 {
			anyMatch := false
			for _, t := range typesOr {
				if detectedSet[t] {
					anyMatch = true
					break
				}
			}
			if !anyMatch {
				continue
			}
		}

		// exclude_types: NONE must match.
		if len(excludeTypes) > 0 {
			excluded := false
			for _, t := range excludeTypes {
				if detectedSet[t] {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		result = append(result, f)
	}
	return result
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// matchesStage reports whether a hook should run for the given stage.
// - stage ""           → run everything (no filtering)
// - stage "pre-commit" → run hooks with no stages set, or stages containing "pre-commit"
// - stage "pre-push"   → run only hooks with stages containing "pre-push"
func matchesStage(hookStages []string, stage string) bool {
	if stage == "" {
		return true
	}
	if len(hookStages) == 0 {
		return stage == "pre-commit"
	}
	return sliceContains(hookStages, stage)
}

func sliceContains(s []string, v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}
