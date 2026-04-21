package hooks

import (
	"os"
	"path/filepath"
	"strings"
)

// extensionTypes maps file extensions to their semantic type tags.
// Every text-based extension also carries the "text" tag.
var extensionTypes = map[string][]string{
	// Systems / compiled
	".go":    {"go", "text"},
	".py":    {"python", "text"},
	".rb":    {"ruby", "text"},
	".java":  {"java", "text"},
	".kt":    {"kotlin", "text"},
	".scala": {"scala", "text"},
	".cs":    {"csharp", "text"},
	".c":     {"c", "text"},
	".h":     {"c", "text"},
	".cpp":   {"c++", "cpp", "text"},
	".cc":    {"c++", "cpp", "text"},
	".cxx":   {"c++", "cpp", "text"},
	".hpp":   {"c++", "cpp", "text"},
	".rs":    {"rust", "text"},
	".swift": {"swift", "text"},
	".m":     {"objective-c", "text"},

	// Scripting
	".js":   {"javascript", "text"},
	".jsx":  {"javascript", "text"},
	".ts":   {"typescript", "javascript", "text"},
	".tsx":  {"typescript", "javascript", "text"},
	".sh":   {"shell", "bash", "text"},
	".bash": {"shell", "bash", "text"},
	".zsh":  {"shell", "zsh", "text"},
	".fish": {"shell", "fish", "text"},
	".ps1":  {"powershell", "text"},
	".lua":  {"lua", "text"},
	".r":    {"r", "text"},
	".R":    {"r", "text"},
	".pl":   {"perl", "text"},
	".php":  {"php", "text"},

	// Data / config formats
	".yaml": {"yaml", "text"},
	".yml":  {"yaml", "text"},
	".json": {"json", "text"},
	".toml": {"toml", "text"},
	".xml":  {"xml", "text"},
	".ini":  {"ini", "text"},
	".cfg":  {"ini", "text"},
	".conf": {"text"},
	".env":  {"text"},
	".csv":  {"csv", "text"},
	".tsv":  {"tsv", "text"},

	// Docs / markup
	".md":       {"markdown", "text"},
	".markdown": {"markdown", "text"},
	".rst":      {"rst", "text"},
	".txt":      {"text"},
	".html":     {"html", "text"},
	".htm":      {"html", "text"},
	".css":      {"css", "text"},
	".scss":     {"scss", "css", "text"},
	".sass":     {"sass", "css", "text"},
	".less":     {"less", "css", "text"},
	".tex":      {"tex", "text"},

	// Infrastructure
	".tf":    {"terraform", "text"},
	".hcl":   {"hcl", "text"},
	".proto": {"proto", "text"},
	".sql":   {"sql", "text"},
	".graphql": {"graphql", "text"},
	".gql":   {"graphql", "text"},
	".Makefile": {"makefile", "text"},
	".mk":    {"makefile", "text"},
	".dockerfile": {"dockerfile", "text"},
}

// DetectFileTypes returns the set of type tags for the given file path.
func DetectFileTypes(path string) []string {
	ext := strings.ToLower(filepath.Ext(path))

	// Special case: files named exactly "Makefile", "Dockerfile", etc.
	base := strings.ToLower(filepath.Base(path))
	switch base {
	case "makefile", "gnumakefile":
		return withExecutableBit(path, []string{"makefile", "text"})
	case "dockerfile":
		return withExecutableBit(path, []string{"dockerfile", "text"})
	}

	if types, ok := extensionTypes[ext]; ok {
		return withExecutableBit(path, types)
	}

	return detectByContent(path)
}

func withExecutableBit(path string, base []string) []string {
	info, err := os.Stat(path)
	if err != nil {
		return append(base, "non_executable")
	}
	if info.Mode()&0o111 != 0 {
		return append(base, "executable")
	}
	return append(base, "non_executable")
}

// detectByContent sniffs the first 512 bytes to distinguish text from binary.
func detectByContent(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	buf = buf[:n]

	// Null byte → binary
	for _, b := range buf {
		if b == 0 {
			return withExecutableBit(path, []string{"binary"})
		}
	}
	return withExecutableBit(path, []string{"text"})
}

// HasType returns true when the file's detected types contain t.
func HasType(path, t string) bool {
	for _, ft := range DetectFileTypes(path) {
		if ft == t {
			return true
		}
	}
	return false
}
