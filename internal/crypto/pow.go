package crypto

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// ProofOfWork performs the mining operation to find a valid hash
// It takes block data and returns the nonce and resulting hash that satisfies the difficulty
func ProofOfWork(index int64, timestamp int64, merkleRoot string, prevHash string, difficulty int) (int64, string) {
	var nonce int64 = 0
	startTime := time.Now()
	log.Printf("Mining block %d...", index)
	
	for {
		hash := CalculateHash(index, timestamp, merkleRoot, prevHash, nonce, difficulty)
		
		if ValidateHash(hash, difficulty) {
			duration := time.Since(startTime)
			log.Printf("Block %d mined in %v. Hash: %s", index, duration, hash)
			return nonce, hash
		}
		
		nonce++
		
		// Every million attempts, log progress
		if nonce%1000000 == 0 {
			duration := time.Since(startTime)
			rate := float64(nonce) / duration.Seconds()
			log.Printf("Mining block %d... %d hashes (%.2f hashes/sec)", index, nonce, rate)
		}
	}
}

// ValidateProofOfWork verifies that a block's hash is valid
func ValidateProofOfWork(index int64, timestamp int64, merkleRoot string, prevHash string, nonce int64, difficulty int, hash string) bool {
	prefix := strings.Repeat("0", difficulty)
	
	// First check if the hash has the required prefix
	if !strings.HasPrefix(hash, prefix) {
		return false
	}
	
	// Then verify the hash calculation
	calculatedHash := CalculateHash(index, timestamp, merkleRoot, prevHash, nonce, difficulty)
	if calculatedHash != hash {
		fmt.Printf("Invalid hash: %s != %s\n", calculatedHash, hash)
		return false
	}
	
	return true
}