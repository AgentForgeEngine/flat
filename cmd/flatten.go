//go:build !mage
// +build !mage

package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"flat/config"
	"flat/encoder"
	"flat/format"
	"flat/hash"
	"flat/metadata"
)

// FlattenCommand represents the flatten command
type FlattenCommand struct {
	Name  string
	Short string
	Long  string
	Run   func(*config.Config, []string) error
	Cfg   *config.Config
}

// Execute runs the flatten command
func (c *FlattenCommand) Execute(args []string) error {
	// Need at least 2 args (source-dir + output.fmdx)
	// Additional args can be flags
	if len(args) < 2 {
		return fmt.Errorf("flatten requires <source-dir> <output.fmdx>")
	}
	return c.Run(c.Cfg, args)
}

// Flatten runs the flatten operation
func Flatten(cfg *config.Config, args []string) error {
	sourceDir := args[0]
	outputPath := args[1]

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

	// Capture platform info
	platformOS := runtime.GOOS
	platformArch := runtime.GOARCH

	var platformUID, platformGID int
	if currentUser, err := user.Current(); err == nil {
		if uid, err := strconv.Atoi(currentUser.Uid); err == nil {
			platformUID = uid
		}
		if gid, err := strconv.Atoi(currentUser.Gid); err == nil {
			platformGID = gid
		}
	}

	hostname, _ := os.Hostname()

	// Write header with platform info
	if err := writer.WriteHeader(platformOS, platformArch, hostname, platformUID, platformGID); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if cfg.Verbose {
		fmt.Printf("Flattening %s to %s\n", sourceAbs, outputPath)
		fmt.Printf("Ignore file: %s\n", cfg.IgnoreFile)
		fmt.Printf("Binary files: %v\n", !cfg.NoBin)
		fmt.Printf("External refs: %v\n", cfg.External)
		fmt.Printf("Platform: %s/%s uid:%d gid:%d\n", platformOS, platformArch, platformUID, platformGID)
		fmt.Println("---")
	}

	// Counters
	var totalFiles, skippedBinary, skippedExcluded, flattenedFiles int
	var startTime = time.Now()

	// Walk directory
	err = filepath.Walk(sourceAbs, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			if cfg.Verbose {
				fmt.Printf("Error accessing %s: %v\n", filePath, err)
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceAbs, filePath)
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
			isBin, _ := format.IsBinary(filePath)
			if isBin {
				skippedBinary++
				if cfg.Verbose {
					fmt.Printf("Skipping (binary): %s\n", relPath)
				}
				return nil
			}
		}

		// Collect metadata
		var flatMeta *format.Metadata

		if cfg.External && info.Mode().IsRegular() {
			// External reference - no content
			meta, err := metadata.CollectExternal(filePath, relPath)
			if err != nil {
				if cfg.Verbose {
					fmt.Printf("Warning: Could not collect external metadata for %s: %v\n", relPath, err)
				}
				return nil
			}

			flatMeta = &format.Metadata{
				Path:         relPath,
				Filename:     meta.Filename,
				Mode:         meta.Mode,
				Modified:     meta.Modified.Format(time.RFC3339),
				Created:      meta.Created.Format(time.RFC3339),
				Symlink:      "",
				Xattrs:       meta.Xattrs,
				ContentType:  meta.ContentType,
				IsExternal:   true,
				ExternalPath: meta.ExternalPath,
				BlockHash:    "",
				UID:          meta.UID,
				GID:          meta.GID,
			}

			// Write entry
			hashResult := &format.HashPair{
				BlockHash: &format.HashResult{},
				FileHash:  &format.HashResult{},
			}
			if err := writer.WriteFileEntry(flatMeta, "", hashResult); err != nil {
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
			meta, err := metadata.Collect(filePath, relPath)
			if err != nil {
				if cfg.Verbose {
					fmt.Printf("Warning: Could not collect metadata for %s: %v\n", relPath, err)
				}
				return nil
			}

			// Read content
			content, err := os.ReadFile(filePath)
			if err != nil {
				if cfg.Verbose {
					fmt.Printf("Warning: Could not read content for %s: %v\n", relPath, err)
				}
				return nil
			}

			// Compute hashes for content
			contentHash := hash.ComputeAllHashes(content)

			flatMeta = &format.Metadata{
				Path:        relPath,
				Filename:    meta.Filename,
				Mode:        meta.Mode,
				Modified:    meta.Modified.Format(time.RFC3339),
				Created:     meta.Created.Format(time.RFC3339),
				Symlink:     "",
				Xattrs:      meta.Xattrs,
				ContentType: meta.ContentType,
				IsExternal:  false,
				UID:         meta.UID,
				GID:         meta.GID,
			}

			// Encode content based on content type
			encoded := encoder.EncodeContent(content, meta.ContentType)

			// Convert hash result to format.HashResult
			contentHashFormat := &format.HashResult{
				SHA256: contentHash.SHA256,
				SHA512: contentHash.SHA512,
				MD5:    contentHash.MD5,
				BLAKE2: contentHash.BLAKE2,
				CRC32:  contentHash.CRC32,
			}

			// Create hash pair (same hash for both since we hash original content)
			hashResultFormat := &format.HashPair{
				BlockHash: contentHashFormat,
				FileHash:  contentHashFormat,
			}

			// Write entry
			if err := writer.WriteFileEntry(flatMeta, encoded, hashResultFormat); err != nil {
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
			meta, err := metadata.Collect(filePath, relPath)
			if err != nil {
				if cfg.Verbose {
					fmt.Printf("Warning: Could not collect metadata for symlink %s: %v\n", relPath, err)
				}
				return nil
			}

			target, _ := os.Readlink(filePath)

			flatMeta = &format.Metadata{
				Path:        relPath,
				Filename:    meta.Filename,
				Mode:        meta.Mode,
				Modified:    meta.Modified.Format(time.RFC3339),
				Created:     meta.Created.Format(time.RFC3339),
				Symlink:     target,
				Xattrs:      meta.Xattrs,
				ContentType: meta.ContentType,
				IsExternal:  cfg.External,
				UID:         meta.UID,
				GID:         meta.GID,
			}

			// External symlinks (path only)
			if cfg.External {
				hashResultFormat := &format.HashPair{
					BlockHash: &format.HashResult{},
					FileHash:  &format.HashResult{},
				}
				if err := writer.WriteFileEntry(flatMeta, "", hashResultFormat); err != nil {
					return err
				}
				flattenedFiles++

				if cfg.Verbose {
					fmt.Printf("External symlink: %s\n", relPath)
				}
				return nil
			}

			// Regular symlink (store metadata, no content)
			hashResultFormat := &format.HashPair{
				BlockHash: &format.HashResult{},
				FileHash:  &format.HashResult{},
			}
			if err := writer.WriteFileEntry(flatMeta, "", hashResultFormat); err != nil {
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

// FlattenCmd returns the flatten command
func FlattenCmd() *FlattenCommand {
	return &FlattenCommand{
		Name:  "flatten",
		Short: "Flatten a directory tree into a .fmdx file",
		Long:  "Flatten a directory tree into a single .fmdx file.",
		Run:   Flatten,
		Cfg:   &config.Config{},
	}
}
