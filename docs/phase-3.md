# Phase 3: Testing, Documentation & Release

## Overview

Phase 3 focuses on finalizing the project with:

1. **Comprehensive Testing**: Test suite for all functionality
2. **Documentation**: README.md, usage examples, edge cases
3. **Edge Case Handling**: Ensure all scenarios work correctly
4. **Performance Optimization**: Handle large directories efficiently
5. **Release**: Build binary, create release notes

## Testing Strategy

### Test Categories

1. **Unit Tests**: Individual functions
2. **Integration Tests**: Command execution
3. **Edge Case Tests**: Special scenarios
4. **Performance Tests**: Large directories

### Test Directory Structure

```
flat/
├── test/
│   ├── data/                    # Test data fixtures
│   │   ├── text/
│   │   │   ├── hello.txt
│   │   │   └── readme.md
│   │   ├── binary/
│   │   │   ├── image.png
│   │   │   └── audio.mp3
│   │   ├── symlinks/
│   │   │   ├── link.txt -> ../text/hello.txt
│   │   │   └── linkdir -> ../text/
│   │   ├── special/
│   │   │   ├── "file with spaces.txt"
│   │   │   ├── unicode_文件.txt
│   │   │   └── empty.txt
│   │   ├── permissions/
│   │   │   ├── readonly.txt (0444)
│   │   │   └── executable.sh (0755)
│   │   └── xattrs/
│   │       └── tagged.txt (with user.comment)
│   ├── flatten_test.go          # Flatten tests
│   ├── unflatten_test.go        # Unflatten tests
│   ├── format_test.go           # Format parser/writer tests
│   ├── hash_test.go             # Hash computation tests
│   ├── binary_test.go           # Binary detection tests
│   └── ignore_test.go           # .flatignore tests
└── test.go                      # Test setup/teardown
```

### Test Cases

#### 1. Basic Flatten/Unflatten

```go
func TestBasicFlattenUnflatten(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    destDir := filepath.Join(tmpDir, "dest")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    os.MkdirAll(sourceDir, 0755)
    os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("hello"), 0644)
    
    // Test flatten
    err := executeFlatten(&config.Config{Verbose: true}, sourceDir, flatFile)
    assert.NoError(t, err)
    
    // Test unflatten
    err = executeUnflatten(&config.Config{Verbose: true}, flatFile, destDir)
    assert.NoError(t, err)
    
    // Verify
    content, _ := os.ReadFile(filepath.Join(destDir, "file.txt"))
    assert.Equal(t, "hello", string(content))
}
```

#### 2. Binary Detection

```go
func TestBinaryDetection(t *testing.T) {
    tests := []struct {
        name     string
        filepath string
        isBinary bool
    }{
        {"PNG file", "test/data/binary/image.png", true},
        {"JPEG file", "test/data/binary/photo.jpg", true},
        {"Text file", "test/data/text/hello.txt", false},
        {"Go file", "test/data/text/main.go", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            isBin, _ := format.IsBinary(tt.filepath)
            assert.Equal(t, tt.isBinary, isBin)
        })
    }
}
```

#### 3. Binary Files with --no-bin

```go
func TestNoBinFlag(t *testing.T) {
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    os.MkdirAll(sourceDir, 0755)
    os.WriteFile(filepath.Join(sourceDir, "text.txt"), []byte("hello"), 0644)
    os.WriteFile(filepath.Join(sourceDir, "image.png"), []byte{0x89, 0x50, 0x4E, 0x47}, 0644)
    
    // Flatten with --no-bin
    err := executeFlatten(&config.Config{NoBin: true, Verbose: true}, sourceDir, flatFile)
    assert.NoError(t, err)
    
    // Read .fmdx and verify only text file is included
    content, _ := os.ReadFile(flatFile)
    assert.Contains(t, string(content), "text.txt")
    assert.NotContains(t, string(content), "image.png")
}
```

#### 4. External References

```go
func TestExternalReferences(t *testing.T) {
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    os.MkdirAll(sourceDir, 0755)
    os.WriteFile(filepath.Join(sourceDir, "large.bin"), make([]byte, 1000000), 0644)
    
    // Flatten with --external
    err := executeFlatten(&config.Config{External: true, Verbose: true}, sourceDir, flatFile)
    assert.NoError(t, err)
    
    // Verify external reference is stored
    entries, _ := format.ParseAllEntries(flatFile)
    assert.Equal(t, true, entries[0].Metadata.IsExternal)
    assert.Empty(t, entries[0].Content) // No content
}
```

