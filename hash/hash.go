package hash

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash/crc32"

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
	result.SHA256 = hex.EncodeToString(sha256Hash[:])

	// SHA-512
	sha512Hash := sha512.Sum512(content)
	result.SHA512 = hex.EncodeToString(sha512Hash[:])

	// MD5
	md5Hash := md5.Sum(content)
	result.MD5 = hex.EncodeToString(md5Hash[:])

	// BLAKE2
	blake2Hash := blake2b.Sum256(content)
	result.BLAKE2 = hex.EncodeToString(blake2Hash[:])

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

// Helper: Format CRC32 as 8-character hex string
func formatCRC32(crc uint32) string {
	hexStr := "0123456789abcdef"
	result := make([]byte, 8)

	result[0] = hexStr[(crc>>28)&0xf]
	result[1] = hexStr[(crc>>24)&0xf]
	result[2] = hexStr[(crc>>20)&0xf]
	result[3] = hexStr[(crc>>16)&0xf]
	result[4] = hexStr[(crc>>12)&0xf]
	result[5] = hexStr[(crc>>8)&0xf]
	result[6] = hexStr[(crc>>4)&0xf]
	result[7] = hexStr[crc&0xf]

	return string(result)
}

// Helper: Convert bytes to hex string
func ToHex(b []byte) string {
	return hex.EncodeToString(b)
}

// Helper: Convert hex string to bytes
func FromHex(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
