# Phase 0: Flat Tool Specification

## Overview

Flat is a CLI tool that flattens directory trees into a single `.fmdx` file and can unflatten them back. The tool is designed for:

- **Backup**: Capture complete directory structure with all metadata
- **Transfer**: Move entire projects as a single file
- **Version Control**: Store project state in a portable format
- **External References**: Reference files without embedding content
- **Binary Detection**: Optional filtering of binary files

## Design Goals

1. **Simplicity**: Default mode is just `flat` - no arguments needed
2. **Safety**: SHA-256 checksums always verified on unflatten
3. **Completeness**: Preserve all POSIX metadata (permissions, timestamps, symlinks, xattrs)
4. **Portability**: Single file format that works across platforms
5. **Performance**: Efficient hashing and encoding

## Technology Stack

- **Language**: Go (single binary, cross-platform, fast I/O)
- **CLI Framework**: Cobra + Viper
- **Checksum Algorithms**: SHA-256 (default), SHA-512, MD5, BLAKE2, CRC32
- **Content Encoding**: Base64 for binary-safe storage
- **Metadata Format**: YAML (human-readable within .fmdx)

## File Format

### Structure

```
---BEGIN-FLAT-FILE-MULTI---
---
mdx_block_hash: <sha256 of YAML metadata>
file_hash: <sha256 of original content>
content_type: <auto-detected MIME type>
---
---MDX---
---
```yaml
path: "relative/path/to/file"
filename: "filename.ext"
mode: "0644"
modified: "2026-03-16T12:00:00Z"
created: "2026-03-16T10:00:00Z"
symlink: ""  # empty or contains target
xattrs:
  user.comment: "test"
is_external: false
external_path: ""
---
base64-encoded-file-content
---MDX---
---
mdx_block_hash: <sha256 of YAML metadata>
file_hash: <sha256 of original content>
content_type: "application/octet-stream"
---
---MDX---
---
```yaml
path: "another/file.bin"
...
---
base64-encoded-binary-content
---MDX---
```

### Delimiters

- **Header**: `---BEGIN-FLAT-FILE-MULTI---`
- **MDX section**: `---MDX---`
- **YAML wrapper**: `---` (on its own line)

### Metadata Fields

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Relative path from source directory |
| `filename` | string | Base filename only |
| `mode` | string | Octal permissions (e.g., "0644") |
| `modified` | timestamp | Last modification time (ISO 8601) |
| `created` | timestamp | Creation time (if available) |
| `symlink` | string | Empty or symlink target path |
| `xattrs` | map | Extended attributes (user.* and security.*) |
| `content_type` | string | Auto-detected MIME type |
| `is_external` | bool | True if external reference |
| `external_path` | string | Original path (for external refs) |

### Hash Algorithms

All hashes are computed and stored:

- **SHA-256** (32 bytes) - Primary verification algorithm
- **SHA-512** (64 bytes) - Extra security
- **MD5** (16 bytes) - Fast verification
- **BLAKE2** (32 bytes) - Modern alternative
- **CRC32** (4 bytes) - Quick error detection

Hashes are stored as hexadecimal strings.

## CLI Specification

### Default Mode

```bash
# In project directory
flat
```

**Behavior**:
- If `{cwd}.fmdx` does NOT exist: flatten current directory to `{cwd}.fmdx`
- If `{cwd}.fmdx` EXISTS: error and require explicit command

### Commands

#### `flat flatten <source-dir> <output.fmdx>`

Flatten a directory tree into a single `.fmdx` file.

**Flags**:
- `-v, --verbose`: Print progress output
- `--no-bin`: Skip binary files (warn what was skipped)
- `--external`: Store external references (path only, no content)
- `--exclude <pattern>`: Exclude files matching pattern (can be repeated)
- `--ignore-file <path>`: Path to `.flatignore` file (default: ".flatignore")

**Example**:
```bash
flat flatten ./src ./output.fmdx
flat flatten -v --no-bin ./project ./backup.fmdx
```

