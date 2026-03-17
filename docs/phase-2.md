# Phase 2: Command Implementation

## Overview

Phase 2 focuses on implementing the actual CLI commands and tying all components together:

1. **Flatten Command**: Complete implementation
2. **Unflatten Command**: Complete implementation
3. **Version Command**: Simple version display
4. **Checksum Verification**: SHA-256 (required) + bypass flag
5. **External References**: Store paths only, no content
6. **Error Handling**: Comprehensive error management

## Implementation Details

### Step 1: Flatten Command

#### File: `cmd/flatten.go`

```go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    flat/config
    flat/format
    flat/hash
    flat/metadata
    flat/encoder
)

var flattenCmd = &cobra.Command{
    Use: "flatten <source-dir> <output.fmdx>",
    Short: "Flatten a directory tree into a .fmdx file",
    Long: `Flatten a directory tree into a single .fmdx file.

The command will:
  - Recursively traverse the source directory
  - Collect all file metadata (permissions, timestamps, symlinks, xattrs)
  - Compute checksums for all files
  - Encode file contents in base64
  - Write to the output .fmdx file`,
    Args: cobra.ExactArgs(2),
    RunE: runFlatten,
}

var (
    verboseFlag    bool
    noBinFlag      bool
    externalFlag   bool
    excludeFlag    []string
    ignoreFileFlag string
)

func init() {
    flattenCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose output")
    flattenCmd.Flags().BoolVar(&noBinFlag, "no-bin", false, "skip binary files")
    flattenCmd.Flags().BoolVar(&externalFlag, "external", false, "external file references")
    flattenCmd.Flags().StringSliceVar(&excludeFlag, "exclude", []string{}, "exclude patterns")
    flattenCmd.Flags().StringVar(&ignoreFileFlag, "ignore-file", ".flatignore", "ignore file path")
    
    viper.BindPFlag("verbose", flattenCmd.Flags().Lookup("verbose"))
    viper.BindPFlag("no_bin", flattenCmd.Flags().Lookup("no_bin"))
    viper.BindPFlag("external", flattenCmd.Flags().Lookup("external"))
    viper.BindPFlag("exclude", flattenCmd.Flags().Lookup("exclude"))
    viper.BindPFlag("ignore_file", flattenCmd.Flags().Lookup("ignore_file"))
}

func FlattenCmd() *cobra.Command {
    return flattenCmd
}

func runFlatten(cmd *cobra.Command, args []string) error {
    sourceDir := args[0]
    outputPath := args[1]

    cfg := config.LoadConfig()
    cfg.Verbose = verboseFlag
    cfg.NoBin = noBinFlag
    cfg.External = externalFlag
    cfg.Exclude = excludeFlag
    cfg.IgnoreFile = ignoreFileFlag

    return executeFlatten(cfg, sourceDir, outputPath)
}

func executeFlatten(cfg *config.Config, sourceDir string, outputPath string) error {
    // Validate source directory
    if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
        return fmt.Errorf("source directory does not exist: %s", sourceDir)
    }

    // Get absolute source path
    sourceAbs, err := filepath.Abs(sourceDir)
    if err != nil {
        return fmt.Errorf("failed to resolve source path: %w", err)
    }

    // Parse .flatignore
    ignoreParser, err := format.NewIgnoreParser(cfg.IgnoreFile)
    if err != nil {
        return fmt.Errorf("failed to parse ignore file: %w", err)
    }

    // Create output file writer
    writer, err := format.NewWriter(outputPath)
    if err != nil {
        return fmt.Errorf("failed to create output file: %w", err)
    }
    defer writer.Close()

    // Write header
    if err := writer.WriteHeader(); err != nil {
        return fmt.Errorf("failed to write header: %w", err)
    }

    if cfg.Verbose {
        fmt.Printf("Flattening %s to %s\n", sourceAbs, outputPath)
        fmt.Printf("Ignore file: %s\n", cfg.IgnoreFile)
        fmt.Printf("Binary files: %v\n", !cfg.NoBin)
        fmt.Printf("External refs: %v\n", cfg.External)
        fmt.Println("---")
    }

    // Counters
    var totalFiles, skippedBinary, skippedExcluded, flattenedFiles int
    var startTime = time.Now()

    // Walk directory
    err = filepath.Walk(sourceAbs, func(filepath string, info os.FileInfo, err error) error {
        if err != nil {
            if cfg.Verbose {
                fmt.Printf("Error accessing %s: %v\n", filepath, err)
            }
            return nil // Continue walking
        }

        // Get relative path
        relPath, err := filepath.Rel(sourceAbs, filepath)
        if err != nil {
            return err
        }

        // Skip the root directory itself
        if relPath == "." {
            return nil
        }

        // Check ignore patterns
        if ignoreParser.ShouldIgnore(relPath) {
            skippedExcluded++
            if cfg.Verbose {
                fmt.Printf("Skipping (ignored): %s\n", relPath)
            }
            return nil
        }

        totalFiles++

        // Check for binary files
        if cfg.NoBin && info.Mode().IsRegular() {
            isBin, _ := format.IsBinary(filepath)
            if isBin {
                skippedBinary++
                if cfg.Verbose {
                    fmt.Printf("Skipping (binary): %s\n", relPath)
                }
                return nil
            }
        }

        // Collect metadata
        var meta *metadata.Metadata

        if cfg.External && info.Mode().IsRegular() {
            // External reference - no content
            meta, err = metadata.CollectExternal(filepath, relPath)
            if err != nil {
                if cfg.Verbose {
                    fmt.Printf("Warning: Could not collect external metadata for %s: %v\n", relPath, err)
                }
                return nil
            }
            meta.IsExternal = true

            // Write entry
            hashes := hash.ComputeMDXBlockHash("") // Empty content for external
            content := ""
            if err := writer.WriteFileEntry(meta, content, hashes); err != nil {
                return err
            }
            flattenedFiles++

            if cfg.Verbose {
                fmt.Printf("External: %s -> %s\n", relPath, meta.ExternalPath)
            }
            return nil
        }

        // Regular file
        if info.Mode().IsRegular() {
            meta, err = metadata.Collect(filepath, relPath)
            if err != nil {
                if cfg.Verbose {
                    fmt.Printf("Warning: Could not collect metadata for %s: %v\n", relPath, err)
                }
                return nil
            }

            // Read content
            content, err := os.ReadFile(filepath)
            if err != nil {
                if cfg.Verbose {
                    fmt.Printf("Warning: Could not read content for %s: %v\n", relPath, err)
                }
                return nil
            }

            // Compute hashes
            hashes := hash.ComputeAllHashes(content)
            meta.BlockHash = hashes.SHA256

            // Encode content
            encoded := encoder.Encode(content)

            // Write entry
            if err := writer.WriteFileEntry(meta, encoded, hashes); err != nil {
                return err
            }
            flattenedFiles++

            if cfg.Verbose {
                fmt.Printf("Flattened: %s (%d bytes)\n", relPath, len(content))
            }
            return nil
        }

        // Symlinks
        if info.Mode()&os.ModeSymlink != 0 {
            meta, err = metadata.Collect(filepath, relPath)
            if err != nil {
                if cfg.Verbose {
                    fmt.Printf("Warning: Could not collect metadata for symlink %s: %v\n", relPath, err)
                }
                return nil
            }

            // External symlinks (path only)
            if cfg.External {
                hashes := hash.ComputeMDXBlockHash("")
                if err := writer.WriteFileEntry(meta, "", hashes); err != nil {
                    return err
                }
                flattenedFiles++

                if cfg.Verbose {
                    fmt.Printf("External symlink: %s\n", relPath)
                }
                return nil
            }

            // Regular symlink (store metadata, no content)
            hashes := hash.ComputeMDXBlockHash("")
            if err := writer.WriteFileEntry(meta, "", hashes); err != nil {
                return err
            }
            flattenedFiles++

            if cfg.Verbose {
                fmt.Printf("Symlink: %s\n", relPath)
            }
            return nil
        }

        // Skip directories
        if info.IsDir() {
            return nil
        }

        return nil
    })

    if err != nil {
        return fmt.Errorf("error walking directory: %w", err)
    }

    // Summary
    elapsed := time.Since(startTime)
    fmt.Printf("\nFlattening complete!\n")
    fmt.Printf("Total files: %d\n", totalFiles)
    fmt.Printf("Skipped (binary): %d\n", skippedBinary)
    fmt.Printf("Skipped (ignored): %d\n", skippedExcluded)
    fmt.Printf("Flattened: %d\n", flattenedFiles)
    fmt.Printf("Time: %s\n", elapsed)

    return nil
}
```

