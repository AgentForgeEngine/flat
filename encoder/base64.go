package encoder

import (
	"encoding/base64"
	"os"
)

// TextMIMETypes lists MIME types that should not be base64 encoded
var TextMIMETypes = map[string]bool{
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

// Encode encodes binary content to base64
func Encode(content []byte) string {
	return base64.StdEncoding.EncodeToString(content)
}

// EncodeContent encodes content based on content type - text files are not encoded
func EncodeContent(content []byte, contentType string) string {
	if TextMIMETypes[contentType] {
		return string(content)
	}
	return Encode(content)
}

// Decode decodes base64 content to binary
func Decode(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// DecodeContent decodes content based on content type - text files are returned as-is
func DecodeContent(encoded string, contentType string) ([]byte, error) {
	if TextMIMETypes[contentType] {
		return []byte(encoded), nil
	}
	return Decode(encoded)
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
