# flat

A CLI tool to flatten directory trees into a single `.fmdx` file and unflatten them back.

## Features

- **Single File Format**: Store entire directory structure in one `.fmdx` file
- **Complete Metadata**: Preserves permissions, timestamps, symlinks, and extended attributes
- **Checksum Verification**: SHA-256 verification ensures data integrity
- **Binary Detection**: Auto-detect and optionally skip binary files
- **External References**: Reference files without embedding content
- **Pattern Exclusion**: Use `.flatignore` to exclude files/directories
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Build from Source

```bash
git clone https://github.com/yourusername/flat.git
cd flat
make build
```

Or with mage:
```bash
mage build
```

### Install to GOPATH

```bash
make install
# or
mage install
```

### Download

Download pre-built binaries from releases.

## Usage

### Quick Start

```bash
# In your project directory, run `flat` alone
flat

# Auto-flattens current directory to {project-name}.fmdx
```

### Flatten

```bash
# Flatten a directory
flat flatten ./src ./output.fmdx

# Skip binary files
flat flatten --no-bin ./project ./backup.fmdx

# External file references
flat flatten --external ./large-files ./refs.fmdx

# Verbose output
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

### Version

```bash
flat version
```

## Flags

### Flatten Command

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Print progress output |
| `--no-bin` | Skip binary files |
| `--external` | Store external references (path only) |
| `--exclude <pattern>` | Exclude files matching pattern (repeatable) |
| `--ignore-file <path>` | Path to .flatignore file (default: ".flatignore") |

### Unflatten Command

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Print progress output |
| `--bypass-checksum` | Skip SHA-256 verification |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `FLAT_VERBOSE=true` | Enable verbose mode by default |

## File Format

### Structure

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
modified: "2026-03-16T12:00:00Z"
created: "2026-03-16T10:00:00Z"
symlink: ""
xattrs:
  user.comment: "test"
content_type: "text/plain"
is_external: false
---
base64-encoded-content
---MDX---
```

### Delimiters

- **Header**: `---BEGIN-FLAT-FILE-MULTI---`
- **MDX Section**: `---MDX---`
- **YAML Wrapper**: `---`

### Metadata Fields

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Relative path from source |
| `filename` | string | Base filename |
| `mode` | string | Octal permissions (e.g., "0644") |
| `modified` | timestamp | Last modification time |
| `created` | timestamp | Creation time |
| `symlink` | string | Empty or symlink target |
| `xattrs` | map | Extended attributes |
| `content_type` | string | Auto-detected MIME type |
| `is_external` | bool | True if external reference |

### Checksum Algorithms

All 5 hash algorithms are computed:

- **SHA-256** (32 bytes) - Primary verification
- **SHA-512** (64 bytes) - Extra security
- **MD5** (16 bytes) - Fast verification
- **BLAKE2** (32 bytes) - Modern alternative
- **CRC32** (4 bytes) - Quick error detection

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

## Testing

```bash
go test ./...
```

## Edge Cases

### Empty Files
- Zero-byte files are handled correctly
- Content section is empty in .fmdx

### Binary Files
- Detected using magic bytes + extension
- Base64 encoded for safe storage
- Can be skipped with `--no-bin`

### Symlinks
- Symlink targets are stored (not dereferenced)
- Recreated during unflatten

### Extended Attributes
- User-defined attributes preserved
- May not restore on all systems

### Permission Errors
- Read errors: skip and warn
- Write errors: error and stop

## Architecture

```
flat/
├── main.go              # Entry point + default mode
├── cmd/
│   ├── flatten.go       # Flatten command
│   ├── unflatten.go     # Unflatten command
│   └── version.go       # Version command
├── config/              # Configuration (env vars)
├── format/              # Format parser/writer
├── hash/                # Hash computation
├── metadata/            # Metadata collection
├── encoder/             # Base64 encoding
└── docs/                # Documentation
```

## License

MIT License
