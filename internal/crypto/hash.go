package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// GenerateID generates a unique ID for a document
func GenerateID(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// CalculateHash calculates the hash of a block's data
// This function takes the components of a block and returns a SHA-256 hash
func CalculateHash(index int64, timestamp int64, merkleRoot string, prevHash string, nonce int64, difficulty int) string {
	record := strconv.FormatInt(index, 10) + 
		strconv.FormatInt(timestamp, 10) + 
		merkleRoot + 
		prevHash + 
		strconv.FormatInt(nonce, 10) + 
		strconv.Itoa(difficulty)
		
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// ValidateHash checks if a hash meets the required difficulty
func ValidateHash(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

// CalculateHashFromString calculates a SHA-256 hash from a string
func CalculateHashFromString(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CalculateDoubleHash calculates a double SHA-256 hash (hash of a hash)
func CalculateDoubleHash(data string) string {
	firstHash := sha256.Sum256([]byte(data))
	secondHash := sha256.Sum256(firstHash[:])
	return hex.EncodeToString(secondHash[:])
}

// HashBlocks takes two hashes and combines them
// This is used in Merkle tree construction
func HashBlocks(left, right string) string {
	combined := left + right
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}