package hash

import (
	"strings"
	"testing"
)

func TestComputeAllHashes(t *testing.T) {
	content := []byte("Hello, World!")
	result := ComputeAllHashes(content)

	if result == nil {
		t.Fatal("ComputeAllHashes returned nil")
	}

	// Check all hashes are non-empty
	if result.SHA256 == "" {
		t.Error("SHA256 hash is empty")
	}
	if result.SHA512 == "" {
		t.Error("SHA512 hash is empty")
	}
	if result.MD5 == "" {
		t.Error("MD5 hash is empty")
	}
	if result.BLAKE2 == "" {
		t.Error("BLAKE2 hash is empty")
	}
	if result.CRC32 == "" {
		t.Error("CRC32 hash is empty")
	}

	// Check SHA256 length (64 hex characters for 32 bytes)
	if len(result.SHA256) != 64 {
		t.Errorf("SHA256 length is %d, expected 64", len(result.SHA256))
	}

	// Check SHA512 length (128 hex characters for 64 bytes)
	if len(result.SHA512) != 128 {
		t.Errorf("SHA512 length is %d, expected 128", len(result.SHA512))
	}

	// Check MD5 length (32 hex characters for 16 bytes)
	if len(result.MD5) != 32 {
		t.Errorf("MD5 length is %d, expected 32", len(result.MD5))
	}

	// Check BLAKE2 length (64 hex characters for 32 bytes)
	if len(result.BLAKE2) != 64 {
		t.Errorf("BLAKE2 length is %d, expected 64", len(result.BLAKE2))
	}

	// Check CRC32 length (8 hex characters for 4 bytes)
	if len(result.CRC32) != 8 {
		t.Errorf("CRC32 length is %d, expected 8", len(result.CRC32))
	}
}

func TestComputeAllHashes_EmptyContent(t *testing.T) {
	content := []byte{}
	result := ComputeAllHashes(content)

	if result == nil {
		t.Fatal("ComputeAllHashes returned nil for empty content")
	}

	// Empty content should still produce valid hashes
	if result.SHA256 == "" {
		t.Error("SHA256 hash is empty for empty content")
	}
}

func TestComputeAllHashes_Deterministic(t *testing.T) {
	content := []byte("test content")
	result1 := ComputeAllHashes(content)
	result2 := ComputeAllHashes(content)

	if result1.SHA256 != result2.SHA256 {
		t.Error("SHA256 hashes are not deterministic")
	}
	if result1.MD5 != result2.MD5 {
		t.Error("MD5 hashes are not deterministic")
	}
}

func TestComputeAllHashes_VariousSizes(t *testing.T) {
	testCases := []struct {
		name string
		size int
	}{
		{"1 byte", 1},
		{"10 bytes", 10},
		{"100 bytes", 100},
		{"1KB", 1024},
		{"10KB", 10240},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := make([]byte, tc.size)
			for i := range content {
				content[i] = byte(i % 256)
			}

			result := ComputeAllHashes(content)
			if result == nil {
				t.Fatal("ComputeAllHashes returned nil")
			}

			// All hashes should be computed
			if result.SHA256 == "" || result.SHA512 == "" || result.MD5 == "" ||
				result.BLAKE2 == "" || result.CRC32 == "" {
				t.Error("Some hashes are empty")
			}
		})
	}
}

func TestVerifySHA256(t *testing.T) {
	content := []byte("test data")
	expectedSHA256 := ComputeAllHashes(content).SHA256

	// Test with correct hash
	if !VerifySHA256(content, expectedSHA256) {
		t.Error("VerifySHA256 returned false for correct hash")
	}

	// Test with incorrect hash
	wrongHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if VerifySHA256(content, wrongHash) {
		t.Error("VerifySHA256 returned true for wrong hash")
	}
}

func TestToHex(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte{}, ""},
		{"single byte", []byte{0x42}, "42"},
		{"multiple bytes", []byte{0xFF, 0x00, 0xAA}, "ff00aa"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToHex(tt.input)
			if result != tt.expected {
				t.Errorf("ToHex(%v) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFromHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{"empty", "", []byte{}, false},
		{"valid hex", "42ff00", []byte{0x42, 0xff, 0x00}, false},
		{"invalid hex", "zzzz", nil, true},
		{"odd length", "abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromHex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromHex(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Errorf("FromHex(%q) returned nil", tt.input)
			}
			if !tt.wantErr && !equalBytes(result, tt.want) {
				t.Errorf("FromHex(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
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

func TestComputeMDXBlockHash(t *testing.T) {
	yamlContent := "path: test/file.txt\nmode: 0644\n"
	result := ComputeMDXBlockHash(yamlContent)

	if result == nil {
		t.Fatal("ComputeMDXBlockHash returned nil")
	}

	// Should produce valid hashes
	if result.SHA256 == "" {
		t.Error("SHA256 hash is empty")
	}

	// Same input should produce same hash
	result2 := ComputeMDXBlockHash(yamlContent)
	if result.SHA256 != result2.SHA256 {
		t.Error("MDX block hash is not deterministic")
	}
}

func TestHash_Alphabetical(t *testing.T) {
	content := []byte("test")
	result := ComputeAllHashes(content)

	// Verify all hashes are lowercase hex
	if strings.ToLower(result.SHA256) != result.SHA256 {
		t.Error("SHA256 hash contains uppercase letters")
	}
	if strings.ToLower(result.SHA512) != result.SHA512 {
		t.Error("SHA512 hash contains uppercase letters")
	}
	if strings.ToLower(result.MD5) != result.MD5 {
		t.Error("MD5 hash contains uppercase letters")
	}
	if strings.ToLower(result.BLAKE2) != result.BLAKE2 {
		t.Error("BLAKE2 hash contains uppercase letters")
	}
}
