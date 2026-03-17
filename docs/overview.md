# Flat Tool - Complete Documentation

## Overview

Flat is a CLI tool built in Go that flattens directory trees into a single `.fmdx` file and can unflatten them back. It preserves all POSIX metadata and uses SHA-256 checksums for verification.

## Quick Reference

### Commands

```bash
flat                              # Auto-flatten current directory (if no .fmdx exists)
flat flatten <src> <output.fmdx>  # Flatten a directory
flat unflatten <input.fmdx> <dest> # Unflatten a file
flat version                      # Show version
```

### Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Verbose output |
| `--no-bin` | Skip binary files |
| `--external` | External file references |
| `--exclude <pattern>` | Exclude patterns |
| `--bypass-checksum` | Skip verification (unflatten only) |

### Environment

```bash
FLAT_VERBOSE=true  # Enable verbose mode
```

## Documentation Files

- **docs/phase-0.md** - Specification and format definition
- **docs/phase-1.md** - Core implementation details
- **docs/phase-2.md** - Command implementation
- **docs/phase-3.md** - Testing, documentation, release
- **README.md** - User guide

## Format Specification

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
---
base64-encoded-content
---MDX---
```

## Project Structure

```
flat/
├── main.go
├── cmd/
│   ├── flatten.go
│   ├── unflatten.go
│   └── version.go
├── config/
├── format/
│   ├── writer.go
│   ├── parser.go
│   ├── magic.go
│   ├── mime.go
│   └── ignore.go
├── hash/
├── metadata/
├── encoder/
├── .flatignore.example
├── go.mod
├── README.md
└── docs/
    ├── phase-0.md
    ├── phase-1.md
    ├── phase-2.md
    └── phase-3.md
```

## Implementation Phases

### Phase 0 - Specification
- Format definition
- CLI specification
- Edge cases documentation

### Phase 1 - Core Implementation
- Go module setup
- Hash computation (5 algorithms)
- Format parser/writer
- Base64 encoder
- Metadata collector

### Phase 2 - Commands
- Flatten command
- Unflatten command
- Version command
- Checksum verification
- Error handling

### Phase 3 - Finalization
- Testing suite
- Documentation
- Release preparation

## Key Features

1. **Auto-flatten mode**: `flat` alone auto-flattens if no .fmdx exists
2. **Checksum verification**: SHA-256 always verified on unflatten
3. **Binary detection**: Magic bytes + extension for accurate detection
4. **External references**: Store paths without embedding content
5. **Pattern exclusion**: .flatignore with glob patterns
6. **Complete metadata**: Permissions, timestamps, symlinks, xattrs

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration
- `gopkg.in/yaml.v3` - YAML marshaling
- `golang.org/x/crypto/blake2` - BLAKE2 hashing
- Standard library: sha256, sha512, md5, crc32, base64

## Testing Strategy

- Unit tests for individual functions
- Integration tests for commands
- Edge case tests (empty files, symlinks, permissions)
- Performance tests (large directories)

## Next Steps

1. Initialize Go module
2. Create directory structure
3. Implement Phase 1 modules
4. Implement Phase 2 commands
5. Write tests (Phase 3)
6. Build and release
