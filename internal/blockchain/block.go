package blockchain

import (
	"document-blockchain/internal/crypto"
	"time"
)

// Block represents each block in the blockchain
type Block struct {
	Index        int64       `json:"index"`
	Timestamp    int64       `json:"timestamp"`
	Documents    []*Document `json:"documents"`
	PrevHash     string      `json:"prevHash"`
	Hash         string      `json:"hash"`
	Nonce        int64       `json:"nonce"`
	Difficulty   int         `json:"difficulty"`
	MerkleRoot   string      `json:"merkleRoot"`
}

// NewBlock creates a new block with the specified data
func NewBlock(index int64, documents []*Document, prevHash string, difficulty int) *Block {
	block := &Block{
		Index:      index,
		Timestamp:  time.Now().Unix(),
		Documents:  documents,
		PrevHash:   prevHash,
		Difficulty: difficulty,
		Nonce:      0,
	}
	
	block.MerkleRoot = calculateMerkleRoot(documents)
	
	return block
}

// Mine performs proof of work to find a valid hash for the block
func (b *Block) Mine() {
	b.Nonce, b.Hash = crypto.ProofOfWork(
		b.Index, 
		b.Timestamp, 
		b.MerkleRoot, 
		b.PrevHash, 
		b.Difficulty,
	)
}

// CalculateHash calculates the hash of the block
func (b *Block) CalculateHash() string {
	return crypto.CalculateHash(
		b.Index,
		b.Timestamp,
		b.MerkleRoot,
		b.PrevHash,
		b.Nonce,
		b.Difficulty,
	)
}

// IsValid checks if a block's hash is valid
func (b *Block) IsValid() bool {
	return crypto.ValidateProofOfWork(
		b.Index,
		b.Timestamp,
		b.MerkleRoot,
		b.PrevHash,
		b.Nonce,
		b.Difficulty,
		b.Hash,
	)
}

// CreateGenesisBlock creates the first block in the blockchain
func CreateGenesisBlock(difficulty int) *Block {
	genesisDoc := CreateGenesisDocument()
	
	block := &Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Documents:    []*Document{genesisDoc},
		PrevHash:     "0",
		Difficulty:   difficulty,
		Nonce:        0,
	}
	
	block.MerkleRoot = calculateMerkleRoot(block.Documents)
	// Pre-calculate hash for genesis block
	block.Hash = block.CalculateHash()
	
	return block
}