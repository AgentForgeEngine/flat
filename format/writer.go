package format

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// FileWriter handles writing .fmdx files
type FileWriter struct {
	outputPath string
	writer     *os.File
}

// NewWriter creates a new file writer
func NewWriter(outputPath string) (*FileWriter, error) {
	f, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}
	return &FileWriter{
		outputPath: outputPath,
		writer:     f,
	}, nil
}

// Close closes the output file
func (w *FileWriter) Close() error {
	return w.writer.Close()
}

// WriteHeader writes the format header
func (w *FileWriter) WriteHeader() error {
	_, err := fmt.Fprintln(w.writer, "---BEGIN-FLAT-FILE-MULTI---")
	return err
}

// WriteFileEntry writes a single file entry to the .fmdx
func (w *FileWriter) WriteFileEntry(metadata *Metadata, content string, hashes *HashResult) error {
	// Write metadata block with hashes
	err := w.writeHashesBlock(hashes, metadata)
	if err != nil {
		return err
	}

	// Write MDX section
	err = w.writeMDXSection(metadata)
	if err != nil {
		return err
	}

	// Write content (if not external)
	if !metadata.IsExternal {
		err = w.writeContentBlock(content)
		if err != nil {
			return err
		}
	}

	// Write MDX delimiter
	_, err = fmt.Fprintln(w.writer, "---MDX---")
	return err
}

// writeHashesBlock writes the hashes and metadata header
func (w *FileWriter) writeHashesBlock(hashes *HashResult, metadata *Metadata) error {
	_, err := fmt.Fprintln(w.writer, "---")
	if err != nil {
		return err
	}

	if metadata.IsExternal {
		_, err = fmt.Fprintf(w.writer, "mdx_block_hash: %s\n", metadata.BlockHash)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w.writer, "content_type: %s\n", metadata.ContentType)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w.writer, "is_external: true\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w.writer, "external_path: %s\n", metadata.ExternalPath)
		if err != nil {
			return err
		}
	} else {
		_, err = fmt.Fprintf(w.writer, "mdx_block_hash: %s\n", hashes.SHA256)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w.writer, "file_hash: %s\n", hashes.SHA256)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w.writer, "content_type: %s\n", metadata.ContentType)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(w.writer, "---")
	return err
}

// writeMDXSection writes the YAML metadata section
func (w *FileWriter) writeMDXSection(metadata *Metadata) error {
	// Write ---
	_, err := fmt.Fprintln(w.writer, "---")
	if err != nil {
		return err
	}

	// Write YAML metadata
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.writer, string(yamlData))
	if err != nil {
		return err
	}

	// Write ---
	_, err = fmt.Fprintln(w.writer, "---")
	return err
}

// writeContentBlock writes the base64 content
func (w *FileWriter) writeContentBlock(content string) error {
	_, err := fmt.Fprintln(w.writer, "---")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.writer, content)
	if err != nil {
		return err
	}
	return nil
}

// Metadata represents a file entry in the .fmdx format
type Metadata struct {
	Path         string            `yaml:"path"`
	Filename     string            `yaml:"filename"`
	Mode         string            `yaml:"mode"`
	Modified     string            `yaml:"modified"`
	Created      string            `yaml:"created"`
	Symlink      string            `yaml:"symlink"`
	Xattrs       map[string]string `yaml:"xattrs"`
	ContentType  string            `yaml:"content_type"`
	IsExternal   bool              `yaml:"is_external"`
	ExternalPath string            `yaml:"external_path"`
	BlockHash    string            `yaml:"mdx_block_hash"`
}

// HashResult holds hash values
type HashResult struct {
	SHA256 string
	SHA512 string
	MD5    string
	BLAKE2 string
	CRC32  string
}
