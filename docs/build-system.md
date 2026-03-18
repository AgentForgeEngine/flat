# Build System

## Overview

Flat uses Mage for building.

## File Structure

```
flat/
├── main.go              # Entry point
├── go.mod               # Go module
├── go.sum               # Dependencies
├── Magefile.go          # Mage build commands
├── README.md            # User guide
└── docs/                # Documentation
```

## Dependencies

- **Go**: 1.26.1+
- **Mage**: 1.16.1+ (optional)

## Notes

- Mage is optional (requires `go install github.com/magefile/mage@latest`)
- Build artifacts are in the root directory