#### 5. .flatignore Patterns

```go
func TestIgnorePatterns(t *testing.T) {
    parser, _ := format.NewIgnoreParser("test/data/.flatignore")
    
    tests := []struct {
        path     string
        shouldIgnore bool
    }{
        {"node_modules/package.js", true},
        {"src/main.go", false},
        {".git/config", true},
        {"README.md", false},
    }
    
    for _, tt := range tests {
        assert.Equal(t, tt.shouldIgnore, parser.ShouldIgnore(tt.path))
    }
}
```

#### 6. Symlink Handling

```go
func TestSymlinkHandling(t *testing.T) {
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    destDir := filepath.Join(tmpDir, "dest")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    os.MkdirAll(sourceDir, 0755)
    os.WriteFile(filepath.Join(sourceDir, "target.txt"), []byte("target"), 0644)
    os.Symlink("target.txt", filepath.Join(sourceDir, "link.txt"))
    
    // Flatten
    err := executeFlatten(&config.Config{}, sourceDir, flatFile)
    assert.NoError(t, err)
    
    // Unflatten
    err = executeUnflatten(&config.Config{}, flatFile, destDir)
    assert.NoError(t, err)
    
    // Verify symlink
    linkDest := filepath.Join(destDir, "link.txt")
    info, _ := os.Lstat(linkDest)
    assert.True(t, info.Mode()&os.ModeSymlink != 0)
    
    target, _ := os.Readlink(linkDest)
    assert.Equal(t, "target.txt", target)
}
```

#### 7. Permission Preservation

```go
func TestPermissionPreservation(t *testing.T) {
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    destDir := filepath.Join(tmpDir, "dest")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    os.MkdirAll(sourceDir, 0755)
    
    // Create file with different permissions
    files := []struct {
        name string
        mode os.FileMode
    }{
        {"readonly.txt", 0444},
        {"executable.sh", 0755},
        {"normal.txt", 0644},
    }
    
    for _, f := range files {
        os.WriteFile(filepath.Join(sourceDir, f.name), []byte("content"), f.mode)
    }
    
    // Flatten and unflatten
    executeFlatten(&config.Config{}, sourceDir, flatFile)
    executeUnflatten(&config.Config{}, flatFile, destDir)
    
    // Verify permissions
    for _, f := range files {
        info, _ := os.Stat(filepath.Join(destDir, f.name))
        assert.Equal(t, f.mode, info.Mode().Perm())
    }
}
```

#### 8. Checksum Verification

```go
func TestChecksumVerification(t *testing.T) {
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    destDir := filepath.Join(tmpDir, "dest")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    os.MkdirAll(sourceDir, 0755)
    os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("hello"), 0644)
    
    // Flatten
    executeFlatten(&config.Config{}, sourceDir, flatFile)
    
    // Tamper with file content
    os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("tampered"), 0644)
    
    // Try to unflatten without bypass (should fail)
    err := executeUnflatten(&config.Config{}, flatFile, destDir)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "checksum mismatch")
    
    // Unflatten with bypass (should succeed)
    err = executeUnflatten(&config.Config{BypassChecksum: true}, flatFile, destDir)
    assert.NoError(t, err)
}
```

#### 9. Empty Files

```go
func TestEmptyFiles(t *testing.T) {
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    os.MkdirAll(sourceDir, 0755)
    os.WriteFile(filepath.Join(sourceDir, "empty.txt"), []byte{}, 0644)
    
    // Flatten
    err := executeFlatten(&config.Config{}, sourceDir, flatFile)
    assert.NoError(t, err)
    
    // Unflatten
    destDir := filepath.Join(tmpDir, "dest")
    err = executeUnflatten(&config.Config{}, flatFile, destDir)
    assert.NoError(t, err)
    
    // Verify empty file
    content, _ := os.ReadFile(filepath.Join(destDir, "empty.txt"))
    assert.Equal(t, 0, len(content))
}
```

#### 10. Special Characters in Filenames