#### `flat unflatten <input.fmdx> <destination-dir>`

Unflatten a `.fmdx` file into a directory structure.

**Flags**:
- `-v, --verbose`: Print progress output
- `--bypass-checksum`: Skip SHA-256 verification (not recommended)

**Example**:
```bash
flat unflatten backup.fmdx ./restored
flat unflatten -v project.fmdx ./output
```

#### `flat version`

Show version information.

**Example**:
```bash
flat version
```

### Environment Variables

- `FLAT_VERBOSE=true`: Enable verbose mode by default

### Help Output

```bash
$ flat --help
flat - Flatten/unflatten directory trees

Usage:
  flat [command]

Available Commands:
  flatten     Flatten a directory tree into a .fmdx file
  unflatten   Unflatten a .fmdx file into a directory structure
  version     Show version information

Flags:
  -h, --help   help for flat

Use "flat [command] --help" for more information about a command.

$ flat flatten --help
flatten a directory tree into a .fmdx file

Usage:
  flat flatten <source-dir> <output.fmdx> [flags]

Flags:
  -v, --verbose          verbose output
      --no-bin           skip binary files
      --external         external file references
      --exclude strings  exclude patterns
      --ignore-file      ignore file path (default ".flatignore")
  -h, --help             help for flatten
```

## File Filtering

### Binary Detection

Files are classified as binary or text using a combined approach:

1. **Magic bytes** (primary): Check first bytes against known signatures
   - PNG: `89 50 4E 47`
   - JPEG: `FF D8 FF`
   - GIF: `47 49 46 38`
   - MP3: `ID3` or MPEG frame headers
   - ZIP: `50 4B 03 04`
   - ELF: `7F 45 4C 46`
   - And many more...

2. **File extension** (secondary): Cross-reference with common extensions
   - Binary: `.png`, `.jpg`, `.mov`, `.mp4`, `.exe`, `.dll`, `.bin`, etc.
   - Text: `.txt`, `.md`, `.go`, `.js`, `.py`, `.json`, `.yaml`, etc.

If `--no-bin` is specified, binary files are skipped entirely.

### Exclusion Patterns (`.flatignore`)

```
# Comments start with #
*.bin
*.exe
*.dll
node_modules/
.git/
.DS_Store
vendor/
dist/
```

Pattern matching supports:
- `*.ext` - Match extension
- `dir/` - Match directory
- `filename` - Exact filename match

## Edge Cases

### Empty Files

- Zero-byte files are valid
- Content section is empty string in base64
- Hashes are computed on empty content

### Binary Files

- Encoded as base64 for safe storage
- MIME type auto-detected
- `--no-bin` flag can skip them

### Symlinks

- Symlink target is stored (not dereferenced)
- During unflatten: recreate as symlink with correct target
- External references can be symlinks

### Extended Attributes (xattrs)

- User-defined attributes (`user.*`) are preserved
- Security attributes (`security.*`) are preserved
- Some systems may not support all xattrs (warn if not restored)

### Permission Errors

- If a file cannot be read: skip and warn
- If a file cannot be written: error and stop

### Special Characters in Filenames

- Base64 encoding handles any characters
- Path separators stored as forward slashes
- Unicode characters handled correctly

## Error Handling

### During Flatten

1. **Directory not found**: Error and exit
2. **Permission denied**: Skip file, warn, continue
3. **I/O error**: Skip file, warn, continue
4. **Invalid .flatignore**: Warn but continue with default patterns

### During Unflatten

