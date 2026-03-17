# Phase 1: Core Implementation

## Overview

Phase 1 focuses on building the foundational components of the Flat tool:

1. **Project Setup**: Go module, directory structure
2. **CLI Framework**: Cobra + Viper integration
3. **Hash Computation**: All 5 checksum algorithms
4. **Format Parser/Writer**: Core .fmdx structure handling
5. **Base64 Encoding**: Content encoding/decoding

## Implementation Steps

### Step 1: Project Initialization

#### Create Directory Structure

```bash
mkdir -p flat/cmd flat/config flat/format flat/hash flat/metadata flat/encoder
```

#### Initialize Go Module

```bash
go mod init flat
```

#### Add Dependencies

```bash
go get github.com/spf13/cobra@v1.8.0
go get github.com/spf13/viper@v1.18.2
go get golang.org/x/crypto@v0.17.0  # For BLAKE2
go get gopkg.in/yaml.v3@v3.0.1      # For YAML
```

#### Create Initial Files

```
flat/
├── main.go              # Root command + default mode
├── go.mod
├── go.sum
└── README.md
```

### Step 2: Config Module

#### File: `config/config.go`

```go
package config

import (
    "github.com/spf13/viper"
)

// Config holds all configuration for the flat tool
type Config struct {
    Verbose        bool
    NoBin          bool
    External       bool
    Exclude        []string
    IgnoreFile     string
    BypassChecksum bool
}

// LoadConfig loads configuration from environment variables and flags
func LoadConfig() *Config {
    cfg := &Config{
        Verbose:    viper.GetBool("FLAT_VERBOSE"),
        IgnoreFile: ".flatignore",
    }
    return cfg
}

// SetVerbose sets verbose mode
func (c *Config) SetVerbose(v bool) {
    c.Verbose = v
}

// SetNoBin sets binary skip mode
func (c *Config) SetNoBin(v bool) {
    c.NoBin = v
}
```

#### File: `cmd/config.go` (Viper setup)

```go
package cmd

import (
    "github.com/spf13/viper"
)

func InitViper() {
    viper.SetConfigName(".flatconfig")
    viper.SetConfigType("toml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("$HOME/.flat")
    viper.AutomaticEnv()  // Read env variables like FLAT_VERBOSE
}
```

### Step 3: Hash Module

#### File: `hash/hash.go`

```go
package hash

import (
    "crypto/md5"
    "crypto/sha256"
    "crypto/sha512"
    "hash/crc32"
    "hash/fnv"

    "golang.org/x/crypto/blake2b"
)

// HashResult holds all computed hash values
type HashResult struct {
    SHA256 string
    SHA512 string
    MD5    string
    BLAKE2 string
    CRC32  string
}

// ComputeAllHashes computes all 5 hash algorithms for the given content
func ComputeAllHashes(content []byte) *HashResult {
    result := &HashResult{}

    // SHA-256
    sha256Hash := sha256.Sum256(content)
    result.SHA256 = toHex(sha256Hash[:])

    // SHA-512
    sha512Hash := sha512.Sum512(content)
    result.SHA512 = toHex(sha512Hash[:])

    // MD5
    md5Hash := md5.Sum(content)
    result.MD5 = toHex(md5Hash[:])

    // BLAKE2
    blake2Hash, _ := blake2b.Sum256(content)
    result.BLAKE2 = toHex(blake2Hash[:])

    // CRC32
    crc := crc32.ChecksumIEEE(content)
    result.CRC32 = formatCRC32(crc)

    return result
}

// ComputeMDXBlockHash computes hash of YAML metadata block (for integrity)
func ComputeMDXBlockHash(yamlContent string) *HashResult {
    return ComputeAllHashes([]byte(yamlContent))
}

// VerifySHA256 verifies SHA-256 hash against computed hash
func VerifySHA256(content []byte, expectedSHA256 string) bool {
    result := ComputeAllHashes(content)
    return result.SHA256 == expectedSHA256
}

// Helper functions
func toHex(b []byte) string {
    hex := make([]byte, len(b)*2)
    for i, v := range b {
        hex[i*2] = "0123456789abcdef"[v>>4]
        hex[i*2+1] = "0123456789abcdef"[v&0xf]
    }
    return string(hex)
}

func formatCRC32(crc uint32) string {
    return formatUint32(crc)
}

func formatUint32(u uint32) string {
    return fmt.Sprintf("%08x", u)
}
```

### Step 4: Encoder Module

#### File: `encoder/base64.go`

