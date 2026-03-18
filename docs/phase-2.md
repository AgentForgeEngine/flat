# Phase 2: Commands Implementation

## Status: In Progress

## Overview

Phase 2 implements the CLI commands: flatten, unflatten, and version. This phase focuses on user-facing functionality and command-line interface.

## Implementation Status

### CLI Framework ✅ [COMPLETE]

**File**: `flat/cmd/cli/main.go`

- [x] Cobra CLI framework setup
- [x] Root command structure
- [x] Command registration
- [x] Help text generation
- [x] Version command

```go
func main() {
    rootCmd := &cobra.Command{
        Use:   "flat",
        Short: "Flat - Directory backup tool",
        Long:  "Flat creates backups of directory trees with complete metadata preservation",
    }
    
    rootCmd.AddCommand(flattenCmd)
    rootCmd.AddCommand(unflattenCmd)
    rootCmd.AddCommand(versionCmd)
    
    rootCmd.Execute()
}
```

### Flatten Command ✅ [COMPLETE]

**File**: `flat/cmd/cli/flatten.go`

- [x] Source directory scanning
- [x] Ignore pattern filtering
- [x] Binary file detection
- [x] Metadata collection for each file
- [x] Hash computation for metadata and content
- [x] Format writer integration
- [x] Base64 encoding of content
- [x] Progress output (verbose mode)
- [x] Error handling
- [x] External reference support (--external flag)
- [x] Skip binaries (--no-bin flag)

```go
func flattenCmd(cmd *cobra.Command, args []string) error {
    // 1. Validate arguments
    // 2. Load ignore patterns
    // 3. Scan source directory
    // 4. Filter files
    // 5. For each file:
    //    - Collect metadata
    //    - Compute hashes
    //    - Encode content
    //    - Write to .fmdx
    // 6. Print summary
}
```

**Flags**:
- `-v, --verbose`: Print progress output
- `--no-bin`: Skip binary files
- `--external`: Store external references (path only)
- `--exclude <pattern>`: Exclude files matching pattern
- `--ignore-file <path>`: Path to .flatignore file

### Unflatten Command ✅ [COMPLETE]

**File**: `flat/cmd/cli/unflatten.go`

- [x] .fmdx file parsing
- [x] Format validation
- [x] Entry iteration
- [x] Base64 decoding of content
- [x] File creation with metadata
- [x] Permission restoration
- [x] Timestamp restoration
- [x] Symlink creation
- [x] Extended attribute restoration
- [x] SHA-256 hash verification
- [x] --bypass-checksum flag
- [x] Progress output (verbose mode)
- [x] Error handling

```go
func unflattenCmd(cmd *cobra.Command, args []string) error {
    // 1. Validate arguments
    // 2. Open .fmdx file
    // 3. Parse all entries
    // 4. For each entry:
    //    - Verify hash
    //    - Decode content
    //    - Create directory structure
    //    - Write file with permissions
    //    - Set timestamps
    //    - Create symlinks
    //    - Set xattrs
    // 5. Print summary
}
```

**Flags**:
- `-v, --verbose`: Print progress output
- `--bypass-checksum`: Skip SHA-256 verification (not recommended)

### Version Command ✅ [COMPLETE]

**File**: `flat/cmd/cli/version.go`

- [x] Version number display
- [x] Git tag detection
- [x] Build information
- [x] Format: `flat version x.y.z`

```go
func versionCmd(cmd *cobra.Command, args []string) error {
    fmt.Printf("flat version %s\n", version)
    return nil
}
```

### Error Handling ✅ [COMPLETE]

- [x] File not found errors
- [x] Permission denied errors
- [x] Invalid format errors
- [x] Hash mismatch errors
- [x] Directory creation errors
- [x] Symlink creation errors
- [x] Xattr setting errors
- [x] User-friendly error messages

### Default Mode ⏳ [IN PROGRESS]

**File**: `flat/cmd/cli/main.go`

- [x] Check for existing .fmdx in cwd
- [ ] Auto-flatten if no .fmdx exists
- [ ] Error if .fmdx exists
- [ ] Help text for default mode

```go
// Default mode behavior:
// - If {cwd}.fmdx does NOT exist: flatten current directory
// - If {cwd}.fmdx EXISTS: error and require explicit command
```

## Testing Status

### Unit Tests

- [x] Flatten command structure
- [x] Unflatten command structure
- [x] Version command structure
- [ ] Flatten file iteration
- [ ] Unflatten file creation
- [ ] Hash verification
- [ ] Error handling

### Integration Tests

- [ ] Flatten single file
- [ ] Flatten directory
- [ ] Unflatten single file
- [ ] Unflatten directory
- [ ] Verify metadata preservation
- [ ] Verify permissions
- [ ] Verify timestamps
- [ ] Verify symlinks
- [ ] Verify xattrs
- [ ] Verify checksums
- [ ] Test --no-bin flag
- [ ] Test --external flag
- [ ] Test .flatignore
- [ ] Test verbose mode
- [ ] Test bypass-checksum flag

## Command Examples

### Flatten

```bash
# Flatten current directory (if no .fmdx exists)
flat

# Flatten with explicit command
flat flatten ./my-project ./backup.fmdx

# Flatten with verbose output
flat flatten ./project ./backup.fmdx -v

# Flatten excluding binaries
flat flatten ./project ./backup.fmdx --no-bin

# Flatten with custom ignore file
flat flatten ./project ./backup.fmdx --ignore-file .flatignore

# Flatten external references only
flat flatten ./project ./backup.fmdx --external
```

### Unflatten

```bash
# Unflatten to current directory
flat unflatten backup.fmdx ./restore

# Unflatten with verbose output
flat unflatten backup.fmdx ./restore -v

# Unflatten without checksum verification
flat unflatten backup.fmdx ./restore --bypass-checksum
```

### Version

```bash
flat version
```

## Notes

1. **Checksum verification** is enabled by default on unflatten
2. **All 5 hash types** are computed but only SHA-256 is verified
3. **External references** store path only, content remains at original location
4. **Binary files** are always base64 encoded in the .fmdx file
5. **Default mode** requires no .fmdx to exist in cwd

## Next Steps

1. Complete default mode implementation
2. Write comprehensive unit tests
3. Write integration tests
4. Test all flag combinations
5. Move to Phase 3: Finalization
