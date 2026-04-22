package hooks

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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

// ── check-shebang-scripts-are-executable ─────────────────────────────────────

func builtinCheckShebangScriptsAreExecutable(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil || len(data) < 2 {
			continue
		}
		if string(data[:2]) != "#!" {
			continue
		}
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if info.Mode()&0o111 == 0 {
			errs = append(errs, fmt.Sprintf("%s: has shebang but is not executable", f))
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── check-vcs-permalinks ──────────────────────────────────────────────────────
// Detects GitHub links that point to a branch (e.g. /blob/main/) instead of a
// commit hash permalink.

// vcsPermalinkRe captures the ref segment (between /blob/ and the next /).
// Matches are filtered in code: skip if the ref is already a commit hash.
var vcsPermalinkRe = regexp.MustCompile(
	`https://github\.com/[^/ ]+/[^/ ]+/blob/([^/. ]+)/[^# ]+#L\d+`,
)

func isVCSCommitHash(ref string) bool {
	if len(ref) < 4 || len(ref) > 64 {
		return false
	}
	for _, c := range ref {
		if (c < 'a' || c > 'f') && (c < 'A' || c > 'F') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

func builtinCheckVCSPermalinks(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for i, line := range strings.Split(string(data), "\n") {
			m := vcsPermalinkRe.FindStringSubmatch(line)
			if m == nil || isVCSCommitHash(m[1]) {
				continue
			}
			errs = append(errs, fmt.Sprintf("%s:%d: non-permanent VCS link", f, i+1))
		}
	}
	if len(errs) > 0 {
		return &Result{
			Pass: false,
			Output: strings.Join(errs, "\n") +
				"\nNon-permanent github link detected.\nOn any page on github press [y] to load a permalink.",
		}
	}
	return &Result{Pass: true}
}

// ── fix-byte-order-marker ─────────────────────────────────────────────────────

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func builtinFixByteOrderMarker(_ *HookContext, files []string, _ []string) *Result {
	var modified []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		if bytes.HasPrefix(data, utf8BOM) {
			if err := os.WriteFile(f, data[3:], 0o644); err == nil {
				modified = append(modified, f)
			}
		}
	}
	if len(modified) > 0 {
		return &Result{Pass: false, Output: "removed byte-order marker", Modified: modified}
	}
	return &Result{Pass: true}
}

// ── file-contents-sorter ──────────────────────────────────────────────────────

func builtinFileContentsSorter(_ *HookContext, files []string, args []string) *Result {
	ignoreCase := false
	unique := false
	for _, a := range args {
		switch a {
		case "--ignore-case":
			ignoreCase = true
		case "--unique":
			unique = true
		}
	}

	var modified []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		var lines []string
		for _, l := range strings.Split(strings.TrimRight(string(data), "\r\n"), "\n") {
			if strings.TrimSpace(l) != "" {
				lines = append(lines, strings.TrimRight(l, "\r"))
			}
		}

		if unique {
			seen := make(map[string]bool, len(lines))
			deduped := lines[:0]
			for _, l := range lines {
				if !seen[l] {
					seen[l] = true
					deduped = append(deduped, l)
				}
			}
			lines = deduped
		}

		sorted := make([]string, len(lines))
		copy(sorted, lines)
		sort.Slice(sorted, func(i, j int) bool {
			a, b := sorted[i], sorted[j]
			if ignoreCase {
				a, b = strings.ToLower(a), strings.ToLower(b)
			}
			return a < b
		})

		newContent := strings.Join(sorted, "\n") + "\n"
		if newContent != string(data) {
			if err := os.WriteFile(f, []byte(newContent), 0o644); err == nil {
				modified = append(modified, f)
			}
		}
	}
	if len(modified) > 0 {
		return &Result{Pass: false, Output: "files were sorted", Modified: modified}
	}
	return &Result{Pass: true}
}

// ── forbid-new-submodules ─────────────────────────────────────────────────────

func builtinForbidNewSubmodules(ctx *HookContext, _ []string, _ []string) *Result {
	cmd := exec.Command("git", "diff", "--diff-filter=A", "--raw", "--staged", "--")
	cmd.Dir = ctx.RepoPath
	out, err := cmd.Output()
	if err != nil {
		return &Result{Pass: true}
	}

	var newSubs []string
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		fields := strings.Fields(parts[0])
		if len(fields) >= 2 && fields[1] == "160000" {
			newSubs = append(newSubs, parts[1])
		}
	}
	if len(newSubs) > 0 {
		var sb strings.Builder
		for _, s := range newSubs {
			fmt.Fprintf(&sb, "%s: new submodule introduced\n", s)
		}
		sb.WriteString("\nThis commit introduces new submodules.\nDid you unintentionally `git add .`?\n")
		return &Result{Pass: false, Output: sb.String()}
	}
	return &Result{Pass: true}
}