```go
package encoder

import "encoding/base64"

// Encode encodes binary content to base64
func Encode(content []byte) string {
    return base64.StdEncoding.EncodeToString(content)
}

// Decode decodes base64 content to binary
func Decode(encoded string) ([]byte, error) {
    return base64.StdEncoding.DecodeString(encoded)
}

// EncodeFile encodes file content to base64
func EncodeFile(path string) (string, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    return Encode(content), nil
}

// DecodeFile decodes base64 content and writes to file
func DecodeFile(encoded string, path string, mode os.FileMode) error {
    content, err := Decode(encoded)
    if err != nil {
        return err
    }
    return os.WriteFile(path, content, mode)
}
```

### Step 5: Format Writer

#### File: `format/writer.go`

```go
package format

import (
    "fmt"
    "os"
    "strings"

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
func (w *FileWriter) WriteFileEntry(metadata *Metadata, content string, hashes *hash.HashResult) error {
    // Write metadata block with hashes
    metadataBlock := fmt.Sprintf(`mdx_block_hash: %s
file_hash: %s
content_type: %s
`, hashes.SHA256, hashes.SHA256, metadata.ContentType)

    if metadata.IsExternal {
        metadataBlock = fmt.Sprintf(`mdx_block_hash: %s
content_type: %s
is_external: true
external_path: %s
`, metadata.BlockHash, metadata.ContentType, metadata.ExternalPath)
    }

    err := w.writeYAMLBlock(metadataBlock)
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

// Helper: Write YAML block wrapped in --- delimiters
func (w *FileWriter) writeYAMLBlock(yamlContent string) error {
    _, err := fmt.Fprintln(w.writer, "---")
    if err != nil {
        return err
    }
    _, err = fmt.Fprint(w.writer, yamlContent)
    if err != nil {
        return err
    }
    _, err = fmt.Fprintln(w.writer, "---")
    return err
}

// Helper: Write MDX section
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

// Helper: Write content block
func (w *FileWriter) writeContentBlock(content string) error {
    _, err := fmt.Fprintln(w.writer, "---")
    if err != nil {
        return err
    }
    _, err = fmt.Fprint(w.writer, content)
    if err != nil {
        return err
    }
    return err
}
```

### Step 6: Format Parser

#### File: `format/parser.go`

```go
package format

import (
    "bufio"
    "fmt"
    "os"
    "strings"
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
    // Read metadata block (between first --- and ---MDX---)
    metadataBlock, hashes, err := r.readMetadataBlock()
    if err != nil {
        return nil, err
    }
    if metadataBlock == "" {
        return nil, nil
    }

    // Parse metadata YAML
    metadata, err := parseYAML(metadataBlock)
    if err != nil {
        return nil, err
    }

    // Parse hashes
    hashesObj := parseHashes(hashes)

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
        Hashes:   hashesObj,
        Content:  content,
    }, nil
}

// readMetadataBlock reads the metadata block with hashes
func (r *FileReader) readMetadataBlock() (string, string, error) {
    var metadataBlock strings.Builder
    var hashesBlock strings.Builder

    // Read lines until ---MDX---
    for r.scanner.Scan() {
        line := r.scanner.Text()

        if strings.HasPrefix(line, "---MDX---") {
            break
        }

        // Separate hashes block and metadata block
        if strings.HasPrefix(line, "mdx_block_hash:") ||
           strings.HasPrefix(line, "file_hash:") ||
           strings.HasPrefix(line, "content_type:") ||
           strings.HasPrefix(line, "is_external:") ||
           strings.HasPrefix(line, "external_path:") {
            hashesBlock.WriteString(line + "\n")
        } else {
            metadataBlock.WriteString(line + "\n")
        }
    }

    return metadataBlock.String(), hashesBlock.String(), nil
}

// readContentBlock reads the base64 content
func (r *FileReader) readContentBlock() (string, error) {
    var contentBuilder strings.Builder

    for r.scanner.Scan() {
        line := r.scanner.Text()

        if strings.HasPrefix(line, "---MDX---") {
            break
        }

        contentBuilder.WriteString(line + "\n")
    }

    return strings.TrimSpace(contentBuilder.String()), nil
}
```

### Step 7: Metadata Collector

#### File: `metadata/collector.go`