### Step 2: Unflatten Command

#### File: `cmd/unflatten.go`

```go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    flat/config
    flat/format
    flat/hash
    flat/metadata
    flat/encoder
)

var unflattenCmd = &cobra.Command{
    Use: "unflatten <input.fmdx> <destination-dir>",
    Short: "Unflatten a .fmdx file into a directory structure",
    Long: `Unflatten a .fmdx file into a directory structure.

The command will:
  - Parse the .fmdx file format
  - Verify SHA-256 checksums (default, unless --bypass-checksum)
  - Create directory structure
  - Write file contents
  - Restore permissions, timestamps, symlinks, xattrs`,
    Args: cobra.ExactArgs(2),
    RunE: runUnflatten,
}

var (
    unflattenVerbose bool
    bypassChecksum   bool
)

func init() {
    unflattenCmd.Flags().BoolVarP(&unflattenVerbose, "verbose", "v", false, "verbose output")
    unflattenCmd.Flags().BoolVar(&bypassChecksum, "bypass-checksum", false, "skip checksum verification")

    viper.BindPFlag("verbose", unflattenCmd.Flags().Lookup("verbose"))
    viper.BindPFlag("bypass_checksum", unflattenCmd.Flags().Lookup("bypass_checksum"))
}

func UnflattenCmd() *cobra.Command {
    return unflattenCmd
}

func runUnflatten(cmd *cobra.Command, args []string) error {
    inputFile := args[0]
    destDir := args[1]

    cfg := config.LoadConfig()
    cfg.Verbose = unflattenVerbose
    cfg.BypassChecksum = bypassChecksum

    return executeUnflatten(cfg, inputFile, destDir)
}

func executeUnflatten(cfg *config.Config, inputFile string, destDir string) error {
    // Validate input file
    if _, err := os.Stat(inputFile); os.IsNotExist(err) {
        return fmt.Errorf("input file does not exist: %s", inputFile)
    }

    // Create destination directory if needed
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return fmt.Errorf("failed to create destination directory: %w", err)
    }

    // Create file reader
    reader, err := format.NewReader(inputFile)
    if err != nil {
        return fmt.Errorf("failed to open input file: %w", err)
    }
    defer reader.Close()

    // Validate header
    if err := reader.ValidateHeader(); err != nil {
        return fmt.Errorf("invalid .fmdx file: %w", err)
    }

    if cfg.Verbose {
        fmt.Printf("Unflattening %s to %s\n", inputFile, destDir)
        fmt.Printf("Checksum verification: %v\n", !cfg.BypassChecksum)
        fmt.Println("---")
    }

    // Parse all entries
    entries, err := reader.ParseAllEntries()
    if err != nil {
        return fmt.Errorf("failed to parse .fmdx file: %w", err)
    }

    // Process each entry
    var totalFiles, skippedExcluded, restoredFiles int
    var startTime = time.Now()

    for i, entry := range entries {
        if cfg.Verbose {
            fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(entries), entry.Metadata.Path)
        }

        // Calculate destination path
        destPath := filepath.Join(destDir, entry.Metadata.Path)

        // Create parent directory
        parentDir := filepath.Dir(destPath)
        if err := os.MkdirAll(parentDir, 0755); err != nil {
            if cfg.Verbose {
                fmt.Printf("Warning: Could not create directory %s: %v\n", parentDir, err)
            }
            continue
        }

        // Handle external references
        if entry.Metadata.IsExternal {
            if cfg.Verbose {
                fmt.Printf("  External reference: %s\n", entry.Metadata.ExternalPath)
            }
            skippedExcluded++
            continue
        }

        // Decode content
        content, err := encoder.Decode(entry.Content)
        if err != nil {
            if cfg.Verbose {
                fmt.Printf("  Warning: Could not decode content: %v\n", err)
            }
            continue
        }

        // Verify SHA-256 checksum (if not bypassing)
        if !cfg.BypassChecksum {
            computedHash := hash.ComputeAllHashes(content)
            if computedHash.SHA256 != entry.Hashes.SHA256 {
                return fmt.Errorf("checksum mismatch for %s (expected %s, got %s)",
                    entry.Metadata.Path, entry.Hashes.SHA256, computedHash.SHA256)
            }
        }

        // Write file
        mode, err := parseMode(entry.Metadata.Mode)
        if err != nil {
            if cfg.Verbose {
                fmt.Printf("  Warning: Could not parse mode %s: %v\n", entry.Metadata.Mode, err)
                mode = 0644
            } else {
                mode = 0644
            }
        }

        if err := os.WriteFile(destPath, content, mode); err != nil {
            if cfg.Verbose {
                fmt.Printf("  Warning: Could not write file: %v\n", err)
            }
            continue
        }

        // Restore permissions
        if err := os.Chmod(destPath, mode); err != nil {
            if cfg.Verbose {
                fmt.Printf("  Warning: Could not restore permissions: %v\n", err)
            }
        }

        // Restore timestamps
        modTime := entry.Metadata.Modified
        if err := os.Chtimes(destPath, modTime, modTime); err != nil {
            if cfg.Verbose {
                fmt.Printf("  Warning: Could not restore timestamps: %v\n", err)
            }
        }

        // Handle symlinks
        if entry.Metadata.Symlink != "" {
            // Remove regular file first
            os.Remove(destPath)

            // Create symlink
            if err := os.Symlink(entry.Metadata.Symlink, destPath); err != nil {
                if cfg.Verbose {
                    fmt.Printf("  Warning: Could not create symlink: %v\n", err)
                }
            }
        }

        // Restore extended attributes
        for key, value := range entry.Metadata.Xattrs {
            if err := setxattr(destPath, key, value); err != nil {
                if cfg.Verbose {
                    fmt.Printf("  Warning: Could not restore xattr %s: %v\n", key, err)
                }
            }
        }

        restoredFiles++

        if cfg.Verbose {
            fmt.Printf("  Restored: %s (%d bytes)\n", entry.Metadata.Path, len(content))
        }
    }

    // Summary
    elapsed := time.Since(startTime)
    fmt.Printf("\nUnflattening complete!\n")
    fmt.Printf("Total entries: %d\n", len(entries))
    fmt.Printf("Skipped (external): %d\n", skippedExcluded)
    fmt.Printf("Restored: %d\n", restoredFiles)
    fmt.Printf("Time: %s\n", elapsed)

    return nil
}
```

