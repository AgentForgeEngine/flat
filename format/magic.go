package format

import (
	"io"
	"os"
	"strings"
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
	{Signature: []byte("PK"), MIMEType: "application/zip"},
	{Signature: []byte("RIFF"), MIMEType: "audio/wav"},
}

// Format markers
const (
	HeaderStart    = "!--~---~BEGIN-FLAT-FILE-MULTI~--~---!"
	HeaderEnd      = "!--~---~END-HEADER~--~---!"
	HeaderBegin    = "!--~---~BEGIN-HEADER~--~---!"
	FileBegin      = "!--~---~BEGIN-FILE~--~---!"
	FileEnd        = "!--~---~END-FILE~--~---!"
	MetadataBegin  = "!--~---~BEGIN-METADATA~--~---!"
	MetadataEnd    = "!--~---~END-METADATA~--~---!"
	ContentBegin   = "!--~---~BEGIN-FILE-CONTENT~--~---!"
	ContentEnd     = "!--~---~END-FILE-CONTENT~--~---!"
	DirectoryBegin = "!--~---~BEGIN-DIRECTORY~--~---!"
	DirectoryEnd   = "!--~---~END-DIRECTORY~--~---!"
)

// IsBinary detects if a file is binary using magic bytes and extension
func IsBinary(filepath string) (bool, string) {
	// First check file extension
	ext := strings.ToLower(getFileExtension(filepath))
	binaryExtensions := map[string]bool{
		"png": true, "jpg": true, "jpeg": true, "gif": true, "bmp": true,
		"mp3": true, "mp4": true, "mov": true, "avi": true, "mkv": true,
		"wav": true, "flac": true, "ogg": true, "wma": true,
		"exe": true, "dll": true, "bin": true, "iso": true, "img": true,
		"elf": true, "so": true, "dylib": true,
		"zip": true, "gz": true, "tar": true, "7z": true, "rar": true,
		"pdf": true, "doc": true, "docx": true, "xls": true, "xlsx": true,
		"ppt": true, "pptx": true, "psd": true, "ai": true, "eps": true,
		"woff": true, "woff2": true, "ttf": true, "otf": true, "eot": true,
		"swf": true, "flv": true, "webm": true,
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
		if n >= len(sig.Signature) {
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

// IsTextFile checks if a file is a text file
func IsTextFile(filepath string) bool {
	isBin, _ := IsBinary(filepath)
	return !isBin
}
