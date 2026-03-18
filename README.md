# flat

A CLI tool to flatten directory trees into a single `.fmdx` file and unflatten them back. Built with Cobra and Viper.

## Features

- **Single File Format**: Store entire directory structure in one `.fmdx` file
- **Complete Metadata**: Preserves permissions, timestamps, symlinks, and extended attributes
- **Exact Content Preservation**: Text files maintain exact trailing newline state
- **Checksum Verification**: SHA-256 verification ensures data integrity
- **Binary Detection**: Auto-detect and optionally skip binary files
- **External References**: Reference files without embedding content
- **Pattern Exclusion**: Use `.flatignore` to exclude files/directories
- **Auto-Ignore Patterns**: Automatically ignores `*.fmdx` and `.agents.yaml` files
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Build from Source

```bash
git clone https://github.com/AgentForgeEngine/flat.git
cd flat
go build -o flat ./cmd/cli
```

### Install Globally

```bash
go install github.com/AgentForgeEngine/flat/cmd/cli@latest
```

### Download

Download pre-built binaries from releases.

## Usage

### Quick Start

```bash
# Show help
flat --help

# Flatten a directory
flat flatten ./src ./output.fmdx

# Unflatten a .fmdx file
flat unflatten output.fmdx ./restored
```

### Flatten

```bash
# Basic flatten
flat flatten ./project ./backup.fmdx

# Skip binary files
flat flatten --no-bin ./project ./backup.fmdx

# External file references (store paths only)
flat flatten --external ./large-files ./refs.fmdx

# Exclude patterns
flat flatten --exclude "*.log" --exclude "temp/" ./project ./backup.fmdx

# Verbose output
flat flatten -v ./project ./output.fmdx

# Custom ignore file
flat flatten --ignore-file .customignore ./project ./backup.fmdx
```

### Unflatten

```bash
# Basic unflatten
flat unflatten backup.fmdx ./restored

# Skip checksum verification (not recommended)
flat unflatten --bypass-checksum input.fmdx ./dest

# Verbose output
flat unflatten -v project.fmdx ./output

# Just restore directory metadata (AGENTS.yaml files)
flat unflatten --just-agents project.fmdx ./dest
```

### Version

```bash
flat version
```

## Flags

### Global Flags

| Flag | Description |
|------|-------------|
| `-h, --help` | Help for any command |
| `-v, --verbose` | Enable verbose output |
| `--ignore-file <path>` | Path to ignore file (default: ".flatignore") |
| `--version` | Show version information |

### Flatten Command Flags

| Flag | Description |
|------|-------------|
| `--no-bin` | Skip binary files |
| `--external` | Store external references (path only) |
| `--exclude <patterns>` | Exclude files matching pattern (repeatable) |
| `--just-agents` | Only clean up `.agents.yaml` files |

### Unflatten Command Flags

| Flag | Description |
|------|-------------|
| `--bypass-checksum` | Skip SHA-256 verification |
| `--just-agents` | Only restore directory metadata |

## File Format

### Structure

```
!--~---~BEGIN-FLAT-FILE-MULTI~--~---!
platform_os: "linux"
platform_arch: "amd64"
platform_hostname: "host"
platform_uid: 1001
platform_gid: 1001
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
end_with_newline: true
!--~---~END-METADATA~--~---!
base64-encoded-content
!--~---~END-FILE-CONTENT~--~---!
```

### Newline Handling

Text files preserve exact trailing newline state:
- Files without trailing newlines: `end_with_newline: false`
- Files with trailing newlines: `end_with_newline: true`
- During flatten: Text files get trailing newline added if missing
- During unflatten: Trailing newlines removed/added based on `end_with_newline`

### Delimiters

| Delimiter | Purpose |
|-----------|---------|
| `!--~---~BEGIN-FLAT-FILE-MULTI~--~---!` | Marks start of .fmdx file |
| `!--~---~END-HEADER~--~---!` | Marks end of header block |
| `!--~---~END-METADATA~--~---!` | Marks end of metadata block |
| `!--~---~END-FILE-CONTENT~--~---!` | Marks end of file content |

### Metadata Fields

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Relative path from source |
| `filename` | string | Base filename |
| `mode` | string | File permissions |
| `modified` | timestamp | Last modification time |
| `created` | timestamp | Creation time |
| `symlink` | string | Symlink target (empty for regular files) |
| `xattrs` | map | Extended attributes |
| `content_type` | string | Auto-detected MIME type |
| `is_external` | bool | True if external reference |
| `end_with_newline` | bool | Whether text file ends with newline |
| `uid` | int | User ID |
| `gid` | int | Group ID |

### Checksum Algorithms

All 5 hash algorithms are computed and stored:
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
large_files/
```

### Auto-Ignored Patterns

These patterns are automatically ignored (cannot be overridden):
- `*.fmdx` - Prevents recursive flattening
- `.agents.yaml` - Directory metadata files

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

# Verify integrity (default)
flat unflatten my-project-backup.fmdx ./verify
```

### Exclude Large Files

```bash
# Create .flatignore
cat > .flatignore << 'EOF'
# Large files
large_files/
*.bin
*.iso

# Build artifacts
dist/
build/
node_modules/
EOF

# Flatten with exclusions
flat flatten ./project ./backup.fmdx
```

### External References (for large files)

```bash
# Store paths instead of embedding content
flat flatten --external ./project ./refs.fmdx

# Unflatten creates references to original locations
flat unflatten refs.fmdx ./dest
```

### Transfer Project

```bash
# On source machine
cd /path/to/project
flat flatten . ./project.fmdx

# Copy and restore
scp project.fmdx user@destination:/tmp/
ssh user@destination "flat unflatten /tmp/project.fmdx /path/to/project"
```

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v ./format
```

## Architecture

```
flat/
├── cmd/
│   ├── cli/              # CLI entry point
│   ├── root.go           # Cobra root command
│   ├── flatten.go        # Flatten command
│   └── unflatten.go      # Unflatten command
├── config/               # Configuration (Viper)
├── format/               # Format parser/writer
├── hash/                 # Hash computation
├── metadata/             # Metadata collection
├── encoder/              # Base64 encoding
└── version/              # Version info
```

## Dependencies

- **Cobra**: CLI framework
- **Viper**: Configuration management
- **YAML v3**: Metadata encoding

## License

MIT License
