package format

import (
	"fmt"
	"os"
	"strconv"

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

// WriteHeader writes the format header with platform info
func (w *FileWriter) WriteHeader(platformOS, platformArch, platformHostname string, platformUID, platformGID int) error {
	_, err := fmt.Fprintln(w.writer, HeaderStart)
	if err != nil {
		return err
	}

	// Write platform info after the multi marker
	_, err = fmt.Fprintln(w.writer, "platform_os: \""+platformOS+"\"")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.writer, "platform_arch: \""+platformArch+"\"")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.writer, "platform_hostname: \""+platformHostname+"\"")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.writer, "platform_uid: "+strconv.Itoa(platformUID))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.writer, "platform_gid: "+strconv.Itoa(platformGID))
	return err
}

// WriteFileEntry writes a single file entry to the .fmdx
func (w *FileWriter) WriteFileEntry(metadata *Metadata, content string, hashes *HashPair) error {
	encodedContent := content

	// Write hashes block
	err := w.writeHashesBlock(hashes, metadata)
	if err != nil {
		return err
	}

	// Write metadata block
	err = w.writeMetadataBlock(metadata)
	if err != nil {
		return err
	}

	// Write content (if not external)
	if !metadata.IsExternal {
		err = w.writeContentBlock(encodedContent)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeHashesBlock writes the hashes and content type
func (w *FileWriter) writeHashesBlock(hashes *HashPair, metadata *Metadata) error {
	if metadata.IsExternal {
		_, err := fmt.Fprintln(w.writer, "mdx_block_hash: "+metadata.BlockHash)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w.writer, "content_type: "+metadata.ContentType)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w.writer, "is_external: true")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w.writer, "external_path: "+metadata.ExternalPath)
		if err != nil {
			return err
		}
	} else {
		_, err := fmt.Fprintln(w.writer, "mdx_block_hash: "+hashes.BlockHash.SHA256)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w.writer, "file_hash: "+hashes.FileHash.SHA256)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w.writer, "content_type: "+metadata.ContentType)
		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(w.writer, HeaderEnd)
	return err
}

// writeMetadataBlock writes the YAML metadata
func (w *FileWriter) writeMetadataBlock(metadata *Metadata) error {
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.writer, string(yamlData))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w.writer, MetadataEnd)
	return err
}

// writeContentBlock writes the encoded content
func (w *FileWriter) writeContentBlock(content string) error {
	_, err := fmt.Fprintln(w.writer, content)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w.writer, ContentEnd)
	return err
}

// Metadata represents a file entry in the .fmdx format
type Metadata struct {
	Path           string            `yaml:"path"`
	Filename       string            `yaml:"filename"`
	Mode           string            `yaml:"mode"`
	Modified       string            `yaml:"modified"`
	Created        string            `yaml:"created"`
	Symlink        string            `yaml:"symlink"`
	Xattrs         map[string]string `yaml:"xattrs"`
	ContentType    string            `yaml:"content_type"`
	IsExternal     bool              `yaml:"is_external"`
	ExternalPath   string            `yaml:"external_path"`
	BlockHash      string            `yaml:"mdx_block_hash"`
	UID            int               `yaml:"uid,omitempty"`
	GID            int               `yaml:"gid,omitempty"`
	EndWithNewline bool              `yaml:"end_with_newline"`
}

// HashResult holds hash values
type HashResult struct {
	SHA256 string
	SHA512 string
	MD5    string
	BLAKE2 string
	CRC32  string
}

// HashPair holds two different hash results (e.g., for metadata vs content)
type HashPair struct {
	BlockHash *HashResult // Hash of YAML metadata block
	FileHash  *HashResult // Hash of file content
}

// DirectoryMetadata holds metadata for a directory entry
type DirectoryMetadata struct {
	Path     string `yaml:"path"`
	Type     string `yaml:"type"`
	Summary  string `yaml:"summary"`
	Created  string `yaml:"created"`
	Modified string `yaml:"modified"`
}

// WriteDirectoryEntry writes a directory entry to the .fmdx
func (w *FileWriter) WriteDirectoryEntry(dirMeta *DirectoryMetadata) error {
	// Write header with content type
	_, err := fmt.Fprintln(w.writer, "mdx_block_hash: ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.writer, "content_type: text/plain")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.writer, HeaderEnd)
	if err != nil {
		return err
	}

	// Write directory metadata
	yamlData, err := yaml.Marshal(dirMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal directory metadata: %w", err)
	}
	_, err = fmt.Fprint(w.writer, string(yamlData))
	if err != nil {
		return err
	}

	// Write directory end delimiter
	_, err = fmt.Fprintln(w.writer, "!--~---~END-DIRECTORY~--~---!")
	return err
}
