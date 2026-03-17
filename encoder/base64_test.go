package encoder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte{}, ""},
		{"simple", []byte("hello"), "aGVsbG8="},
		{"binary", []byte{0x00, 0x01, 0x02, 0x03}, "AAECAw=="},
		{"text", []byte("Hello, World!"), "SGVsbG8sIFdvcmxkIQ=="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Encode(tt.input)
			if result != tt.expected {
				t.Errorf("Encode(%v) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{"empty", "", []byte{}, false},
		{"simple", "aGVsbG8=", []byte("hello"), false},
		{"binary", "AAECAw==", []byte{0x00, 0x01, 0x02, 0x03}, false},
		{"invalid", "!!!invalid!!!", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalBytes(result, tt.want) {
				t.Errorf("Decode(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := []byte("Hello, World! This is a test.")
	encoded := Encode(original)
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if !equalBytes(decoded, original) {
		t.Errorf("Round trip failed: %v != %v", decoded, original)
	}
}

func TestEncodeDecode_BinaryData(t *testing.T) {
	// Test with various binary patterns
	testCases := [][]byte{
		{0x00, 0x00, 0x00, 0x00},             // All zeros
		{0xFF, 0xFF, 0xFF, 0xFF},             // All ones
		{0x00, 0xFF, 0x00, 0xFF},             // Alternating
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, // Sequential
		make([]byte, 1024),                   // 1KB of zeros
	}

	for i, original := range testCases {
		encoded := Encode(original)
		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Test case %d: Decode error: %v", i, err)
		}
		if !equalBytes(decoded, original) {
			t.Errorf("Test case %d: Round trip failed", i)
		}
	}
}

func TestEncodeDecode_Empty(t *testing.T) {
	original := []byte{}
	encoded := Encode(original)
	if encoded != "" {
		t.Errorf("Encode empty returned %q, expected empty string", encoded)
	}

	decoded, err := Decode("")
	if err != nil {
		t.Fatalf("Decode empty error: %v", err)
	}
	if !equalBytes(decoded, original) {
		t.Errorf("Decode empty returned %v, expected empty slice", decoded)
	}
}

func TestEncodeFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Hello, World!")
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	encoded, err := EncodeFile(testFile)
	if err != nil {
		t.Fatalf("EncodeFile error: %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if !equalBytes(decoded, content) {
		t.Errorf("EncodeFile round trip failed: %v != %v", decoded, content)
	}
}

func TestEncodeFile_NotFound(t *testing.T) {
	_, err := EncodeFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("EncodeFile should return error for non-existent file")
	}
}

func TestDecodeFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "output.txt")
	content := []byte("Test content for DecodeFile")
	encoded := Encode(content)

	err := DecodeFile(encoded, testFile, 0644)
	if err != nil {
		t.Fatalf("DecodeFile error: %v", err)
	}

	// Verify file was created
	decoded, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	if !equalBytes(decoded, content) {
		t.Errorf("DecodeFile wrote wrong content: %v != %v", decoded, content)
	}
}

func TestDecodeFile_InvalidBase64(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "output.txt")

	err := DecodeFile("!!!invalid!!!", testFile, 0644)
	if err == nil {
		t.Error("DecodeFile should return error for invalid base64")
	}
}

func TestEncodeDecode_LargeData(t *testing.T) {
	// Create 100KB of data
	original := make([]byte, 100*1024)
	for i := range original {
		original[i] = byte(i % 256)
	}

	encoded := Encode(original)
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if len(decoded) != len(original) {
		t.Errorf("Length mismatch: got %d, want %d", len(decoded), len(original))
	}

	if !equalBytes(decoded, original) {
		t.Error("Data mismatch in large data test")
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