```go
func TestSpecialCharacters(t *testing.T) {
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    destDir := filepath.Join(tmpDir, "dest")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    os.MkdirAll(sourceDir, 0755)
    
    // Files with special characters
    files := []string{
        "file with spaces.txt",
        "file-with-dashes.txt",
        "file_with_underscores.txt",
        "文件 with unicode.txt",
        "file\"with\"quotes.txt",
    }
    
    for _, f := range files {
        os.WriteFile(filepath.Join(sourceDir, f), []byte("content"), 0644)
    }
    
    // Flatten
    err := executeFlatten(&config.Config{}, sourceDir, flatFile)
    assert.NoError(t, err)
    
    // Unflatten
    err = executeUnflatten(&config.Config{}, flatFile, destDir)
    assert.NoError(t, err)
    
    // Verify all files
    for _, f := range files {
        path := filepath.Join(destDir, f)
        _, err := os.Stat(path)
        assert.NoError(t, err, "File %s not restored", f)
    }
}
```

## Performance Tests

### Large Directory Test

```go
func TestLargeDirectory(t *testing.T) {
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    // Create 1000 files
    os.MkdirAll(sourceDir, 0755)
    for i := 0; i < 1000; i++ {
        os.WriteFile(filepath.Join(sourceDir, fmt.Sprintf("file%d.txt", i)), 
            []byte("content"), 0644)
    }
    
    // Time flatten
    start := time.Now()
    err := executeFlatten(&config.Config{}, sourceDir, flatFile)
    elapsed := time.Since(start)
    
    assert.NoError(t, err)
    t.Logf("Flattened 1000 files in %s", elapsed)
    
    // Time unflatten
    destDir := filepath.Join(tmpDir, "dest")
    start = time.Now()
    err = executeUnflatten(&config.Config{}, flatFile, destDir)
    elapsed = time.Since(start)
    
    assert.NoError(t, err)
    t.Logf("Unflattened in %s", elapsed)
}
```

### Memory Usage Test

```go
func TestMemoryUsage(t *testing.T) {
    tmpDir := t.TempDir()
    sourceDir := filepath.Join(tmpDir, "source")
    flatFile := filepath.Join(tmpDir, "test.fmdx")
    
    // Create directory with large files
    os.MkdirAll(sourceDir, 0755)
    os.WriteFile(filepath.Join(sourceDir, "large.bin"), make([]byte, 100*1024*1024), 0644)
    
    // Flatten
    startMem := getMemStats()
    err := executeFlatten(&config.Config{}, sourceDir, flatFile)
    endMem := getMemStats()
    
    assert.NoError(t, err)
    t.Logf("Memory used: %d MB", endMem.Alloc-startMem.Alloc)
}
```

## Documentation

### README.md Structure

```markdown
# Flat - Directory Tree Flattening Tool

## Overview

Flat is a CLI tool that flattens directory trees into a single `.fmdx` file and can unflatten them back. It preserves all POSIX metadata including permissions, timestamps, symlinks, and extended attributes.

## Features

- ✅ Single file format (`.fmdx`)
- ✅ Complete metadata preservation
- ✅ SHA-256 checksum verification
- ✅ Binary file detection
- ✅ External file references
- ✅ .flatignore pattern support
- ✅ Cross-platform compatibility

## Installation

### Build from Source

```bash
git clone https://github.com/yourusername/flat.git
cd flat
go build -o flat
```

### Download Binary

Download pre-built binary from releases.

## Usage

### Quick Start

```bash
# In your project directory
flat

# This auto-flattens the current directory to {project-name}.fmdx
```

### Flatten

```bash
# Flatten a directory
flat flatten ./src ./output.fmdx

# Skip binary files
flat flatten --no-bin ./project ./backup.fmdx

# External references
flat flatten --external ./large-files ./refs.fmdx

# With verbose output
flat flatten -v ./project ./output.fmdx
```

### Unflatten

```bash
# Unflatten a .fmdx file
flat unflatten backup.fmdx ./restored

# Skip checksum verification (not recommended)
flat unflatten --bypass-checksum input.fmdx ./dest

