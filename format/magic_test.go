package format

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsBinary_Extensions(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		expected bool
	}{
		// Image files
		{"PNG", "image.png", true},
		{"JPEG", "photo.jpg", true},
		{"GIF", "animation.gif", true},
		{"BMP", "image.bmp", true},

		// Video files
		{"MP4", "video.mp4", true},
		{"MOV", "movie.mov", true},
		{"AVI", "clip.avi", true},
		{"MKV", "film.mkv", true},

		// Audio files
		{"MP3", "song.mp3", true},
		{"WAV", "audio.wav", true},

		// Archive files
		{"ZIP", "archive.zip", true},
		{"GZ", "backup.gz", true},
		{"TAR", "archive.tar", true},
		{"7Z", "archive.7z", true},

		// Executable files
		{"EXE", "program.exe", true},
		{"DLL", "library.dll", true},
		{"BIN", "data.bin", true},

		// Text files (should be false)
		{"TXT", "readme.txt", false},
		{"MD", "README.md", false},
		{"GO", "main.go", false},
		{"JS", "script.js", false},
		{"JSON", "config.json", false},
		{"YAML", "config.yaml", false},
		{"HTML", "index.html", false},
		{"CSS", "style.css", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isBin, _ := IsBinary(tc.filename)
			if isBin != tc.expected {
				t.Errorf("IsBinary(%q) = %v, expected %v", tc.filename, isBin, tc.expected)
			}
		})
	}
}

func TestIsTextFile(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		expected bool
	}{
		{"TXT", "file.txt", true},
		{"MD", "README.md", true},
		{"GO", "main.go", true},
		{"PNG", "image.png", false},
		{"EXE", "program.exe", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isText := IsTextFile(tc.filename)
			if isText != tc.expected {
				t.Errorf("IsTextFile(%q) = %v, expected %v", tc.filename, isText, tc.expected)
			}
		})
	}
}

func TestIsBinary_WithRealFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a text file
	textFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(textFile, []byte("hello world"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a fake PNG file (with PNG magic bytes)
	pngFile := filepath.Join(tmpDir, "image.png")
	err = os.WriteFile(pngFile, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, 0644)
	if err != nil {
		t.Fatalf("Failed to create PNG file: %v", err)
	}

	// Test text file detection
	isBin, _ := IsBinary(textFile)
	if isBin {
		t.Error("Text file should not be detected as binary")
	}

	// Test PNG detection
	isBin, mimeType := IsBinary(pngFile)
	if !isBin {
		t.Error("PNG file should be detected as binary")
	}
	if mimeType != "binary" && mimeType != "image/png" {
		t.Errorf("MIME type should be 'binary' or 'image/png', got %q", mimeType)
	}
}

func TestGetFileExtension(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		expected string
	}{
		{"simple", "file.txt", "txt"},
		{"multiple dots", "file.name.txt", "txt"},
		{"no extension", "file", ""},
		{"hidden file", ".gitignore", "gitignore"},
		{"extension at start", ".config", "config"},
		{"empty string", "", ""},
		{"just dot", ".", ""},
		{"directory", "path/to/dir", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getFileExtension(tc.filename)
			if result != tc.expected {
				t.Errorf("getFileExtension(%q) = %q, expected %q", tc.filename, result, tc.expected)
			}
		})
	}
}

func TestIsBinary_MagicBytes(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		filename string
		content  []byte
		expected bool
	}{
		{"PNG magic", "test.png", []byte{0x89, 0x50, 0x4E, 0x47}, true},
		{"JPEG magic", "test.jpg", []byte{0xFF, 0xD8, 0xFF}, true},
		{"ZIP magic", "test.zip", []byte{0x50, 0x4B, 0x03, 0x04}, true},
		{"ELF magic", "test.elf", []byte{0x7F, 0x45, 0x4C, 0x46}, true},
		{"Gzip magic", "test.gz", []byte{0x1F, 0x8B}, true},
		{"Text file", "test.txt", []byte("hello"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filepath := filepath.Join(tmpDir, tc.filename)
			err := os.WriteFile(filepath, tc.content, 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			isBin, _ := IsBinary(filepath)
			if isBin != tc.expected {
				t.Errorf("IsBinary(%q) = %v, expected %v", tc.filename, isBin, tc.expected)
			}
		})
	}
}

func TestIsBinary_ExtensionPriority(t *testing.T) {
	// Test that extension is checked first (before magic bytes)
	tmpDir := t.TempDir()

	// Create a file with .txt extension but PNG magic bytes
	// Extension check happens first, so it should NOT be detected as binary
	pngAsTxt := filepath.Join(tmpDir, "image.txt")
	err := os.WriteFile(pngAsTxt, []byte{0x89, 0x50, 0x4E, 0x47}, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should NOT be detected as binary due to .txt extension (extension checked first)
	isBin, _ := IsBinary(pngAsTxt)
	if isBin {
		t.Error("File with .txt extension should NOT be detected as binary even with PNG magic bytes")
	}

	// Create a file with .png extension but text content
	// Should be detected as binary due to .png extension
	txtAsPng := filepath.Join(tmpDir, "text.png")
	err = os.WriteFile(txtAsPng, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should be detected as binary due to .png extension
	isBin, _ = IsBinary(txtAsPng)
	if !isBin {
		t.Error("File with .png extension should be detected as binary even with text content")
	}
}

func TestIsBinary_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty file with various extensions
	testCases := []struct {
		name     string
		filename string
	}{
		{"empty.txt", "empty.txt"},
		{"empty.png", "empty.png"},
		{"empty.bin", "empty.bin"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filepath := filepath.Join(tmpDir, tc.filename)
			err := os.WriteFile(filepath, []byte{}, 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			isBin, _ := IsBinary(filepath)
			// Empty files with binary extensions should be detected as binary
			if tc.filename[len(tc.filename)-3:] != "txt" && tc.filename[len(tc.filename)-3:] != "md" &&
				tc.filename[len(tc.filename)-3:] != "go" && tc.filename[len(tc.filename)-3:] != "js" &&
				tc.filename[len(tc.filename)-3:] != "json" && tc.filename[len(tc.filename)-3:] != "yaml" {
				if !isBin {
					t.Errorf("Empty file with binary extension should be detected as binary")
				}
			}
		})
	}
}
