package format

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// IgnoreParser handles .flatignore file parsing
type IgnoreParser struct {
	patterns []string
}

// NewIgnoreParser creates a new ignore parser
func NewIgnoreParser(ignorePath string) (*IgnoreParser, error) {
	parser := &IgnoreParser{
		patterns: make([]string, 0),
	}

	f, err := os.Open(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return parser, nil // No ignore file, no patterns
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parser.patterns = append(parser.patterns, line)
	}

	return parser, scanner.Err()
}

// ShouldIgnore checks if a path should be ignored
func (p *IgnoreParser) ShouldIgnore(relPath string) bool {
	filename := filepath.Base(relPath)

	for _, pattern := range p.patterns {
		if matchesPattern(pattern, relPath, filename) {
			return true
		}
	}

	return false
}

// matchesPattern checks if a pattern matches a path
func matchesPattern(pattern, relPath, filename string) bool {
	// Directory pattern (ends with /)
	if strings.HasSuffix(pattern, "/") {
		dirPattern := strings.TrimSuffix(pattern, "/")
		return strings.HasPrefix(relPath, dirPattern+"/") || relPath == dirPattern
	}

	// Extension pattern (starts with *)
	if strings.HasPrefix(pattern, "*.") {
		ext := strings.TrimPrefix(pattern, "*.") // This is "bin" for "*.bin"
		return strings.HasSuffix(filename, "."+ext)
	}

	// Glob pattern (contains *)
	if strings.Contains(pattern, "*") {
		// Try matching against filename first
		match, err := filepath.Match(pattern, filename)
		if err == nil && match {
			return true
		}
		// Also try matching against full path
		match, err = filepath.Match(pattern, relPath)
		if err == nil && match {
			return true
		}
		// Try matching pattern against filename with wildcard prefix
		// This handles cases like "*.bak" matching "backup.bak"
		if strings.HasPrefix(pattern, "*") && len(pattern) > 1 {
			suffix := pattern[1:]
			if strings.HasSuffix(filename, suffix) {
				return true
			}
		}
		// Try matching pattern against filename with wildcard suffix
		// This handles cases like "test*" matching "test_main.go"
		if strings.HasSuffix(pattern, "*") && len(pattern) > 1 {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(filename, prefix) {
				return true
			}
		}
		return false
	}

	// Exact filename match
	return filename == pattern || relPath == pattern
}

// AddPattern adds a pattern to the parser
func (p *IgnoreParser) AddPattern(pattern string) {
	p.patterns = append(p.patterns, pattern)
}

// GetPatterns returns all patterns
func (p *IgnoreParser) GetPatterns() []string {
	return p.patterns
}