# Verbose output
flat unflatten -v project.fmdx ./output
```

### Flags

#### Flatten Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Print progress output |
| `--no-bin` | Skip binary files |
| `--external` | Store external references |
| `--exclude <pattern>` | Exclude files matching pattern |
| `--ignore-file <path>` | Path to .flatignore file |

#### Unflatten Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Print progress output |
| `--bypass-checksum` | Skip SHA-256 verification |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `FLAT_VERBOSE=true` | Enable verbose mode |

### File Format

The `.fmdx` format uses a structured layout:

```
---BEGIN-FLAT-FILE-MULTI---
---
mdx_block_hash: <sha256>
file_hash: <sha256>
content_type: <mime>
---
---MDX---
---
```yaml
path: "relative/path"
filename: "name.ext"
mode: "0644"
...
```
---
base64-content
---MDX---
```

### .flatignore

Create a `.flatignore` file to exclude files:

```
# Comments start with #
*.bin
*.exe
node_modules/
.git/
.DS_Store
```

## Examples

### Backup Your Project

```bash
# Create backup
flat flatten ./my-project ./my-project-backup.fmdx

# Later, restore it
flat unflatten my-project-backup.fmdx ./restored
```

### Transfer Project

```bash
# On source machine
cd /path/to/project
flat flatten . ./project.fmdx

# Copy project.fmdx to destination
scp project.fmdx user@destination:/tmp/

# On destination machine
flat unflatten /tmp/project.fmdx ./project
```

### Exclude Large Files

```bash
# Create .flatignore
echo "large_files/" > .flatignore
echo "*.bin" >> .flatignore

# Flatten with exclusions
flat flatten ./project ./backup.fmdx
```

## Testing

```bash
go test ./...
```

## Contributing

Contributions welcome! Please read the contributing guidelines.

## License

MIT License
```

### Edge Cases Documentation

```markdown
## Edge Cases

### Empty Files

- Zero-byte files are handled correctly
- Content section is empty in .fmdx
- Hashes computed on empty content

### Binary Files

- Detected using magic bytes + extension
- Base64 encoded for safe storage
- --no-bin flag can skip them

### Symlinks

- Symlink targets are stored (not dereferenced)
- Recreated during unflatten
- External symlinks supported

### Extended Attributes

- User-defined attributes (user.*) preserved
- Security attributes (security.*) preserved
- Some systems may not support all xattrs

### Permission Errors

- Read errors: skip and warn
- Write errors: error and stop

### Special Characters

- Base64 encoding handles any characters
- Unicode filenames supported
- Paths stored as forward slashes

### Large Files

- No size limit
- Memory-efficient streaming
- Recommended for files < 1GB
```

## Release Checklist

- [ ] All tests passing
- [ ] README.md updated
- [ ] Version number updated
- [ ] Binary built for all platforms (linux, darwin, windows)
- [ ] Release notes created
- [ ] Tags created (v0.1.0, etc.)
- [ ] GitHub release created
- [ ] Download links tested

## Release Notes Template

```markdown
# v0.1.0 - Initial Release

## Features

- ✅ Flatten directory trees to .fmdx format
- ✅ Unflatten .fmdx back to directory structure
- ✅ Complete POSIX metadata preservation
- ✅ SHA-256 checksum verification
- ✅ Binary file detection and filtering
- ✅ External file references
- ✅ .flatignore pattern support
- ✅ Verbose output mode

## Format Specification

- Header: `---BEGIN-FLAT-FILE-MULTI---`
- MDX sections: `---MDX---`
- YAML metadata with hashes
- Base64 encoded content

## Known Issues

- Extended attributes may not restore on all systems

## Thanks

To all contributors and testers!
```

## Final Implementation Steps

1. **Code Implementation** (Phase 1 & 2)
   - [ ] Initialize Go module
   - [ ] Implement all modules
   - [ ] Integrate Cobra CLI
   - [ ] Add error handling

2. **Testing** (Phase 3)
   - [ ] Create test data fixtures
   - [ ] Write unit tests
   - [ ] Write integration tests
   - [ ] Run all tests
   - [ ] Fix any failures

3. **Documentation** (Phase 3)
   - [ ] Write README.md
   - [ ] Add usage examples
   - [ ] Document edge cases
   - [ ] Create .flatignore.example

4. **Release** (Phase 3)
   - [ ] Update version in code
   - [ ] Build binaries for all platforms
   - [ ] Create release notes
   - [ ] Tag release
   - [ ] Create GitHub release

5. **Cleanup**
   - [ ] Code review
   - [ ] Remove debug code
   - [ ] Optimize performance
   - [ ] Final testing
