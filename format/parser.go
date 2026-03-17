package format

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileReader handles reading .fmdx files
type FileReader struct {
	inputPath string
	scanner   *bufio.Scanner
	line      int
}

// NewReader creates a new file reader
func NewReader(inputPath string) (*FileReader, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	return &FileReader{
		inputPath: inputPath,
		scanner:   bufio.NewScanner(f),
		line:      0,
	}, nil
}

// Close closes the input file
func (r *FileReader) Close() error {
	return r.scanner.Err()
}

// ValidateHeader validates the file header
func (r *FileReader) ValidateHeader() error {
	if !r.scanner.Scan() {
		return fmt.Errorf("empty file")
	}
	line := r.scanner.Text()
	if line != "---BEGIN-FLAT-FILE-MULTI---" {
		return fmt.Errorf("invalid file header: expected '---BEGIN-FLAT-FILE-MULTI---', got '%s'", line)
	}
	return nil
}

// ParseAllEntries parses all file entries from the .fmdx
func (r *FileReader) ParseAllEntries() ([]*FileEntry, error) {
	var entries []*FileEntry

	for r.scanner.Scan() {
		entry, err := r.parseEntry()
		if err != nil {
			return nil, err
		}
		if entry != nil {
			entries = append(entries, entry)
		}
	}

	if err := r.scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// parseEntry parses a single file entry
func (r *FileReader) parseEntry() (*FileEntry, error) {
	// Read hashes block
	hashesBlock, err := r.readHashesBlock()
	if err != nil {
		return nil, err
	}
	if hashesBlock == nil || len(hashesBlock) == 0 {
		return nil, nil
	}

	// Read MDX section (metadata)
	metadata, err := r.readMDXSection()
	if err != nil {
		return nil, err
	}
	if metadata == nil {
		return nil, nil
	}

	// Read content (if not external)
	var content string
	if !metadata.IsExternal {
		content, err = r.readContentBlock()
		if err != nil {
			return nil, err
		}
	}

	return &FileEntry{
		Metadata: metadata,
		Content:  content,
	}, nil
}

// readHashesBlock reads the hashes block
func (r *FileReader) readHashesBlock() (map[string]string, error) {
	hashes := make(map[string]string)

	for r.scanner.Scan() {
		line := r.scanner.Text()

		if strings.HasPrefix(line, "---MDX---") {
			break
		}

		if strings.HasPrefix(line, "---") {
			continue
		}

		// Parse key: value
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				hashes[key] = value
			}
		}
	}

	return hashes, nil
}

// readMDXSection reads the YAML metadata section
func (r *FileReader) readMDXSection() (*Metadata, error) {
	var yamlContent strings.Builder

	// Skip the opening ---
	if !r.scanner.Scan() {
		return nil, fmt.Errorf("unexpected end of file in MDX section")
	}

	// Read until ---MDX---
	for !strings.HasPrefix(r.scanner.Text(), "---MDX---") {
		yamlContent.WriteString(r.scanner.Text() + "\n")

		if !r.scanner.Scan() {
			return nil, fmt.Errorf("unexpected end of file in MDX section")
		}
	}

	// Parse YAML
	var metadata Metadata
	err := yaml.Unmarshal([]byte(yamlContent.String()), &metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

// readContentBlock reads the base64 content
func (r *FileReader) readContentBlock() (string, error) {
	var contentBuilder strings.Builder

	// Skip the opening ---
	if !r.scanner.Scan() {
		return "", fmt.Errorf("unexpected end of file before content")
	}

	// Read until ---MDX---
	for !strings.HasPrefix(r.scanner.Text(), "---MDX---") {
		contentBuilder.WriteString(r.scanner.Text() + "\n")

		if !r.scanner.Scan() {
			// End of file without MDX marker
			break
		}
	}

	return strings.TrimSpace(contentBuilder.String()), nil
}

// FileEntry represents a single file entry in the .fmdx
type FileEntry struct {
	Metadata *Metadata
	Content  string
	Hashes   *HashResult
}
