# Flat Implementation Checklist

## Phase 0: Documentation ✓

- [x] docs/phase-0.md - Specification
- [x] docs/phase-1.md - Core implementation
- [x] docs/phase-2.md - Commands
- [x] docs/phase-3.md - Testing & Release
- [x] docs/overview.md - Quick reference
- [x] README.md - User guide
- [x] .flatignore.example - Example ignore file

## Phase 1: Core Implementation

### Project Setup ✅ [COMPLETE]
- [x] Initialize Go module (`go mod init flat`)
- [x] Create directory structure
- [x] Add dependencies (cobra, yaml, blake2)
- [x] Create .gitignore

### Config Module (`config/config.go`) ✅ [COMPLETE]
- [x] Config struct definition
- [x] Environment variable loading (FLAT_VERBOSE)
- [x] Flag binding (via args)

### Hash Module (`hash/hash.go`) ✅ [COMPLETE]
- [x] SHA-256 computation
- [x] SHA-512 computation
- [x] MD5 computation
- [x] BLAKE2 computation
- [x] CRC32 computation
- [x] Hash result struct
- [x] Helper functions (toHex, etc.)

### Encoder Module (`encoder/base64.go`) ✅ [COMPLETE]
- [x] Base64 encode function
- [x] Base64 decode function
- [x] File encode helper
- [x] File decode helper

### Format Writer (`format/writer.go`) ✅ [COMPLETE]
- [x] FileWriter struct
- [x] Write header function
- [x] Write file entry function
- [x] Write YAML block function
- [x] Write MDX section function
- [x] Write content block function
- [x] Close function

### Format Parser (`format/parser.go`) ✅ [COMPLETE]
- [x] FileReader struct
- [x] Validate header function
- [x] Parse all entries function
- [x] Parse single entry function
- [x] Read metadata block function
- [x] Read content block function
- [x] Parse YAML helper
- [x] Parse hashes helper

### Metadata Collector (`metadata/collector.go`) ✅ [COMPLETE]
- [x] Metadata struct definition
- [x] Collect function (regular files)
- [x] CollectExternal function (external refs)
- [x] Get extended attributes
- [x] List extended attributes
- [x] Set extended attributes
- [x] Detect content type
- [x] Parse mode helper

### Binary Detection (`format/magic.go`) ✅ [COMPLETE]
- [x] Magic byte signatures
- [x] IsBinary function
- [x] Magic byte checking
- [x] Extension checking
- [x] Combined detection logic

### Ignore Parser (`format/ignore.go`) ✅ [COMPLETE]
- [x] IgnoreParser struct
- [x] New parser function
- [x] ShouldIgnore function
- [x] Pattern matching
- [x] Glob pattern support

## Phase 2: Commands

### Flatten Command (`cmd/flatten.go`) ✅ [COMPLETE]
- [x] Cobra command definition
- [x] Flag definitions
- [x] RunE function
- [x] Source directory validation
- [x] .flatignore parsing
- [x] FileWriter initialization
- [x] Header writing
- [x] Directory walk function
- [x] Relative path calculation
- [x] Ignore pattern checking
- [x] Binary file detection & skipping
- [x] Metadata collection
- [x] Content reading
- [x] Hash computation
- [x] Base64 encoding
- [x] File entry writing
- [x] Verbose output
- [x] Summary statistics

### Unflatten Command (`cmd/unflatten.go`) ✅ [COMPLETE]
- [x] Cobra command definition
- [x] Flag definitions
- [x] RunE function
- [x] Input file validation
- [x] Destination directory creation
- [x] FileReader initialization
- [x] Header validation
- [x] Parse all entries
- [x] Destination path calculation
- [x] Parent directory creation
- [x] External reference handling
- [x] Base64 decoding
- [x] SHA-256 verification
- [x] Mode parsing
- [x] File writing
- [x] Permission restoration
- [x] Timestamp restoration
- [x] Symlink handling
- [x] Xattr restoration
- [x] Verbose output
- [x] Summary statistics

### Version Command (`cmd/version.go`) ✅ [COMPLETE]
- [x] Cobra command definition
- [x] Version variable
- [x] Print version function

### Main CLI (`main.go`) ✅ [COMPLETE]
- [x] Root command definition
- [x] Default mode logic
- [x] Check for {cwd}.fmdx
- [x] Auto-flatten if missing
- [x] Error if exists
- [x] Add subcommands
- [x] Main function

## Phase 3: Testing & Release ✅ [IN PROGRESS]

### Test Data Setup ✅ [COMPLETE]
- [x] Create test/data/text/ directory
- [x] Create test/data/binary/ directory
- [x] Create test/data/symlinks/ directory
- [x] Create test/data/special/ directory
- [x] Create test/data/permissions/ directory
- [x] Create test/data/xattrs/ directory
- [x] Create .flatignore file for testing

### Unit Tests ✅ [COMPLETE]
- [x] hash/hash_test.go
- [x] encoder/base64_test.go
- [x] format/format_test.go (magic.go, ignore.go)
- [x] metadata/metadata_test.go
- [x] config/config_test.go

### Integration Tests ✅ [COMPLETE]
- [x] test/integration_test.go (flatten, unflatten, empty files)

### Edge Case Tests
- [ ] Empty files
- [ ] Symlinks
- [ ] Permissions
- [ ] Special characters
- [ ] Binary files
- [ ] External references
- [ ] Checksum verification
- [ ] .flatignore patterns

### Performance Tests
- [ ] Large directory test
- [ ] Memory usage test
- [ ] Time benchmark

### Documentation ✅ [COMPLETE]
- [x] README.md (already created)
- [x] Update with actual usage
- [x] Add examples
- [x] Create .gitignore
- [x] Update version numbers

### Release Preparation ⏳ [PENDING]
- [ ] Build binary (linux, darwin, windows)
- [ ] Create release notes
- [ ] Tag version (v0.1.0)
- [ ] Create GitHub release
- [ ] Test download links

## Environment Variables ✅ [COMPLETE]

- [x] FLAT_VERBOSE implementation

## Error Handling ✅ [COMPLETE]

- [x] All commands have proper error messages
- [x] Exit codes used correctly
- [x] File I/O error handling
- [x] Checksum mismatch errors

## Code Quality

- [ ] No TODO comments
- [ ] All functions have tests
- [ ] Code follows Go standards
- [ ] Comments for complex logic
- [ ] No hardcoded paths

## Final Checks

- [ ] All tests passing
- [ ] go fmt applied
- [ ] go vet clean
- [ ] README.md accurate
- [ ] Documentation complete
- [ ] Binary builds successfully

## Notes

- SHA-256 is always verified on unflatten (unless --bypass-checksum)
- All 5 hash algorithms are computed
- Format uses ---BEGIN-FLAT-FILE-MULTI--- header
- Section delimiters: !--~---~   ~--~---!
- No FMDX delimiter (only section delimiter)
- Output extension: .fmdx
- Default mode: `flat` auto-flattens if no .fmdx exists

## Phase 4: Directory Metadata ✅ [COMPLETE]

### Implementation Tasks
- [x] Create metadata/directory.go with DirectoryMetadata struct
- [x] Add WriteDirectoryEntry() to format/writer.go
- [x] Add DIR-ENTRY delimiter support to format/parser.go
- [x] Update cmd/flatten.go to read .flatdir files
- [x] Update cmd/unflatten.go to create directories and AGENTS.yaml
- [x] Enforce 8KB summary limit
- [x] Add directory metadata to phase-0.md
- [x] Update README.md with directory metadata examples
- [x] Create integration tests for directory metadata
- [x] Test .flatdir parsing and AGENTS.yaml generation