// ── destroyed-symlinks ────────────────────────────────────────────────────────

func builtinDestroyedSymlinks(ctx *HookContext, files []string, _ []string) *Result {
	if len(files) == 0 {
		return &Result{Pass: true}
	}

	gitArgs := append([]string{"status", "--porcelain=v2", "-z", "--"}, files...)
	cmd := exec.Command("git", gitArgs...)
	cmd.Dir = ctx.RepoPath
	out, err := cmd.Output()
	if err != nil {
		return &Result{Pass: true}
	}

	var destroyed []string
	for _, entry := range strings.Split(string(out), "\x00") {
		fields := strings.Fields(entry)
		// porcelain v2 ordinary changed entry: "1 XY sub mH mI mW hH hI path"
		if len(fields) < 9 || fields[0] != "1" {
			continue
		}
		modeHEAD := fields[3]
		modeIndex := fields[4]
		if modeHEAD == "120000" && modeIndex != "120000" && modeIndex != "000000" {
			path := strings.Join(fields[8:], " ")
			destroyed = append(destroyed, path)
		}
	}
	if len(destroyed) > 0 {
		var sb strings.Builder
		sb.WriteString("Destroyed symlinks:\n")
		for _, d := range destroyed {
			fmt.Fprintf(&sb, "- %s\n", d)
		}
		return &Result{Pass: false, Output: sb.String()}
	}
	return &Result{Pass: true}
}

// ── detect-aws-credentials ────────────────────────────────────────────────────

func awsSecretsFromFile(credFile string) [][]byte {
	path := credFile
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	targetVars := map[string]bool{
		"aws_secret_access_key": true,
		"aws_security_token":    true,
		"aws_session_token":     true,
	}
	var secrets [][]byte
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx > 0 {
			key := strings.ToLower(strings.TrimSpace(line[:idx]))
			val := strings.TrimSpace(line[idx+1:])
			if targetVars[key] && val != "" {
				secrets = append(secrets, []byte(val))
			}
		}
	}
	return secrets
}

