# Build System - Make & Mage

## Overview

Flat includes both Make and Mage build systems for flexibility.

## Make Commands

### Build
```bash
make build          # Build flat binary
make build-debug    # Build with debug symbols
make build-release  # Build optimized release binary
make cross-build    # Build for multiple platforms
```

### Install
```bash
make install        # Install to $GOPATH/bin
make uninstall      # Uninstall from $GOPATH/bin
```

### Test
```bash
make test           # Run all tests with coverage
make test-short     # Run tests without verbose output
make test-race      # Run tests with race detector
make coverage       # Run tests with coverage report
make coverage-html  # Generate HTML coverage report
```

### Lint
```bash
make lint           # Run linters (go vet, gofmt)
make format         # Format code with gofmt
make vet            # Run go vet
```

### Clean
```bash
make clean          # Remove build artifacts
```

### Other
```bash
make tidy           # Run go mod tidy
make docs           # Display documentation info
make help           # Display help
```

## Mage Commands

```bash
mage build          # Build flat binary
mage builddebug     # Build with debug symbols
mage buildrelease   # Build optimized release binary
mage crossbuild     # Build for multiple platforms
mage install        # Install to $GOPATH/bin
mage uninstall      # Uninstall from $GOPATH/bin
mage test           # Run all tests
mage testshort      # Run tests without verbose output
mage testrace       # Run tests with race detector
mage coverage       # Run tests with coverage report
mage coveragehtml   # Generate HTML coverage report
mage lint           # Run linters
mage format         # Format code
mage vet            # Run go vet
mage clean          # Remove build artifacts
mage tidy           # Run go mod tidy
mage docs           # Display documentation info
mage help           # Display help
```

## Usage Examples

### Build and Test
```bash
make build
make test
```

### Build and Install
```bash
make build
make install
# flat is now available in $GOPATH/bin
```

### Run Tests with Coverage
```bash
make coverage
make coverage-html
# Open coverage.html in browser
```

### Cross-Platform Build
```bash
make cross-build
# Creates: flat-linux-amd64, flat-linux-arm64, flat-darwin-amd64, flat-darwin-arm64, flat-windows-amd64.exe
```

## File Structure

```
flat/
├── main.go              # Entry point
├── go.mod               # Go module
├── go.sum               # Dependencies
├── Makefile             # Make build commands
├── Magefile.go          # Mage build commands
├── README.md            # User guide
└── docs/                # Documentation
```

## Dependencies

- **Go**: 1.26.1+
- **Mage**: 1.16.1+ (optional)
- **gofmt**: Included in Go toolchain
- **go test**: Included in Go toolchain

## Notes

- Make is the primary build system (works without additional tools)
- Mage is optional (requires `go install github.com/magefile/mage@latest`)
- All build artifacts are in the root directory
- Coverage reports are generated in the root directory
