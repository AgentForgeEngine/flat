package metadata

import (
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
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
	UID          int               `yaml:"uid,omitempty"`
	GID          int               `yaml:"gid,omitempty"`
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
		Created:     time.Now(),
		ContentType: detectContentType(filepath),
	}

	// Try to get created time (varies by OS)
	if bt, ok := stat.Sys().(*syscall.Stat_t); ok {
		metadata.Created = time.Unix(int64(bt.Ctim.Sec), int64(bt.Ctim.Nsec))
	}

	// Check if symlink
	if stat.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(filepath)
		if err != nil {
			return nil, err
		}
		metadata.Symlink = target
	}

	// Collect UID/GID from stat
	if statSys, ok := stat.Sys().(*syscall.Stat_t); ok {
		metadata.UID = int(statSys.Uid)
		metadata.GID = int(statSys.Gid)
	}

	// Collect extended attributes
	xattrs, err := getXattrs(filepath)
	if err == nil && len(xattrs) > 0 {
		metadata.Xattrs = xattrs
	}

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
		ContentType:  detectContentType(filepath),
		IsExternal:   true,
		ExternalPath: filepath,
	}

	return metadata, nil
}

// getXattrs gets extended attributes for a file
func getXattrs(filepath string) (map[string]string, error) {
	xattrs := make(map[string]string)

	// Get list of attributes
	attrs, err := listxattr(filepath)
	if err != nil {
		return xattrs, err
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

// listxattr lists all extended attributes for a file
func listxattr(filepath string) ([]string, error) {
	var attrs []string

	// Get attribute list using unix.Listxattr - first call to get size
	buf := make([]byte, 4096)
	n, err := unix.Listxattr(filepath, buf)
	if err != nil {
		return attrs, err
	}

	if n <= 0 {
		return attrs, nil
	}

	// Trim to actual size
	buf = buf[:n]

	attrs = make([]string, 0)
	for i := 0; i < len(buf); {
		for j := i; j < len(buf) && buf[j] != 0; j++ {
			attrs = append(attrs, string(buf[i:j]))
			i = j + 1
		}
		i++
	}

	return attrs, nil
}

// getxattr gets a single extended attribute value
func getxattr(filepath, key string) ([]byte, error) {
	// Use unix.Getxattr to get extended attribute - first call to get size
	buf := make([]byte, 4096)
	n, err := unix.Getxattr(filepath, key, buf)
	if err != nil {
		return []byte{}, err
	}

	if n <= 0 {
		return []byte{}, nil
	}

	return buf[:n], nil
}

// setxattr sets an extended attribute on a file
func setxattr(filepath, key string, value []byte) error {
	// Use unix.Setxattr to set extended attribute
	err := unix.Setxattr(filepath, key, value, 0)
	if err != nil {
		return err
	}

	return nil
}

// SetXattr sets an extended attribute on a file (public wrapper)
func SetXattr(filepath, key, value string) error {
	return setxattr(filepath, key, []byte(value))
}

// detectContentType auto-detects MIME type based on file extension
func detectContentType(filepath string) string {
	ext := getFileExtension(filepath)

	mimeMap := map[string]string{
		"txt":   "text/plain",
		"md":    "text/markdown",
		"go":    "text/x-go",
		"js":    "application/javascript",
		"json":  "application/json",
		"yaml":  "application/yaml",
		"yml":   "application/yaml",
		"html":  "text/html",
		"css":   "text/css",
		"png":   "image/png",
		"jpg":   "image/jpeg",
		"jpeg":  "image/jpeg",
		"gif":   "image/gif",
		"bmp":   "image/bmp",
		"svg":   "image/svg+xml",
		"mp3":   "audio/mpeg",
		"wav":   "audio/wav",
		"flac":  "audio/flac",
		"mp4":   "video/mp4",
		"mov":   "video/quicktime",
		"avi":   "video/x-msvideo",
		"mkv":   "video/x-matroska",
		"zip":   "application/zip",
		"gz":    "application/gzip",
		"tar":   "application/x-tar",
		"7z":    "application/x-7z-compressed",
		"exe":   "application/x-executable",
		"dll":   "application/x-dll",
		"bin":   "application/octet-stream",
		"sh":    "application/x-shellscript",
		"py":    "text/x-python",
		"rb":    "text/x-ruby",
		"rs":    "text/x-rust",
		"c":     "text/x-c",
		"cpp":   "text/x-c++",
		"h":     "text/x-c-header",
		"hpp":   "text/x-c++-header",
		"sql":   "application/sql",
		"xml":   "application/xml",
		"ico":   "image/x-icon",
		"woff":  "font/woff",
		"woff2": "font/woff2",
		"ttf":   "font/ttf",
		"otf":   "font/otf",
	}

	if mime, ok := mimeMap[ext]; ok {
		return mime
	}

	return "application/octet-stream"
}

// getFileExtension returns the file extension
func getFileExtension(filepath string) string {
	lastDot := -1
	for i := len(filepath) - 1; i >= 0; i-- {
		if filepath[i] == '.' {
			lastDot = i
			break
		}
	}

	if lastDot == -1 || lastDot == len(filepath)-1 {
		return ""
	}

	return filepath[lastDot+1:]
}

// IsTextFile checks if a file is text-based based on content type
func IsTextFile(contentType string) bool {
	textTypes := map[string]bool{
		"text/plain":                true,
		"text/markdown":             true,
		"text/x-go":                 true,
		"text/x-python":             true,
		"text/x-ruby":               true,
		"text/x-rust":               true,
		"text/x-c":                  true,
		"text/x-c++":                true,
		"text/x-c-header":           true,
		"text/html":                 true,
		"text/css":                  true,
		"application/json":          true,
		"application/yaml":          true,
		"application/xml":           true,
		"application/javascript":    true,
		"application/sql":           true,
		"application/x-shellscript": true,
	}
	return textTypes[contentType]
}
