package encoder

import (
	"encoding/base64"
	"os"
)

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