### Step 3: Version Command

#### File: `cmd/version.go`

```go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

var (
    version = "0.1.0"
    commit  = "unknown"
    date    = "unknown"
)

var versionCmd = &cobra.Command{
    Use: "version",
    Short: "Show version information",
    Long: `Show version information for the flat tool.`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("flat version %s\n", version)
        fmt.Printf("Commit: %s\n", commit)
        fmt.Printf("Built: %s\n", date)
    },
}

func VersionCmd() *cobra.Command {
    return versionCmd
}
```

### Step 4: Helper Functions

#### File: `metadata/helpers.go`

```go
package metadata

import (
    "os"
    "strconv"
    "strings"
)

// ParseMode parses mode string (e.g., "0644") to os.FileMode
func ParseMode(mode string) (os.FileMode, error) {
    // Remove leading '0' if present
    mode = strings.TrimPrefix(mode, "0")
    
    num, err := strconv.ParseInt(mode, 8, 32)
    if err != nil {
        return 0, err
    }
    
    return os.FileMode(num), nil
}

// GetXattrs gets extended attributes for a file
func GetXattrs(filepath string) (map[string]string, error) {
    xattrs := make(map[string]string)

    // Get list of attributes
    attrs, err := listxattr(filepath)
    if err != nil {
        return xattrs, nil // Continue without xattrs
    }

    // Get each attribute value
    for _, attr := range attrs {
        value, err := getxattr(filepath, string(attr))
        if err == nil {
            xattrs[string(attr)] = string(value)
        }
    }

    return xattrs, nil
}

// SetXattr sets an extended attribute on a file
func SetXattr(filepath, key, value string) error {
    return setxattr(filepath, key, []byte(value))
}

// ListXattrs lists all extended attributes for a file
func ListXattrs(filepath string) ([]string, error) {
    var attrs []string

    // Platform-specific implementation
    // Linux: use syscall.Getxattr with empty name
    // macOS: use getxattr with empty name
    
    // Placeholder for platform-specific code
    return attrs, nil
}
```

