# /test-architecture

Review and enforce the **Arrange-Act-Assert (AAA)** pattern across all Go test files in this repository.
This skill is for **code review and architectural compliance** — it does not run tests itself.

---

## AAA Structure — Required in Every Test

Every `t.Run(...)` block must have three clearly separated phases, each marked with a comment:

```go
t.Run("file with RSA private key header returns 1", func(t *testing.T) {
    // Arrange
    dir := t.TempDir()
    path := writeTestFile(t, dir, "key.pem", []byte("-----BEGIN RSA PRIVATE KEY-----\n"))

    // Act
    got := RunDetectPrivateKey([]string{path})

    // Assert
    if got != 1 {
        t.Errorf("got %d, want 1", got)
    }
})
```

When Arrange and Act collapse into a single expression, combine them:

```go
t.Run("no files returns 0", func(t *testing.T) {
    // Arrange + Act
    got := RunDetectPrivateKey(nil)

    // Assert
    if got != 0 {
        t.Errorf("got %d, want 0", got)
    }
})
```

---

## Phase rules

| Phase | Must do | Must NOT do |
|---|---|---|
| **Arrange** | Create temp dirs/files, set up inputs, configure flags | Call the function under test |
| **Act** | Call exactly one function — the subject under test | Assert anything |
| **Assert** | Check return values and side effects | Set up more inputs |

---

## Test helper: `writeTestFile`

All file-based tests use the shared helper from `upstream-hooks/testutil_test.go`:

```go
func writeTestFile(t *testing.T, dir, name string, content []byte) string
```

Never call `os.WriteFile` directly in a test — use `writeTestFile` so failures surface via `t.Fatal`.

---

## Table-driven tests

For functions with many input variants, use a table-driven structure. Each row is still AAA — Arrange lives in the table definition, Act and Assert in the loop:

```go
tests := []struct {
    name    string
    input   string   // Arrange data
    want    int
}{
    {"empty input", "", 0},
    {"invalid YAML", "key: :\tbad", 1},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Arrange: inputs defined above

        // Act
        got := RunCheckYAML([]string{tt.input})

        // Assert
        if got != tt.want {
            t.Errorf("got %d, want %d", got, tt.want)
        }
    })
}
```

---

## Git-dependent tests

Hooks that read git state (`no-commit-to-branch`, `forbid-new-submodules`, `check-case-conflict`, `destroyed-symlinks`) run against the **real repository**. Tests for these must use `t.Skip` when prerequisites cannot be guaranteed:

```go
repo, err := openRepo()
if err != nil {
    t.Skip("not inside a git repository")
}
head, err := repo.Head()
if err != nil || !head.Name().IsBranch() {
    t.Skip("HEAD is detached or cannot be resolved")
}
```

Never assume a specific branch name — resolve it from `repo.Head()` at test time.

---

## Coverage checklist

Every hook file in `upstream-hooks/` must have a matching `_test.go` file. Currently covered:

- `check_added_large_files_test.go`
- `check_case_conflict_test.go`
- `check_executables_have_shebangs_test.go`
- `check_json_test.go`
- `check_merge_conflict_test.go`
- `check_shebang_scripts_are_executable_test.go`
- `check_symlinks_test.go`
- `check_toml_test.go`
- `check_vcs_permalinks_test.go`
- `check_xml_test.go`
- `check_yaml_test.go`
- `destroyed_symlinks_test.go`
- `detect_aws_credentials_test.go`
- `detect_private_key_test.go`
- `end_of_file_fixer_test.go`
- `file_contents_sorter_test.go`
- `fix_byte_order_marker_test.go`
- `forbid_new_submodules_test.go`
- `mixed_line_ending_test.go`
- `no_commit_to_branch_test.go`
- `pretty_format_json_test.go`
- `sort_simple_yaml_test.go`
- `trailing_whitespace_fixer_test.go`
- `util_test.go`

When adding a new hook, add its test file before merging.