```go
package metadata

import (
    "os"
    "syscall"
    "time"
)

// Metadata holds all POSIX metadata for a file
type Metadata struct {
    Path         string            `yaml:"path"`
    Filename     string            `yaml:"filename"`
    Mode         string            `yaml:"mode"`
    Modified     time.Time         `yaml:"modified"`
    Created      time.Time         `yaml:"created"`
    Symlink      string            `yaml:"symlink"`
    Xattrs       map[string]string `yaml:"xattrs"`
    ContentType  string            `yaml:"content_type"`
    IsExternal   bool              `yaml:"is_external"`
    ExternalPath string            `yaml:"external_path"`
    BlockHash    string            `yaml:"mdx_block_hash"`
}

// Collect gathers all metadata for a file
func Collect(filepath string, relPath string) (*Metadata, error) {
    stat, err := os.Lstat(filepath)
    if err != nil {
        return nil, err
    }

    metadata := &Metadata{
        Path:        relPath,
        Filename:    stat.Name(),
        Mode:        stat.Mode().String(),
        Modified:    stat.ModTime(),
        Xattrs:      make(map[string]string),
    }

    // Try to get created time (varies by OS)
    if bt, ok := stat.Sys().(*syscall.Stat_t); ok {
        metadata.Created = time.Unix(int64(bt.Ctim.Sec), int64(bt.Ctim.Nsec))
    } else {
        metadata.Created = stat.ModTime() // Fallback
    }

    // Check if symlink
    if stat.Mode()&os.ModeSymlink != 0 {
        target, err := os.Readlink(filepath)
        if err != nil {
            return nil, err
        }
        metadata.Symlink = target
    }

    // Collect extended attributes
    xattrs, err := getXattrs(filepath)
    if err != nil {
        // Warning only, continue without xattrs
    }
    metadata.Xattrs = xattrs

    // Auto-detect content type
    metadata.ContentType = detectContentType(filepath, stat)

    return metadata, nil
}

// CollectExternal gathers metadata for external reference (no content)
func CollectExternal(filepath string, relPath string) (*Metadata, error) {
    stat, err := os.Lstat(filepath)
    if err != nil {
        return nil, err
    }

    metadata := &Metadata{
        Path:         relPath,
        Filename:     stat.Name(),
        Mode:         stat.Mode().String(),
        Modified:     stat.ModTime(),
        Created:      time.Now(),
        Symlink:      "",
        Xattrs:       make(map[string]string),
        ContentType:  detectContentType(filepath, stat),
        IsExternal:   true,
        ExternalPath: filepath,
    }

    return metadata, nil
}

// Helper: Get extended attributes
func getXattrs(filepath string) (map[string]string, error) {
    xattrs := make(map[string]string)

    // Get user.* attributes
    attrs, err := listxattr(filepath)
    if err != nil {
        return xattrs, err
    }

    for _, attr := range attrs {
        value, err := getxattr(filepath, string(attr))
        if err == nil {
            xattrs[string(attr)] = string(value)
        }
    }

    return xattrs, nil
}

// Helper: Auto-detect content type
func detectContentType(filepath string, stat os.FileInfo) string {
    ext := strings.ToLower(filepath[strings.LastIndex(filepath, ".")+1:])

    mimeMap := map[string]string{
        "txt":  "text/plain",
        "md":   "text/markdown",
        "go":   "text/x-go",
        "js":   "application/javascript",
        "json": "application/json",
        "yaml": "application/yaml",
        "yml":  "application/yaml",
        "html": "text/html",
        "css":  "text/css",
        "png":  "image/png",
        "jpg":  "image/jpeg",
        "jpeg": "image/jpeg",
        "gif":  "image/gif",
        "mp3":  "audio/mpeg",
        "mp4":  "video/mp4",
        "mov":  "video/quicktime",
        "zip":  "application/zip",
        "gz":   "application/gzip",
    }

    if mime, ok := mimeMap[ext]; ok {
        return mime
    }

    return "application/octet-stream"
}
```

### Step 8: Binary Detection

#### File: `format/magic.go`

```go
package format

import (
    "io"
    "os"
)

// MagicByte represents a magic byte signature
type MagicByte struct {
    Signature []byte
    MIMEType  string
}

// Common magic byte signatures
var magicSignatures = []MagicByte{
    {Signature: []byte{0x89, 0x50, 0x4E, 0x47}, MIMEType: "image/png"},
    {Signature: []byte{0xFF, 0xD8, 0xFF}, MIMEType: "image/jpeg"},
    {Signature: []byte{0x47, 0x49, 0x46, 0x38}, MIMEType: "image/gif"},
    {Signature: []byte{0x50, 0x4B, 0x03, 0x04}, MIMEType: "application/zip"},
    {Signature: []byte{0x7F, 0x45, 0x4C, 0x46}, MIMEType: "application/x-elf"},
    {Signature: []byte("ID3"), MIMEType: "audio/mpeg"},
    {Signature: []byte{0x1F, 0x8B}, MIMEType: "application/gzip"},
}

// IsBinary detects if a file is binary using magic bytes and extension
func IsBinary(filepath string) (bool, string) {
    // First check file extension
    ext := strings.ToLower(filepath[strings.LastIndex(filepath, ".")+1:])
    binaryExtensions := map[string]bool{
        "png": true, "jpg": true, "jpeg": true, "gif": true,
        "mp3": true, "mp4": true, "mov": true, "avi": true,
        "exe": true, "dll": true, "bin": true, "iso": true,
        "zip": true, "gz": true, "tar": true, "7z": true,
    }

    if binaryExtensions[ext] {
        return true, "binary"
    }

    // Check magic bytes
    f, err := os.Open(filepath)
    if err != nil {
        return false, ""
    }
    defer f.Close()

    buf := make([]byte, 16)
    n, err := io.ReadFull(f, buf)
    if err != nil {
        return false, ""
    }

    for _, sig := range magicSignatures {
        if len(buf) >= len(sig.Signature) {
            match := true
            for i := range sig.Signature {
                if buf[i] != sig.Signature[i] {
                    match = false
                    break
                }
            }
            if match {
                return true, sig.MIMEType
            }
        }
    }

    return false, ""
}
```

