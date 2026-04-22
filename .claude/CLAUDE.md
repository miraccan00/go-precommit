# go-precommit — Claude context

Fast, dependency-free alternative to the Python `pre-commit` tool. Ships the most common hooks as native Go code in a single static binary.

Module: `github.com/miraccan00/go-precommit`
Go version: 1.25

---

## Repository layout

```
main.go                        — binary entry point (delegates to cmd/)
cmd/
  root.go                      — cobra root command
  run.go                       — `go-precommit run` — main entry point for hook execution
  install.go                   — `go-precommit install` — writes .git/hooks/pre-commit
  global_install.go            — `go-precommit global-install`

internal/
  config/config.go             — parses .pre-commit-config.yaml
  hooks/
    types.go                   — HookContext, Result, BuiltinFn, BuiltinDef
    registry.go                — map[hookID]BuiltinDef — all built-in Go hooks
    builtin.go                 — implementations of every built-in hook
    filetypes.go               — DetectFileTypes() — extension/content-based tagging
  runner/runner.go             — orchestrates hook execution; file filtering; output
  gitutil/gitutil.go           — StagedFiles(), AllFiles(), CurrentBranch()
  repo/
    manager.go                 — remote repo clone/cache (~/.cache/go-precommit/repos/)
    hookdefs.go                — loads .pre-commit-hooks.yaml from a cloned repo
  lang/setup.go                — language env setup (python venv, go install, npm)

upstream-hooks/                — standalone Go equivalents of pre-commit/pre-commit-hooks
  *.go                         — one file per hook (RunXxx(args []string) int)
  *_test.go                    — AAA-pattern tests (one per hook file)
  util.go                      — shared helpers: openRepo, addedFiles, trackedFiles, …
  testutil_test.go             — writeTestFile, readTestFile test helpers

.pre-commit-config.yaml        — hooks used on this repo itself
.pre-commit-hooks.yaml         — hook manifest for remote consumers
```

---

## Key concepts

### Two hook layers

**Layer 1 — `internal/hooks/builtin.go` + `registry.go`**
Used when `go-precommit run` is invoked directly. The runner checks `hooks.Registry[hookID]` first; if found, it calls the Go function in-process (no subprocess). The `entry:` field in `.pre-commit-config.yaml` is for the Python `pre-commit` tool only — the runner ignores it for registered built-ins.

**Layer 2 — `upstream-hooks/`**
Standalone binaries that mirror the Python `pre-commit/pre-commit-hooks` API (`RunXxx(args []string) int`). Each is also registered via `.pre-commit-hooks.yaml` for remote use.

### Runner execution path (`internal/runner/runner.go`)

```
Run(files, hookFilter, stage)
  └── for each repo/hook in config:
        ├── filter by hookFilter and stage
        ├── hooks.Registry[id] found? → runBuiltin (in-process, fast)
        └── not found?              → runExternal (subprocess, language env)
```

Built-ins are **always preferred** over the config `entry:` — prevents infinite recursion when the config entry is `go-precommit run <hook-id> --files`.

### File type detection (`internal/hooks/filetypes.go`)

`DetectFileTypes(path)` returns tags like `text`, `binary`, `go`, `yaml`, `json`, `executable`, etc. based on file extension, content sniff (null byte = binary), and filesystem mode (executable bit).

---

## Adding a new built-in hook

1. Implement `func builtinXxx(ctx *HookContext, files []string, args []string) *Result` in `internal/hooks/builtin.go`
2. Register it in `internal/hooks/registry.go`
3. Add an entry to `.pre-commit-hooks.yaml`
4. Add an entry to `.pre-commit-config.yaml` (for this repo's own CI)
5. Add a matching `upstream-hooks/xxx.go` + `upstream-hooks/xxx_test.go` following AAA

---

## Testing

```bash
go test ./...                        # all tests
go test ./upstream-hooks/... -v      # verbose hook tests
go test ./upstream-hooks/... -run TestRunCheckYAML
```

All tests follow **Arrange-Act-Assert** — see `.claude/skills/test-architecture/SKILL.md`.

Hook tests that depend on git state (`no-commit-to-branch`, `forbid-new-submodules`, `check-case-conflict`, `destroyed-symlinks`) run against the live repository and use `t.Skip` when prerequisites are not met.

---

## Build & install

```bash
go build -o go-precommit .
sudo mv go-precommit /usr/local/bin/   # or wherever on PATH
```

The installed binary at `/usr/local/bin/go-precommit` must be refreshed manually after rebuilding. `go install ./...` writes to `~/go/bin/` which may differ from the binary on PATH.

---

## CI / pre-commit hooks on this repo

Hooks run on every commit (pre-commit stage): `trailing-whitespace`, `end-of-file-fixer`, `mixed-line-ending`, `check-yaml`, `check-json`, `check-toml`, `detect-private-key` (excludes `_test.go` and `internal/hooks/builtin.go`), `detect-aws-credentials`, `check-merge-conflict`, `check-added-large-files`, `no-commit-to-branch` (blocks `main`), `go vet`, `golangci-lint`.

On push (pre-push stage): `go build`.

---

## Important decisions

- **Built-ins always win over config `entry:`** — the runner used to check `hookCfg.Entry == ""` before using a built-in, which caused infinite recursion when the config had `entry: go-precommit run <hook-id> --files`. Fixed by unconditionally preferring the registry.
- **`detect-private-key` excludes `_test.go` and `internal/hooks/builtin.go`** — these files contain PEM header strings as Go regex literals / test fixtures and would otherwise trigger false positives.