1. **Invalid .fmdx format**: Error and exit
2. **Hash mismatch**: Error and exit (unless `--bypass-checksum`)
3. **Directory not found**: Create parent directories
4. **Permission error**: Skip file, warn, continue
5. **External file missing**: No validation (user's responsibility)

### Default Mode

1. **No arguments**: Check for `{cwd}.fmdx`
   - Missing: auto-flatten
   - Exists: error and require explicit command

## Directory Structure

```
flat/
├── main.go              # Entry point + default mode logic
├── cmd/
│   ├── flatten.go       # flat flatten <src> <output>
│   ├── unflatten.go     # flat unflatten <input> <dest>
│   └── version.go       # flat version
├── config/
│   └── config.go        # Config struct + env var parsing
├── format/
│   ├── writer.go        # Write .fmdx file format
│   ├── parser.go        # Parse .fmdx file format
│   ├── magic.go         # Binary detection (magic bytes + extension)
│   ├── mime.go          # Auto-detect MIME types
│   └── ignore.go        # .flatignore pattern matching
├── hash/
│   └── hash.go          # SHA-256, SHA-512, MD5, BLAKE2, CRC32
├── metadata/
│   └── collector.go     # Collect POSIX metadata
├── encoder/
│   └── base64.go        # Base64 encode/decode
├── .flatignore.example  # Example ignore file
├── go.mod               # Go module
├── go.sum               # Dependencies
├── README.md            # Documentation
└── docs/
    ├── phase-0.md       # This file
    ├── phase-1.md       # Phase 1 implementation
    └── phase-2.md       # Phase 2 implementation
```

## Testing Strategy

### Test Cases

1. **Basic flatten/unflatten**
   - Single file
   - Directory with files
   - Nested directory structure

2. **Metadata preservation**
   - File permissions (0644, 0755, etc.)
   - Timestamps (modified, created)
   - Symlinks (relative, absolute)
   - Extended attributes

3. **Binary handling**
   - Text files (no change)
   - Binary files (base64 encode/decode)
   - Mixed content

4. **Edge cases**
   - Empty files
   - Files with special characters
   - Very large files
   - Files with null bytes

5. **Checksum verification**
   - Valid hashes (success)
   - Tampered file (hash mismatch)
   - --bypass-checksum flag

6. **Filtering**
   - --no-bin flag (skip binaries)
   - .flatignore patterns
   - --external flag (external refs)

7. **Error handling**
   - Permission errors
   - Missing directories
   - Invalid .fmdx format

### Test Data

Create test directories with:
- Regular text files (`.txt`, `.md`, `.go`, `.js`)
- Binary files (`.png`, `.jpg`, `.bin`, `.exe`)
- Symlinks (pointing to files and directories)
- Files with different permissions
- Files with xattrs set
- Nested directories
- Empty files
- Files with special characters in names

## Dependencies

### Go Modules

```go
module flat

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    golang.org/x/crypto v0.17.0  // BLAKE2
    golang.org/x/text v0.14.0    // MIME detection
)
```

### External Libraries

- **Cobra**: CLI framework
- **Viper**: Configuration management
- **crypto/sha256**: SHA-256 hashing (stdlib)
- **crypto/sha512**: SHA-512 hashing (stdlib)
- **crypto/md5**: MD5 hashing (stdlib)
- **golang.org/x/crypto/blake2**: BLAKE2 hashing
- **hash/crc32**: CRC32 hashing (stdlib)
- **encoding/base64**: Base64 encoding (stdlib)
- **gopkg.in/yaml.v3**: YAML marshaling

## Versioning

Version format: `0.x.y`

- **0.x.y**: Pre-1.0 releases (breaking changes possible)
- Version displayed by `flat version`
- Auto-extracted from git tags if available

## Future Enhancements (Post-1.0)

- Compression (gzip/zstd) for smaller .fmdx files
- Incremental backups (compare with previous .fmdx)
- Encryption (optional AES-256 encryption)
- Compression of base64 content
- Multi-file .fmdx (split large projects)
- Web UI for visualization
- Diff mode (show changes between .fmdx files)

## Notes

1. **SHA-256 is always verified** on unflatten unless `--bypass-checksum` is used
2. **All 5 hash types are computed** but only SHA-256 is required for verification
3. **External references** are stored as paths only (no content in .fmdx)
4. **Binary files** can be skipped with `--no-bin` flag
5. **Default mode** is `flat` alone - auto-flatten if no .fmdx exists