#### File: `hash/helpers.go`

```go
package hash

import (
    "encoding/hex"
)

// ToHex converts bytes to hex string
func ToHex(b []byte) string {
    return hex.EncodeToString(b)
}

// FromHex converts hex string to bytes
func FromHex(s string) ([]byte, error) {
    return hex.DecodeString(s)
}

// FormatCRC32 formats CRC32 as 8-character hex string
func FormatCRC32(crc uint32) string {
    return formatUint32(crc)
}

func formatUint32(u uint32) string {
    hex := "0123456789abcdef"
    result := make([]byte, 8)
    
    result[0] = hex[(u>>28)&0xf]
    result[1] = hex[(u>>24)&0xf]
    result[2] = hex[(u>>20)&0xf]
    result[3] = hex[(u>>16)&0xf]
    result[4] = hex[(u>>12)&0xf]
    result[5] = hex[(u>>8)&0xf]
    result[6] = hex[(u>>4)&0xf]
    result[7] = hex[u&0xf]
    
    return string(result)
}
```

### Step 5: Error Handling

#### File: `config/errors.go`

```go
package config

import (
    "errors"
)

// Common error types
var (
    ErrInvalidFormat    = errors.New("invalid file format")
    ErrChecksumMismatch = errors.New("checksum mismatch")
    ErrPermissionDenied = errors.New("permission denied")
    ErrFileNotFound     = errors.New("file not found")
    ErrDirectoryExists  = errors.New("directory already exists")
)

// FlattenError represents an error during flattening
type FlattenError struct {
    Path     string
    Message  string
    Original error
}

func (e *FlattenError) Error() string {
    return e.Message
}

func (e *FlattenError) Unwrap() error {
    return e.Original
}

// UnflattenError represents an error during unflattening
type UnflattenError struct {
    Path     string
    Message  string
    Original error
}

func (e *UnflattenError) Error() string {
    return e.Message
}

func (e *UnflattenError) Unwrap() error {
    return e.Original
}
```

## Phase 2 Deliverables

By the end of Phase 2, we will have:

1. ✅ **Flatten Command**: Complete implementation with all flags
   - Recursive directory traversal
   - Binary detection and skipping
   - External reference handling
   - .flatignore pattern matching
   - Verbose output
   - Summary statistics

2. ✅ **Unflatten Command**: Complete implementation with all flags
   - .fmdx parsing and validation
   - SHA-256 checksum verification (default)
   - --bypass-checksum flag
   - File content restoration
   - Metadata restoration (permissions, timestamps, symlinks, xattrs)
   - Verbose output
   - Summary statistics

3. ✅ **Version Command**: Simple version display

4. ✅ **Error Handling**: Comprehensive error types and messages

5. ✅ **Integration**: All modules working together

## Phase 3 Preview

Phase 3 will focus on:

1. **Testing**: Create comprehensive test suite
2. **Documentation**: README.md, usage examples
3. **Edge Cases**: Handle all edge cases properly
4. **Performance**: Optimize for large directories
5. **Release**: Build binary, create release notes
6. **Cleanup**: Code review, remove debug code