### Step 9: Ignore File Parser

#### File: `format/ignore.go`

```go
package format

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"
)

// IgnoreParser handles .flatignore file parsing
type IgnoreParser struct {
    patterns []string
}

// NewIgnoreParser creates a new ignore parser
func NewIgnoreParser(ignorePath string) (*IgnoreParser, error) {
    parser := &IgnoreParser{
        patterns: make([]string, 0),
    }

    f, err := os.Open(ignorePath)
    if err != nil {
        if os.IsNotExist(err) {
            return parser, nil // No ignore file, no patterns
        }
        return nil, err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        parser.patterns = append(parser.patterns, line)
    }

    return parser, scanner.Err()
}

// ShouldIgnore checks if a path should be ignored
func (p *IgnoreParser) ShouldIgnore(relPath string) bool {
    filename := filepath.Base(relPath)

    for _, pattern := range p.patterns {
        if matchesPattern(pattern, relPath, filename) {
            return true
        }
    }

    return false
}

// matchesPattern checks if a pattern matches a path
func matchesPattern(pattern, relPath, filename string) bool {
    // Directory pattern (ends with /)
    if strings.HasSuffix(pattern, "/") {
        dirPattern := strings.TrimSuffix(pattern, "/")
        return strings.HasPrefix(relPath, dirPattern+"/") || relPath == dirPattern
    }

    // Extension pattern (starts with *)
    if strings.HasPrefix(pattern, "*.") {
        ext := pattern[1:]
        return strings.HasSuffix(filename, "."+ext)
    }

    // Exact filename match
    if !strings.Contains(pattern, "*") {
        return filename == pattern || relPath == pattern
    }

    // Glob pattern
    match, _ := filepath.Match(pattern, filename)
    return match
}
```

### Step 10: Main CLI (Phase 1 Partial)

#### File: `main.go`

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    flat/cmd
)

var rootCmd = &cobra.Command{
    Use: "flat",
    Short: "Flat - Flatten/unflatten directory trees",
    Long: `Flat is a tool to flatten directory structures into a single .fmdx file
and unflatten them back.

When run alone (flat):
  - If {cwd}.fmdx doesn't exist: auto-flatten current directory
  - If {cwd}.fmdx exists: error and require explicit command`,
    RunE: runDefaultMode,
}

func runDefaultMode(cmd *cobra.Command, args []string) error {
    cwd, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get working directory: %w", err)
    }

    fmdxPath := filepath.Join(cwd, filepath.Base(cwd)+".fmdx")

    if _, err := os.Stat(fmdxPath); os.IsNotExist(err) {
        // Auto-flatten
        fmt.Printf("Auto-flattening %s to %s\n", cwd, fmdxPath)
        return cmdFlatten(nil, []string{cwd, fmdxPath})
    }

    return fmt.Errorf("%s already exists. Use 'flat flatten' or 'flat unflatten'", filepath.Base(cwd)+".fmdx")
}

func main() {
    cmd.InitViper()
    rootCmd.AddCommand(cmd.FlattenCmd())
    rootCmd.AddCommand(cmd.UnflattenCmd())
    rootCmd.AddCommand(cmd.VersionCmd())

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## Phase 1 Deliverables

By the end of Phase 1, we will have:

1. ✅ Go module initialized with all dependencies
2. ✅ Config module (env vars, viper setup)
3. ✅ Hash module (5 algorithms: SHA-256, SHA-512, MD5, BLAKE2, CRC32)
4. ✅ Encoder module (base64 encode/decode)
5. ✅ Format writer (write .fmdx files)
6. ✅ Format parser (parse .fmdx files)
7. ✅ Metadata collector (POSIX metadata)
8. ✅ Binary detection (magic bytes + extension)
9. ✅ Ignore parser (.flatignore patterns)
10. ✅ Basic CLI structure (cobra commands)

## Phase 2 Preview

Phase 2 will focus on:

1. Complete flatten command implementation
2. Complete unflatten command implementation
3. External reference handling
4. Checksum verification
5. Full CLI flag integration
6. Error handling and edge cases
7. Testing and documentation
