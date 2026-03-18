# Phase 0: Flat File Format Specification

## Overview

Flat uses a custom file format (`.fmdx`) to store directory trees with complete metadata preservation. The format uses **context-specific section delimiters** to avoid conflicts with content that might contain delimiter-like strings.

## Format Specification

### Structure

```
!--~---~BEGIN-FLAT-FILE-MULTI~--~---!
mdx_block_hash: <sha256>
file_hash: <sha256>
content_type: <mime>
!--~---~END-HEADER~--~---!
path: "relative/path"
filename: "name.ext"
mode: "0644"
modified: "2026-03-16T12:00:00Z"
created: "2026-03-16T10:00:00Z"
symlink: ""
xattrs: {}
content_type: "text/plain"
is_external: false
external_path: ""
mdx_block_hash: ""
!--~---~END-METADATA~--~---!
base64-encoded-content
!--~---~END-FILE-CONTENT~--~---!
```

### Multi-File Example

```
!--~---~BEGIN-FLAT-FILE-MULTI~--~---!
mdx_block_hash: <sha256>
file_hash: <sha256>
content_type: <mime>
!--~---~END-HEADER~--~---!
path: "file1.md"
filename: "file1.md"
mode: "0644"
modified: "2026-03-16T12:00:00Z"
created: "2026-03-16T10:00:00Z"
symlink: ""
xattrs: {}
content_type: "text/markdown"
is_external: false
!--~---~END-METADATA~--~---!
UyBGaWxlIDEgLSBDb250ZW50
!--~---~END-FILE-CONTENT~--~---!
!--~---~BEGIN-FLAT-FILE-MULTI~--~---!
mdx_block_hash: <sha256>
file_hash: <sha256>
content_type: <mime>
!--~---~END-HEADER~--~---!
path: "file2.md"
filename: "file2.md"
mode: "0644"
modified: "2026-03-16T12:00:00Z"
created: "2026-03-16T10:00:00Z"
symlink: ""
xattrs: {}
content_type: "text/markdown"
is_external: false
!--~---~END-METADATA~--~---!
UyBGaWxlIDIgLSBDb250ZW50
!--~---~END-FILE-CONTENT~--~---!
```

### Delimiters

All delimiters follow the pattern `!--~---~...~--~---!`:

| Delimiter | Purpose |
|-----------|---------|
| `!--~---~BEGIN-FLAT-FILE-MULTI~--~---!` | Marks start of .fmdx file |
| `!--~---~END-HEADER~--~---!` | Marks end of header block (hashes and content type) |
| `!--~---~END-METADATA~--~---!` | Marks end of file metadata block |
| `!--~---~END-FILE-CONTENT~--~---!` | Marks end of file content, start of next entry |

### Header Block

Contains:
- `mdx_block_hash`: SHA-256 hash of the metadata block
- `file_hash`: SHA-256 hash of original file content
- `content_type`: Auto-detected MIME type

### Metadata Block

Contains YAML with:
- `path`: Relative path from source directory
- `filename`: Base filename only
- `mode`: Octal permissions (e.g., "0644")
- `modified`: Last modification time (ISO 8601)
- `created`: Creation time (if available)
- `symlink`: Empty or symlink target path
- `xattrs`: Extended attributes (user.* and security.*)
- `content_type`: Auto-detected MIME type
- `is_external`: True if external reference
- `external_path`: Original path (for external refs)
- `mdx_block_hash`: SHA-256 hash of metadata YAML block

### Content Block

- **Always base64 encoded** (prevents delimiter conflicts)
- Can contain any characters
- Decoded on unflatten

## Why Context-Specific Delimiters?

Using separate delimiters for each section boundary solves several problems:

1. **No delimiter-in-content conflicts** - Each delimiter is unique
2. **Content can contain any string** - No need for escaping
3. **Clear parsing logic** - Parser knows which delimiter to expect
4. **Future extensibility** - Easy to add new section types

## Metadata Fields

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
| `mdx_block_hash` | string | SHA-256 hash of metadata YAML block |

## Hash Algorithms

All hashes are computed and stored:

- **SHA-256** (32 bytes) - Primary verification algorithm
- **SHA-512** (64 bytes) - Extra security
- **MD5** (16 bytes) - Fast verification
- **BLAKE2** (32 bytes) - Modern alternative
- **CRC32** (4 bytes) - Quick error detection

Hashes are stored as hexadecimal strings.

## Content Encoding

All content is **always base64 encoded** to prevent:
- Delimiter conflicts in content
- Binary data corruption
- Encoding issues with special characters
- Whitespace preservation problems

This means:
- Text files are encoded (not stored as plain text)
- Binary files are encoded (as before)
- All content is safe for storage in .fmdx

## CLI Specification

### Default Mode

```bash
# In project directory
flat
```

**Behavior**:
- If `{cwd}.fmdx` does NOT exist: flatten current directory to `{cwd}.fmdx`
- If `{cwd}.fmdx` EXISTS: error and require explicit command

### Flatten Command

```bash
flat flatten <source-dir> <output.fmdx>
```

**Flags**:
- `-v, --verbose`: Print progress output
- `--no-bin`: Skip binary files
- `--external`: Store external references (path only)
- `--exclude <pattern>`: Exclude files matching pattern
- `--ignore-file <path>`: Path to .flatignore file

### Unflatten Command

```bash
flat unflatten <input.fmdx> <destination-dir>
```

**Flags**:
- `-v, --verbose`: Print progress output
- `--bypass-checksum`: Skip SHA-256 verification (not recommended)

### Version Command

```bash
flat version
```

Displays version information.

## .flatignore

Create a `.flatignore` file to exclude files:

```
# Comments start with #
*.bin
*.exe
node_modules/
.git/
.DS_Store
dist/
```

### Pattern Matching

- `*.ext` - Match extension
- `dir/` - Match directory
- `filename` - Exact filename match

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

# Copy and restore
scp project.fmdx user@destination:/tmp/
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
   - Text files (base64 encoded)
   - Binary files (base64 encoded)
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
    golang.org/x/crypto v0.17.0  # BLAKE2
    golang.org/x/text v0.14.0    # MIME detection
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
- Multi-file .fmdx (split large projects)
- Web UI for visualization
- Diff mode (show changes between .fmdx files)

## Notes

1. **SHA-256 is always verified** on unflatten unless `--bypass-checksum` is used
2. **All 5 hash types are computed** but only SHA-256 is required for verification
3. **External references** are stored as paths only (no content in .fmdx)
4. **Binary files** are always base64 encoded (as are all files now)
5. **Default mode** is `flat` alone - auto-flatten if no .fmdx exists
6. **No YAML wrapper lines** - Direct delimiter-to-delimiter structure

## Implementation Phases

### Phase 0: Specification ✅ [COMPLETE]
- [x] Format definition
- [x] CLI specification
- [x] Edge cases documentation
- [x] Context-specific delimiter format
- [x] No YAML wrapper lines

### Phase 1: Core Implementation ✅ [COMPLETE]
- [x] Go module setup
- [x] Hash computation (5 algorithms)
- [x] Format parser/writer
- [x] Base64 encoder (always encode)
- [x] Metadata collector
- [x] Binary detection (for --no-bin flag)
- [x] Ignore pattern matching

### Phase 2: Commands ✅ [COMPLETE]
- [x] Flatten command
- [x] Unflatten command
- [x] Version command
- [x] Checksum verification
- [x] Error handling

### Phase 3: Finalization ✅ [COMPLETE]
- [x] Testing suite
- [x] Documentation
- [x] Release preparation
- [x] Performance optimization