func builtinDetectAWSCredentials(_ *HookContext, files []string, args []string) *Result {
	allowMissing := false
	credFiles := []string{"~/.aws/config", "~/.aws/credentials", "/etc/boto.cfg", "~/.boto"}

	for i, a := range args {
		switch a {
		case "--allow-missing-credentials":
			allowMissing = true
		case "--credentials-file":
			if i+1 < len(args) {
				credFiles = append(credFiles, args[i+1])
			}
		}
	}

	var secrets [][]byte
	for _, cf := range credFiles {
		secrets = append(secrets, awsSecretsFromFile(cf)...)
	}
	for _, envVar := range []string{"AWS_SECRET_ACCESS_KEY", "AWS_SECURITY_TOKEN", "AWS_SESSION_TOKEN"} {
		if v := os.Getenv(envVar); v != "" {
			secrets = append(secrets, []byte(v))
		}
	}

	if len(secrets) == 0 {
		if allowMissing {
			return &Result{Pass: true}
		}
		return &Result{
			Pass:   false,
			Output: "No AWS keys were found in the configured credential files and environment variables.",
		}
	}

	var errs []string
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, secret := range secrets {
			if bytes.Contains(content, secret) {
				hidden := string(secret)
				if len(hidden) > 4 {
					hidden = hidden[:4] + strings.Repeat("*", len(hidden)-4)
				}
				errs = append(errs, fmt.Sprintf("AWS secret found in %s: %s", f, hidden))
				break
			}
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── pretty-format-json ────────────────────────────────────────────────────────

func builtinPrettyFormatJSON(_ *HookContext, files []string, args []string) *Result {
	indent := "    "
	autofix := false
	for i, a := range args {
		switch a {
		case "--autofix":
			autofix = true
		case "--indent":
			if i+1 < len(args) {
				indent = args[i+1]
			}
		}
	}

	var errs []string
	var modified []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var v interface{}
		if err := json.Unmarshal(data, &v); err != nil {
			errs = append(errs, fmt.Sprintf("%s: invalid JSON: %v", f, err))
			continue
		}
		pretty, err := json.MarshalIndent(v, "", indent)
		if err != nil {
			continue
		}
		pretty = append(pretty, '\n')
		if !bytes.Equal(pretty, data) {
			if autofix {
				if err := os.WriteFile(f, pretty, 0o644); err == nil {
					modified = append(modified, f)
				}
			} else {
				errs = append(errs, fmt.Sprintf("%s: not pretty-formatted", f))
			}
		}
	}
	if len(errs) > 0 || len(modified) > 0 {
		return &Result{
			Pass:     len(errs) == 0,
			Output:   strings.Join(errs, "\n"),
			Modified: modified,
		}
	}
	return &Result{Pass: true}
}

// ── check-xml ─────────────────────────────────────────────────────────────────

func builtinCheckXML(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		dec := xml.NewDecoder(bytes.NewReader(data))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s: Failed to xml parse (%v)", f, err))
				break
			}
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── check-illegal-windows-names ───────────────────────────────────────────────

var illegalWindowsNameRe = regexp.MustCompile(
	`(?i)(^|/)((CON|PRN|AUX|NUL|COM\d|LPT\d)(\.|/|$)|[<>:"\\|?*])`,
)

func builtinCheckIllegalWindowsNames(_ *HookContext, files []string, _ []string) *Result {
	var errs []string
	for _, f := range files {
		if illegalWindowsNameRe.MatchString(f) {
			errs = append(errs, fmt.Sprintf("%s: illegal Windows filename", f))
		}
	}
	if len(errs) > 0 {
		return &Result{Pass: false, Output: strings.Join(errs, "\n")}
	}
	return &Result{Pass: true}
}

// ── sort-simple-yaml ──────────────────────────────────────────────────────────
// Sorts top-level keys of a simple YAML file (root must be a plain mapping).

func builtinSortSimpleYAML(_ *HookContext, files []string, _ []string) *Result {
	var modified []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		var root yaml.Node
		if err := yaml.Unmarshal(data, &root); err != nil {
			continue
		}
		if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
			continue
		}
		mapping := root.Content[0]
		if mapping.Kind != yaml.MappingNode || len(mapping.Content)%2 != 0 {
			continue
		}

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

		alreadySorted := true
		for i, p := range pairs {
			if p.key.Value != sorted[i].key.Value {
				alreadySorted = false
				break
			}
		}
		if alreadySorted {
			continue
		}

		flat := make([]*yaml.Node, 0, len(mapping.Content))
		for _, p := range sorted {
			flat = append(flat, p.key, p.val)
		}
		mapping.Content = flat

		out, err := yaml.Marshal(&root)
		if err != nil {
			continue
		}
		if err := os.WriteFile(f, out, 0o644); err == nil {
			modified = append(modified, f)
		}
	}
	if len(modified) > 0 {
		return &Result{Pass: false, Output: "files were sorted", Modified: modified}
	}
	return &Result{Pass: true}
}
