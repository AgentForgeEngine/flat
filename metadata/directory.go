package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// MaxSummarySize is the maximum size for directory summaries (8KB)
const MaxSummarySize = 8192

// DirectoryMetadata holds metadata for a directory
type DirectoryMetadata struct {
	Path     string `yaml:"path"`
	Type     string `yaml:"type"`
	Summary  string `yaml:"summary"`
	Created  string `yaml:"created"`
	Modified string `yaml:"modified"`
}

// CollectDirectory gathers metadata for a directory with .agent file
func CollectDirectory(dirPath string, relPath string) (*DirectoryMetadata, error) {
	stat, err := os.Lstat(dirPath)
	if err != nil {
		return nil, err
	}

	// Read .agent file
	agentPath := filepath.Join(dirPath, ".agent")
	agentContent, err := os.ReadFile(agentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .agent: %w", err)
	}

	// Parse YAML
	var dirMeta DirectoryMetadata
	if err := yaml.Unmarshal(agentContent, &dirMeta); err != nil {
		return nil, fmt.Errorf("failed to parse .agent: %w", err)
	}

	// Validate path
	if dirMeta.Path == "" {
		dirMeta.Path = relPath
	}

	// Validate type
	if dirMeta.Type == "" {
		dirMeta.Type = "directory"
	}

	// Validate summary
	if dirMeta.Summary == "" {
		return nil, fmt.Errorf("summary is required in .agent")
	}

	// Check summary size
	if len(dirMeta.Summary) > MaxSummarySize {
		return nil, fmt.Errorf("summary exceeds %d bytes limit (got %d bytes)",
			MaxSummarySize, len(dirMeta.Summary))
	}

	// Set timestamps from filesystem
	dirMeta.Modified = stat.ModTime().Format(time.RFC3339)
	dirMeta.Created = time.Now().Format(time.RFC3339)

	// Try to get created time
	if bt, ok := stat.Sys().(*struct {
		Btimes []struct {
			Sec  int
			Nsec int
		}
	}); ok && len(bt.Btimes) > 0 {
		dirMeta.Created = time.Unix(int64(bt.Btimes[0].Sec), int64(bt.Btimes[0].Nsec)).Format(time.RFC3339)
	}

	return &dirMeta, nil
}

// ReadFlatdir reads and validates a .agent file
func ReadFlatdir(agentPath string) (*DirectoryMetadata, error) {
	content, err := os.ReadFile(agentPath)
	if err != nil {
		return nil, err
	}

	var dirMeta DirectoryMetadata
	if err := yaml.Unmarshal(content, &dirMeta); err != nil {
		return nil, fmt.Errorf("failed to parse .agent: %w", err)
	}

	// Validate required fields
	if dirMeta.Summary == "" {
		return nil, fmt.Errorf("summary is required")
	}

	// Check size limit
	if len(dirMeta.Summary) > MaxSummarySize {
		return nil, fmt.Errorf("summary exceeds %d bytes limit (got %d bytes)",
			MaxSummarySize, len(dirMeta.Summary))
	}

	// Default values
	if dirMeta.Type == "" {
		dirMeta.Type = "directory"
	}

	return &dirMeta, nil
}

// WriteFlatdir writes a .agent file with directory summary
func WriteFlatdir(dirPath string, summary string) error {
	agentPath := filepath.Join(dirPath, ".agent")

	content, err := yaml.Marshal(DirectoryMetadata{
		Type:    "directory",
		Summary: summary,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal .agent: %w", err)
	}

	if err := os.WriteFile(agentPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write .agent: %w", err)
	}

	return nil
}

// WriteAgents writes AGENTS.yaml file for a directory
func WriteAgents(dirPath string, dirMeta *DirectoryMetadata) error {
	agentsPath := filepath.Join(dirPath, "AGENTS.yaml")

	// Ensure path is relative
	path := dirMeta.Path
	if filepath.IsAbs(path) {
		path = strings.TrimPrefix(path, "/")
	}

	content, err := yaml.Marshal(DirectoryMetadata{
		Path:     path,
		Type:     dirMeta.Type,
		Summary:  dirMeta.Summary,
		Created:  dirMeta.Created,
		Modified: dirMeta.Modified,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal AGENTS.yaml: %w", err)
	}

	if err := os.WriteFile(agentsPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write AGENTS.yaml: %w", err)
	}

	return nil
}

// HasFlatdir checks if a directory has a .agent file
func HasFlatdir(dirPath string) bool {
	agentPath := filepath.Join(dirPath, ".agent")
	_, err := os.Stat(agentPath)
	return err == nil
}

// FindFlatdirs recursively finds all directories with .agent files
func FindFlatdirs(rootPath string) ([]string, error) {
	var agents []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if HasFlatdir(path) {
				relPath, err := filepath.Rel(rootPath, path)
				if err != nil {
					return err
				}
				agents = append(agents, relPath)
			}
		}

		return nil
	})

	return agents, err
}
