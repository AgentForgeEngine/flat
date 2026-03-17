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

### Project Setup
- [ ] Initialize Go module (`go mod init flat`)
- [ ] Create directory structure
- [ ] Add dependencies (cobra, viper, yaml, blake2)
- [ ] Create .gitignore

### Config Module (`config/config.go`)
- [ ] Config struct definition
- [ ] Environment variable loading (FLAT_VERBOSE)
- [ ] Viper initialization
- [ ] Flag binding

### Hash Module (`hash/hash.go`)
- [ ] SHA-256 computation
- [ ] SHA-512 computation
- [ ] MD5 computation
- [ ] BLAKE2 computation
- [ ] CRC32 computation
- [ ] Hash result struct
- [ ] Helper functions (toHex, etc.)

### Encoder Module (`encoder/base64.go`)
- [ ] Base64 encode function
- [ ] Base64 decode function
- [ ] File encode helper
- [ ] File decode helper

### Format Writer (`format/writer.go`)
- [ ] FileWriter struct
- [ ] Write header function
- [ ] Write file entry function
- [ ] Write YAML block function
- [ ] Write MDX section function
- [ ] Write content block function
- [ ] Close function

### Format Parser (`format/parser.go`)
- [ ] FileReader struct
- [ ] Validate header function
- [ ] Parse all entries function
- [ ] Parse single entry function
- [ ] Read metadata block function
- [ ] Read content block function
- [ ] Parse YAML helper
- [ ] Parse hashes helper

### Metadata Collector (`metadata/collector.go`)
- [ ] Metadata struct definition
- [ ] Collect function (regular files)
- [ ] CollectExternal function (external refs)
- [ ] Get extended attributes
- [ ] List extended attributes
- [ ] Set extended attributes
- [ ] Detect content type
- [ ] Parse mode helper

### Binary Detection (`format/magic.go`)
- [ ] Magic byte signatures
- [ ] IsBinary function
- [ ] Magic byte checking
- [ ] Extension checking
- [ ] Combined detection logic

### Ignore Parser (`format/ignore.go`)
- [ ] IgnoreParser struct
- [ ] New parser function
- [ ] ShouldIgnore function
- [ ] Pattern matching
- [ ] Glob pattern support

## Phase 2: Commands

### Flatten Command (`cmd/flatten.go`)
- [ ] Cobra command definition
- [ ] Flag definitions
- [ ] RunE function
- [ ] Source directory validation
- [ ] .flatignore parsing
- [ ] FileWriter initialization
- [ ] Header writing
- [ ] Directory walk function
- [ ] Relative path calculation
- [ ] Ignore pattern checking
- [ ] Binary file detection & skipping
- [ ] Metadata collection
- [ ] Content reading
- [ ] Hash computation
- [ ] Base64 encoding
- [ ] File entry writing
- [ ] Verbose output
- [ ] Summary statistics

### Unflatten Command (`cmd/unflatten.go`)
- [ ] Cobra command definition
- [ ] Flag definitions
- [ ] RunE function
- [ ] Input file validation
- [ ] Destination directory creation
- [ ] FileReader initialization
- [ ] Header validation
- [ ] Parse all entries
- [ ] Destination path calculation
- [ ] Parent directory creation
- [ ] External reference handling
- [ ] Base64 decoding
- [ ] SHA-256 verification
- [ ] Mode parsing
- [ ] File writing
- [ ] Permission restoration
- [ ] Timestamp restoration
- [ ] Symlink handling
- [ ] Xattr restoration
- [ ] Verbose output
- [ ] Summary statistics

### Version Command (`cmd/version.go`)
- [ ] Cobra command definition
- [ ] Version variable
- [ ] Print version function

### Main CLI (`main.go`)
- [ ] Root command definition
- [ ] Default mode logic
- [ ] Check for {cwd}.fmdx
- [ ] Auto-flatten if missing
- [ ] Error if exists
- [ ] Add subcommands
- [ ] Viper initialization
- [ ] Main function

## Phase 3: Testing & Release

### Test Data Setup
- [ ] Create test/data/text/ directory
- [ ] Create test/data/binary/ directory
- [ ] Create test/data/symlinks/ directory
- [ ] Create test/data/special/ directory
- [ ] Create test/data/permissions/ directory
- [ ] Create test/data/xattrs/ directory
- [ ] Create .flatignore file for testing

### Unit Tests
- [ ] hash/hash_test.go
- [ ] encoder/base64_test.go
- [ ] format/format_test.go
- [ ] metadata/metadata_test.go
- [ ] config/config_test.go

### Integration Tests
- [ ] test/flatten_test.go
- [ ] test/unflatten_test.go
- [ ] test/integration_test.go

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

### Documentation
- [ ] README.md (already created)
- [ ] Update with actual usage
- [ ] Add examples
- [ ] Create .gitignore
- [ ] Update version numbers

### Release Preparation
- [ ] Build binary (linux, darwin, windows)
- [ ] Create release notes
- [ ] Tag version (v0.1.0)
- [ ] Create GitHub release
- [ ] Test download links

## Environment Variables

- [ ] FLAT_VERBOSE implementation

## Error Handling

- [ ] All commands have proper error messages
- [ ] Exit codes used correctly
- [ ] Viper error handling
- [ ] File I/O error handling
- [ ] Checksum mismatch errors

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
- MDX delimiters: ---MDX---
- No FMDX delimiter (only MDX)
- Output extension: .fmdx
- Default mode: `flat` auto-flattens if no .fmdx exists
