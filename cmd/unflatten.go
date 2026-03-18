//go:build !mage
// +build !mage

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"flat/config"
	"flat/encoder"
	"flat/format"
	"flat/hash"
	"flat/metadata"
)

// UnflattenCommand represents the unflatten command
type UnflattenCommand struct {
	Name  string
	Short string
	Long  string
	Run   func(*config.Config, []string) error
	Cfg   *config.Config
}

// Execute runs the unflatten command
func (c *UnflattenCommand) Execute(args []string) error {
	// Need at least 2 args (input.fmdx + destination-dir)
	// Additional args can be flags
	if len(args) < 2 {
		return fmt.Errorf("unflatten requires <input.fmdx> <destination-dir>")
	}
	return c.Run(c.Cfg, args)
}

// Unflatten runs the unflatten operation
func Unflatten(cfg *config.Config, args []string) error {
	inputFile := args[0]
	destDir := args[1]

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

	// Parse all entries
	entries, err := reader.ParseAllEntries()
	if err != nil {
		return fmt.Errorf("failed to parse .fmdx file: %w", err)
	}

	// Check for platform mismatch (from first file's header)
	if len(entries) > 0 {
		sourceOS := entries[0].Hashes["platform_os"]
		sourceArch := entries[0].Hashes["platform_arch"]
		sourceUID := entries[0].Hashes["platform_uid"]
		sourceGID := entries[0].Hashes["platform_gid"]

		currentOS := runtime.GOOS
		currentArch := runtime.GOARCH

		if sourceOS != currentOS || sourceArch != currentArch {
			fmt.Println("\n⚠️  Platform Mismatch Detected:")
			fmt.Printf("  Source: %s / %s uid:%s gid:%s\n", sourceOS, sourceArch, sourceUID, sourceGID)
			fmt.Printf("  Current: %s / %s\n", currentOS, currentArch)
			fmt.Println("\nCompatibility Warnings:")

			if sourceOS != currentOS {
				if currentOS == "windows" {
					fmt.Println("  • File permissions may not restore correctly (Windows uses ACLs, not Unix modes)")
					fmt.Println("  • Symlinks may not work or require admin privileges on Windows")
					fmt.Println("  • Extended attributes will be skipped (not supported on Windows)")
					fmt.Println("  • Line endings may change (LF → CRLF on Windows)")
				} else if currentOS == "darwin" {
					fmt.Println("  • File permissions may not restore exactly (macOS has different permission semantics)")
					fmt.Println("  • Some extended attributes may not transfer correctly")
				} else {
					fmt.Println("  • File permissions may not restore exactly")
					fmt.Println("  • Symlinks and extended attributes may have compatibility issues")
				}
			}

			if sourceUID != "" && sourceUID != "0" {
				fmt.Println("  • File ownership (UID/GID) will not be preserved (requires root/admin)")
			}

			fmt.Println("\nRestore will continue with best-effort compatibility.")
		}
	}

	if cfg.Verbose {
		fmt.Printf("Unflattening %s to %s\n", inputFile, destDir)
		fmt.Printf("Checksum verification: %v\n", !cfg.BypassChecksum)
		fmt.Println("---")
	}

	// Process each entry
	var skippedExcluded, restoredFiles int
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

		// Decode content based on content type
		content, err := encoder.DecodeContent(entry.Content, entry.Metadata.ContentType)
		if err != nil {
			if cfg.Verbose {
				fmt.Printf("  Warning: Could not decode content: %v\n", err)
			}
			continue
		}

		// Verify SHA-256 checksum (if not bypassing)
		if !cfg.BypassChecksum {
			computedHash := hash.ComputeAllHashes(content)
			expectedHash := entry.Hashes["file_hash"]
			if computedHash.SHA256 != expectedHash {
				return fmt.Errorf("checksum mismatch for %s (expected %s, got %s)",
					entry.Metadata.Path, expectedHash, computedHash.SHA256)
			}
		}

		// Write file
		mode, err := parseMode(entry.Metadata.Mode)
		if err != nil {
			if cfg.Verbose {
				fmt.Printf("  Warning: Could not parse mode %s: %v\n", entry.Metadata.Mode, err)
			}
			mode = 0644
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
		modTime, err := time.Parse(time.RFC3339, entry.Metadata.Modified)
		if err != nil {
			if cfg.Verbose {
				fmt.Printf("  Warning: Could not parse modified time: %v\n", err)
			}
			modTime = time.Now()
		}
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
			if err := metadata.SetXattr(destPath, key, value); err != nil {
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

// UnflattenCmd returns the unflatten command
func UnflattenCmd() *UnflattenCommand {
	return &UnflattenCommand{
		Name:  "unflatten",
		Short: "Unflatten a .fmdx file into a directory structure",
		Long:  "Unflatten a .fmdx file into a directory structure.",
		Run:   Unflatten,
		Cfg:   &config.Config{},
	}
}

// parseMode parses mode string (e.g., "0644" or "-rw-r--r--") to os.FileMode
func parseMode(mode string) (os.FileMode, error) {
	// Handle "-rw-r--r--" format
	if len(mode) >= 10 && mode[0] == '-' {
		// Convert permission string to octal
		perm := mode[1:]
		var result int64 = 0
		for i := 0; i < 3; i++ {
			result *= 8
			switch perm[i] {
			case 'r':
				result += 4
			case 'w':
				result += 2
			}
		}
		return os.FileMode(result), nil
	}

	// Handle "0644" format
	if len(mode) >= 4 && mode[0] == '0' {
		mode = mode[1:]
	}

	var num int
	fmt.Sscanf(mode, "%o", &num)

	return os.FileMode(num), nil
}
