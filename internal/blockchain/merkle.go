package blockchain

import (
	"document-blockchain/internal/crypto"
)

// calculateMerkleRoot calculates the Merkle root of a list of documents
func calculateMerkleRoot(documents []*Document) string {
	if len(documents) == 0 {
		return ""
	}
	
	if len(documents) == 1 {
		return documents[0].Content
	}
	
	// Extract document hashes
	hashes := make([]string, len(documents))
	for i, doc := range documents {
		hashes[i] = doc.Content
	}
	
	// Keep combining hashes until we have just one (the root)
	for len(hashes) > 1 {
		// If odd number of hashes, duplicate the last one
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}
		
		// Combine pairs of hashes
		tempHashes := make([]string, 0, len(hashes)/2)
		for i := 0; i < len(hashes); i += 2 {
			combined := crypto.HashBlocks(hashes[i], hashes[i+1])
			tempHashes = append(tempHashes, combined)
		}
		
		hashes = tempHashes
	}
	
	return hashes[0]
}

// VerifyMerkleRoot verifies that a document is included in a Merkle tree
func VerifyMerkleRoot(document *Document, documents []*Document, merkleRoot string) bool {
	calculatedRoot := calculateMerkleRoot(documents)
	return calculatedRoot == merkleRoot
}