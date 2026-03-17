package format

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewIgnoreParser(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with non-existent file (should return parser with no patterns)
	parser, err := NewIgnoreParser(filepath.Join(tmpDir, "nonexistent"))
	if err != nil {
		t.Fatalf("NewIgnoreParser should not error for non-existent file: %v", err)
	}
	if len(parser.GetPatterns()) != 0 {
		t.Errorf("Parser should have no patterns for non-existent file, got %d", len(parser.GetPatterns()))
	}

	// Test with valid file
	ignoreFile := filepath.Join(tmpDir, ".flatignore")
	content := `# Comment
*.bin
*.exe

node_modules/
`
	err = os.WriteFile(ignoreFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}

	parser, err = NewIgnoreParser(ignoreFile)
	if err != nil {
		t.Fatalf("NewIgnoreParser should not error for valid file: %v", err)
	}

	patterns := parser.GetPatterns()
	if len(patterns) != 3 {
		t.Errorf("Parser should have 3 patterns, got %d: %v", len(patterns), patterns)
	}
}

func TestShouldIgnore_ExactMatch(t *testing.T) {
	parser := &IgnoreParser{
		patterns: []string{"README.md", "LICENSE", "Makefile"},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"README.md", true},
		{"LICENSE", true},
		{"Makefile", true},
		{"main.go", false},
		{"test.go", false},
	}

	for _, tt := range tests {
		if result := parser.ShouldIgnore(tt.path); result != tt.expected {
			t.Errorf("ShouldIgnore(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func TestShouldIgnore_ExtensionPattern(t *testing.T) {
	parser := &IgnoreParser{
		patterns: []string{"*.bin", "*.exe", "*.so"},
	}

	tests := []struct {
		filename string
		expected bool
	}{
		{"data.bin", true},
		{"program.exe", true},
		{"library.so", true},
		{"file.so.1", false}, // .so.1 doesn't match .so pattern
		{"main.go", false},
		{"readme.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if result := parser.ShouldIgnore(tt.filename); result != tt.expected {
				t.Errorf("ShouldIgnore(%q) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestShouldIgnore_DirectoryPattern(t *testing.T) {
	parser := &IgnoreParser{
		patterns: []string{"node_modules/", ".git/", "vendor/"},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"node_modules/package.js", true},
		{"node_modules/", true},
		{".git/config", true},
		{".git/", true},
		{"vendor/lib.go", true},
		{"src/main.go", false},
		{"test/test.go", false},
	}

	for _, tt := range tests {
		if result := parser.ShouldIgnore(tt.path); result != tt.expected {
			t.Errorf("ShouldIgnore(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func TestShouldIgnore_GlobPattern(t *testing.T) {
	parser := &IgnoreParser{
		patterns: []string{"test*", "*test.go", "*.bak"},
	}

	tests := []struct {
		filename string
		expected bool
	}{
		{"test_main.go", true},
		{"test.go", true},
		{"main_test.go", true},
		{"backup.bak", true},
		{"main.go", false},
		{"config.yaml", false},
	}

	for _, tt := range tests {
		if result := parser.ShouldIgnore(tt.filename); result != tt.expected {
			t.Errorf("ShouldIgnore(%q) = %v, expected %v", tt.filename, result, tt.expected)
		}
	}
}

func TestShouldIgnore_MultiplePatterns(t *testing.T) {
	parser := &IgnoreParser{
		patterns: []string{"*.bin", "node_modules/", "test_*", ".git/"},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"data.bin", true},
		{"node_modules/package.js", true},
		{"test_main.go", true},
		{".git/config", true},
		{"main.go", false},
		{"src/utils.go", false},
	}

	for _, tt := range tests {
		if result := parser.ShouldIgnore(tt.path); result != tt.expected {
			t.Errorf("ShouldIgnore(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func TestShouldIgnore_CommentsAndEmptyLines(t *testing.T) {
	// Simulate parsing a file with comments and empty lines
	parser := &IgnoreParser{
		patterns: []string{"*.bin", "*.exe"}, // Comments and empty lines are filtered
	}

	if parser.ShouldIgnore("data.bin") {
		t.Log("Binary file correctly ignored")
	} else {
		t.Error("Binary file should be ignored")
	}
}

func TestIgnoreParser_AddPattern(t *testing.T) {
	parser := &IgnoreParser{
		patterns: []string{"*.bin"},
	}

	parser.AddPattern("*.exe")
	parser.AddPattern("node_modules/")

	patterns := parser.GetPatterns()
	if len(patterns) != 3 {
		t.Errorf("Should have 3 patterns, got %d: %v", len(patterns), patterns)
	}
}

func TestShouldIgnore_PathWithSubdirectories(t *testing.T) {
	parser := &IgnoreParser{
		patterns: []string{"node_modules/", "dist/"},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"node_modules/package/index.js", true},
		{"src/node_modules/package.js", false}, // Should not match - not at root
		{"dist/bundle.js", true},
		{"build/dist/bundle.js", false},       // Should not match - not at root
		{"my_node_modules/package.js", false}, // Should not match - different name
	}

	for _, tt := range tests {
		if result := parser.ShouldIgnore(tt.path); result != tt.expected {
			t.Errorf("ShouldIgnore(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func TestIgnoreParser_EmptyPatterns(t *testing.T) {
	parser := &IgnoreParser{}

	// Should not ignore anything with empty patterns
	tests := []string{"file.txt", "data.bin", "node_modules/package.js"}
	for _, path := range tests {
		if parser.ShouldIgnore(path) {
			t.Errorf("ShouldIgnore(%q) should be false with empty patterns", path)
		}
	}
}

func TestNewIgnoreParser_FileWithOnlyComments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with only comments
	ignoreFile := filepath.Join(tmpDir, ".flatignore")
	content := `# This is a comment
# Another comment

`
	err := os.WriteFile(ignoreFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}

	parser, err := NewIgnoreParser(ignoreFile)
	if err != nil {
		t.Fatalf("NewIgnoreParser should not error: %v", err)
	}

	if len(parser.GetPatterns()) != 0 {
		t.Errorf("Parser should have no patterns for comment-only file, got %d", len(parser.GetPatterns()))
	}
}
