# go-precommit

A fast, dependency-free alternative to [pre-commit](https://pre-commit.com/) written in Go.

## The Problem

The Python `pre-commit` tool is widely adopted but comes with friction:

- **Requires Python and pip** — every developer machine and CI runner needs a working Python environment
- **Slow cold starts** — virtualenv creation and package installation on first run can take 30–60 seconds
- **Runtime overhead** — even simple checks like trailing whitespace spawn Python subprocesses
- **CI bloat** — Docker images or CI pipelines need Python installed just to run pre-commit hooks

`go-precommit` solves this by shipping the most common hooks as native Go code inside a single static binary. No Python, no Node, no external `git` binary required for the majority of use cases.

## Key Features

### Zero-dependency built-in hooks (pure Go, no subprocess)

All hooks below run entirely in-process. When a config references one of these IDs — regardless of the `repo:` URL — `go-precommit` uses the Go implementation directly, skipping any Python virtualenv.

#### File hygiene

| Hook ID | What it does |
|---|---|
| `trailing-whitespace` | Strips trailing spaces and tabs; **auto-fixes files** |
| `end-of-file-fixer` | Ensures every file ends with exactly one newline; **auto-fixes files** |
| `fix-byte-order-marker` | Removes the UTF-8 BOM (`\xEF\xBB\xBF`); **auto-fixes files** |
| `mixed-line-ending` | Detects or normalises mixed CRLF/LF line endings; **auto-fixes files** |
| `file-contents-sorter` | Sorts lines in target files alphabetically (`--ignore-case`, `--unique`) |

#### Syntax checks

| Hook ID | What it does |
|---|---|
| `check-yaml` | Validates YAML syntax (`--allow-multiple-documents`) |
| `check-json` | Validates JSON syntax |
| `check-toml` | Validates TOML syntax |
| `check-xml` | Validates XML syntax |
| `pretty-format-json` | Enforces consistent JSON formatting (`--autofix`, `--indent`, `--top-keys`) |
| `sort-simple-yaml` | Sorts top-level keys of simple YAML mappings; **auto-fixes files** |

#### Security

| Hook ID | What it does |
|---|---|
| `detect-private-key` | Blocks PEM private keys (RSA, EC, OPENSSH, PGP, PuTTY, OpenVPN …) |
| `detect-aws-credentials` | Detects AWS secret keys from your local credential files (`--allow-missing-credentials`) |

#### Git hygiene

| Hook ID | What it does |
|---|---|
| `check-merge-conflict` | Detects unresolved merge conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`) |
| `check-added-large-files` | Rejects files exceeding a size threshold (default 500 KB, `--maxkb`) |
| `no-commit-to-branch` | Blocks direct commits to protected branches (`--branch`, `--pattern`) |
| `forbid-new-submodules` | Prevents addition of new git submodules |
| `destroyed-symlinks` | Detects symlinks that were accidentally replaced by plain files |

#### Portability & filesystem

| Hook ID | What it does |
|---|---|
| `check-case-conflict` | Finds filename conflicts that would break case-insensitive filesystems (macOS, Windows) |
| `check-illegal-windows-names` | Detects filenames illegal on Windows (`CON`, `PRN`, `AUX`, reserved characters) |
| `check-symlinks` | Detects broken symbolic links |
| `check-executables-have-shebangs` | Ensures executable files start with `#!` |
| `check-shebang-scripts-are-executable` | Ensures files with `#!` have the executable bit set |
| `check-vcs-permalinks` | Detects non-permanent GitHub blob links (branch-based instead of commit-hash-based) |

### Use go-precommit as a remote repo

Reference go-precommit directly in any project's `.pre-commit-config.yaml`. The binary is downloaded and cached automatically — no Go toolchain required on the target machine:

```yaml
repos:
  - repo: https://github.com/miraccan00/go-precommit
    rev: v1.0.0   # pin to a release tag
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: detect-private-key
      - id: no-commit-to-branch
        args: [--branch, main, --branch, develop]
      - id: check-added-large-files
        args: [--maxkb, "1000"]
```

The `.pre-commit-hooks.yaml` file in this repository is the **hook manifest** — it declares every available hook ID, its binary entry point, accepted file types, and flags. pre-commit (and go-precommit itself) reads this file when the repo is used remotely.

### Compatible `.pre-commit-config.yaml` format

Uses the same config format as the original Python tool. Existing configs that reference `pre-commit/pre-commit-hooks` work without modification:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
        args: [--maxkb, "1000"]

  - repo: local
    hooks:
      - id: no-commit-to-branch
        args: [--branch, main, --branch, develop]
      - id: golangci-lint
        name: Go Lint
        entry: golangci-lint run
        language: system
        types: [go]
        pass_filenames: false
```

### Remote repo support with local caching

Remote repositories are cloned once and cached under `~/.cache/go-precommit/repos/`. Subsequent runs use the cache — no network access needed. Supports:
- Tag references (`v1.2.3`)
- Branch references
- Full commit hash checkout (fallback)

### Language environment support

For hooks that genuinely need a runtime (Python linters, Node formatters, etc.), `go-precommit` automatically sets up isolated environments:

| Language | Mechanism |
|---|---|
| `python` / `python3` | `python3 -m venv` + `pip install .` |
| `golang` / `go` | `go install ./...` with isolated `GOBIN` |
| `node` / `nodejs` | `npm install` + symlinked `node_modules` |
| `system` / `script` | Uses tools already on `PATH` — no setup |

Environments are cached under `~/.cache/go-precommit/envs/` keyed by a SHA-256 hash of `(repoURL + rev + additionalDependencies)`. They are created once and reused.

### System hooks — run any command

For tools already installed on your system (linters, formatters, scanners):

```yaml
- repo: local
  hooks:
    - id: gofmt
      name: gofmt
      entry: gofmt -l -w
      language: system
      types: [go]

    - id: eslint
      name: ESLint
      entry: npx eslint --fix
      language: system
      types: [javascript, typescript]
```

### Smart file filtering

Every hook supports the full pre-commit filter spec:

- `types` — file must match **all** listed types (AND logic)
- `types_or` — file must match **at least one** type (OR logic)
- `exclude_types` — skip files matching these types
- `files` — regex: only process matching paths
- `exclude` — regex: skip matching paths
- Top-level `exclude` — global regex applied before every hook

File type detection covers 30+ extensions and includes binary detection (null-byte sniff), executable bit detection, and common format tagging (`text`, `go`, `python`, `yaml`, `json`, `toml`, `markdown`, `shell`, `javascript`, `typescript`, `rust`, `java`, `xml`, `html`, `css`, `executable`, `non_executable`, `binary`, etc.).

### Git integration via go-git

Uses [go-git](https://github.com/go-git/go-git) for all git operations — no `git` binary required:

- Staged files only (default — what will be committed)
- All tracked files (`--all-files`)
- Specific files (`--files a.go,b.go`)

### Fail-fast mode

Add `fail_fast: true` at the top of your config to stop after the first failing hook.

### Single static binary + Docker

Compiles to a single self-contained binary (~8 MB). No installation prerequisites on the target machine.

A minimal Docker image (~16 MB) is included for CI use:

```bash
docker run --rm \
  -v "$(pwd)":/workspace \
  -v "$HOME/.cache/go-precommit":/root/.cache/go-precommit \
  go-precommit run
```

## Installation

### Build from source

```bash
git clone https://github.com/miraccan00/go-precommit
cd go-precommit
go build -o go-precommit .
sudo mv go-precommit /usr/local/bin/
```

### Docker

```bash
docker build -t go-precommit .
```

## Usage

### Install the git hook

Run once per repository to wire up the pre-commit hook:

```bash
go-precommit install
```

This writes `.git/hooks/pre-commit` pointing to the current binary. If a non-`go-precommit` hook already exists at that path, the install is aborted safely.

```bash
# Install a different hook type
go-precommit install --hook-type pre-push
```

### Run hooks manually

```bash
# Against staged files (same as what the git hook does)
go-precommit run

# Against all tracked files
go-precommit run --all-files

# Only specific hooks
go-precommit run trailing-whitespace check-yaml

# Specific files
go-precommit run --files main.go,cmd/root.go

# Show output even for passing hooks
go-precommit run --verbose

# Use a different config file
go-precommit run --config path/to/config.yaml
```

## Configuration reference

```yaml
# .pre-commit-config.yaml

# Global regex — skip matching files for every hook
exclude: "^vendor/|^generated/"

# Stop after the first failing hook
fail_fast: false

repos:
  - repo: local          # or a GitHub URL
    # rev: v1.0.0        # required for remote repos
    hooks:
      - id: trailing-whitespace

        # Optional overrides (all fields are optional):
        name: My hook name
        entry: my-command --flag         # override the command
        language: system                 # python | golang | node | system | script
        types: [python, text]            # AND filter
        types_or: [javascript, typescript]  # OR filter
        exclude_types: [binary]
        files: "src/.*\\.go$"           # regex include
        exclude: "_test\\.go$"          # regex exclude
        args: [--fix, --strict]
        pass_filenames: true            # default true
        always_run: false
        additional_dependencies: ["types-pyyaml"]
```

## How it works

```
git commit
    └── .git/hooks/pre-commit
            └── go-precommit run
                    ├── reads .pre-commit-config.yaml
                    ├── resolves staged files via go-git
                    └── for each hook:
                            ├── builtin? → run Go implementation directly
                            ├── remote repo? → clone/cache → read .pre-commit-hooks.yaml
                            │                               → setup language env (once)
                            │                               → exec subprocess
                            └── local system hook? → exec subprocess
```

## Comparison

| | go-precommit | pre-commit (Python) |
|---|---|---|
| Runtime required | None (static binary) | Python 3.x + pip |
| Built-in hooks | 25 (pure Go) | 50 (Python) |
| Common hook speed | ~10ms (in-process) | ~500ms (subprocess) |
| First-run setup | Instant | 30–60s (virtualenv) |
| External git binary | Not required | Required |
| Config format | `.pre-commit-config.yaml` ✓ | `.pre-commit-config.yaml` ✓ |
| Remote repos | ✓ | ✓ |
| Python/Node hooks | ✓ (env auto-setup) | ✓ |
| Docker image size | ~16 MB | ~200 MB+ |

## Contributing

### Development hooks (`.pre-commit-config.yaml`)

This repository uses go-precommit on itself. After cloning, install the git hooks:

```bash
go build -o go-precommit .
./go-precommit install          # pre-commit hook
./go-precommit install --hook-type pre-push
```

The hooks that run on every commit/push:

| Stage | Hook | What it checks |
|---|---|---|
| pre-commit | `trailing-whitespace` | Strips trailing spaces/tabs |
| pre-commit | `end-of-file-fixer` | Ensures single trailing newline |
| pre-commit | `mixed-line-ending` | Detects mixed CRLF/LF |
| pre-commit | `check-yaml` | YAML syntax |
| pre-commit | `check-json` | JSON syntax |
| pre-commit | `check-toml` | TOML syntax |
| pre-commit | `detect-private-key` | No private keys in source |
| pre-commit | `detect-aws-credentials` | No AWS secrets in source |
| pre-commit | `check-merge-conflict` | No unresolved conflict markers |
| pre-commit | `check-added-large-files` | Files under 500 KB |
| pre-commit | `no-commit-to-branch` | Blocks direct commits to `main` |
| pre-commit | `go vet` | Go static analysis |
| pre-commit | `golangci-lint` | Go linting |
| pre-push | `go build` | Project compiles cleanly |

### `.pre-commit-hooks.yaml`

This file is the **hook manifest** for remote use. It must stay in the repository root. Every hook listed here becomes available when someone points their config at `repo: https://github.com/miraccan00/go-precommit`. Adding a new built-in hook requires a matching entry here.

## License

MIT
