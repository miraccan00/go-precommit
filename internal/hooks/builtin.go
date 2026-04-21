package hooks

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/miraccan00/go-precommit/internal/gitutil"
	"gopkg.in/yaml.v3"
)

// ── trailing-whitespace ──────────────────────────────────────────────────────

func builtinTrailingWhitespace(_ *HookContext, files []string, _ []string) *Result {
	var modified []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		changed := false
		for i, line := range lines {
			trimmed := strings.TrimRight(line, " \t")
			if trimmed != line {
				lines[i] = trimmed
				changed = true
			}
		}

		if changed {
			if err := os.WriteFile(f, []byte(strings.Join(lines, "\n")), 0o644); err == nil {
				modified = append(modified, f)
			}
		}
	}

	if len(modified) > 0 {
		return &Result{
			Pass:     false,
			Output:   "files were modified by this hook",
			Modified: modified,
		}
	}
	return &Result{Pass: true}
}

// ── end-of-file-fixer ────────────────────────────────────────────────────────

func builtinEndOfFileFixer(_ *HookContext, files []string, _ []string) *Result {
	var modified []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		if len(data) == 0 {
			continue
		}

		// Strip all trailing newlines, then add exactly one.
		fixed := strings.TrimRight(string(data), "\n\r") + "\n"
		if fixed != string(data) {
			if err := os.WriteFile(f, []byte(fixed), 0o644); err == nil {
				modified = append(modified, f)
			}
		}
	}

	if len(modified) > 0 {
		return &Result{
			Pass:     false,
			Output:   "files were modified by this hook",
			Modified: modified,
		}
	}
	return &Result{Pass: true}
}

// ── check-yaml ───────────────────────────────────────────────────────────────

func builtinCheckYAML(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var v interface{}
		if err := yaml.Unmarshal(data, &v); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", f, err))
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── check-json ───────────────────────────────────────────────────────────────

func builtinCheckJSON(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var v interface{}
		if err := json.Unmarshal(data, &v); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", f, err))
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── check-toml ───────────────────────────────────────────────────────────────
// Basic TOML syntax check (key=value, sections, comments only).

func builtinCheckTOML(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	keyRe := regexp.MustCompile(`^\s*[a-zA-Z0-9_.\-"]+\s*=`)
	sectionRe := regexp.MustCompile(`^\s*\[`)
	commentRe := regexp.MustCompile(`^\s*#`)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || commentRe.MatchString(line) ||
				sectionRe.MatchString(line) || keyRe.MatchString(line) {
				continue
			}
			errs = append(errs, fmt.Sprintf("%s:%d: unexpected line: %s", f, lineNum, line))
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── check-merge-conflict ─────────────────────────────────────────────────────

func builtinCheckMergeConflict(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if strings.HasPrefix(line, "<<<<<<<") ||
				strings.HasPrefix(line, "=======") ||
				strings.HasPrefix(line, ">>>>>>>") {
				errs = append(errs, fmt.Sprintf("%s:%d: conflict marker", f, lineNum))
				break
			}
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── detect-private-key ───────────────────────────────────────────────────────

var privateKeyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`-----BEGIN\s+(RSA|DSA|EC|OPENSSH|PGP|PRIVATE)\s+PRIVATE KEY`),
	regexp.MustCompile(`-----BEGIN PRIVATE KEY-----`),
}

func builtinDetectPrivateKey(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, pat := range privateKeyPatterns {
			if pat.Match(data) {
				errs = append(errs, fmt.Sprintf("%s: possible private key", f))
				break
			}
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── check-added-large-files ──────────────────────────────────────────────────

func builtinCheckAddedLargeFiles(_ *HookContext, files []string, args []string) *Result {
	maxKB := 500
	for i, arg := range args {
		if arg == "--maxkb" && i+1 < len(args) {
			if v, err := strconv.Atoi(args[i+1]); err == nil {
				maxKB = v
			}
		}
	}

	maxBytes := int64(maxKB * 1024)
	var errs []string
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if info.Size() > maxBytes {
			errs = append(errs, fmt.Sprintf("%s: %d KB exceeds limit of %d KB",
				f, info.Size()/1024, maxKB))
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── mixed-line-ending ────────────────────────────────────────────────────────

func builtinMixedLineEnding(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		hasCRLF := bytes.Contains(data, []byte("\r\n"))
		// Remove all CRLF and check if bare LF remains.
		stripped := bytes.ReplaceAll(data, []byte("\r\n"), []byte(""))
		hasLF := bytes.Contains(stripped, []byte("\n"))
		if hasCRLF && hasLF {
			errs = append(errs, fmt.Sprintf("%s: mixed line endings (CRLF and LF)", f))
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── check-case-conflict ──────────────────────────────────────────────────────

func builtinCheckCaseConflict(_ *HookContext, files []string, _ []string) *Result {
	seen := make(map[string]string)
	var errs []string
	for _, f := range files {
		lower := strings.ToLower(f)
		if existing, ok := seen[lower]; ok {
			errs = append(errs, fmt.Sprintf("case conflict: %s vs %s", existing, f))
		} else {
			seen[lower] = f
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── check-symlinks ───────────────────────────────────────────────────────────

func builtinCheckSymlinks(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		info, err := os.Lstat(f)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if _, err := os.Stat(f); err != nil {
				errs = append(errs, fmt.Sprintf("%s: broken symlink", f))
			}
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── no-commit-to-branch ──────────────────────────────────────────────────────

func builtinNoCommitToBranch(ctx *HookContext, _ []string, args []string) *Result {
	protected := []string{"main", "master"}

	// Parse --branch flags from args.
	var custom []string
	for i, arg := range args {
		if arg == "--branch" && i+1 < len(args) {
			custom = append(custom, args[i+1])
		}
	}
	if len(custom) > 0 {
		protected = custom
	}

	branch, err := gitutil.CurrentBranch(ctx.RepoPath)
	if err != nil {
		// No HEAD yet (empty repo) — nothing to protect against.
		return &Result{Pass: true}
	}

	for _, b := range protected {
		if branch == b {
			return &Result{
				Pass:   false,
				Output: fmt.Sprintf("direct commit to branch %q is not allowed", branch),
			}
		}
	}
	return &Result{Pass: true}
}

// ── check-executables-have-shebangs ─────────────────────────────────────────

func builtinCheckExecutablesHaveShebangs(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if info.Mode()&0o111 == 0 {
			continue // not executable
		}

		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		buf := make([]byte, 2)
		n, _ := fh.Read(buf)
		_ = fh.Close()

		if n < 2 || string(buf[:2]) != "#!" {
			errs = append(errs, fmt.Sprintf("%s: executable without shebang", f))
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}
