package blockchain

import (
	"errors"
	"log"
	"sync"
)

// Blockchain represents the entire blockchain
type Blockchain struct {
	Blocks      []*Block    `json:"blocks"`
	Difficulty  int         `json:"difficulty"`
	PendingDocs []*Document `json:"pendingDocuments"`
	mu          sync.Mutex  // Mutex for concurrent access
}

// NewBlockchain creates a new blockchain with a genesis block
func NewBlockchain(difficulty int) *Blockchain {
	genesisBlock := CreateGenesisBlock(difficulty)
	return &Blockchain{
		Blocks:      []*Block{genesisBlock},
		Difficulty:  difficulty,
		PendingDocs: []*Document{},
	}
}

// AddDocument adds a new document to pending documents
func (bc *Blockchain) AddDocument(doc *Document) bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if VerifyDocument(doc) {
		bc.PendingDocs = append(bc.PendingDocs, doc)
		return true
	} 
	
	log.Println("Document signature verification failed")
	return false
}

// MineBlock creates a new block with pending documents
func (bc *Blockchain) MineBlock() (*Block, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if len(bc.PendingDocs) == 0 {
		return nil, errors.New("no pending documents to mine")
	}
	
	lastBlock := bc.Blocks[len(bc.Blocks)-1]
	
	newBlock := NewBlock(
		lastBlock.Index+1,
		bc.PendingDocs,
		lastBlock.Hash,
		bc.Difficulty,
	)
	
	// Perform proof of work
	newBlock.Mine()
	
	bc.Blocks = append(bc.Blocks, newBlock)
	bc.PendingDocs = []*Document{} // Clear pending documents
	
	return newBlock, nil
}

// GetLatestBlock returns the most recent block in the chain
func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

// IsChainValid checks if the blockchain is valid
func (bc *Blockchain) IsChainValid() bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		previousBlock := bc.Blocks[i-1]
		
		// Check if block has valid hash
		if !currentBlock.IsValid() {
			log.Printf("Block %d has invalid hash", currentBlock.Index)
			return false
		}
		
		// Check if the previous hash matches
		if currentBlock.PrevHash != previousBlock.Hash {
			log.Printf("Block %d has invalid previous hash", currentBlock.Index)
			return false
		}
		
		// Check if all documents in the block have valid signatures
		for _, doc := range currentBlock.Documents {
			if !VerifyDocument(doc) {
				log.Printf("Document %s in block %d has invalid signature", doc.ID, currentBlock.Index)
				return false
			}
		}
	}
	
	return true
}

// GetBlockByHash returns a block by its hash
func (bc *Blockchain) GetBlockByHash(hash string) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	for _, block := range bc.Blocks {
		if block.Hash == hash {
			return block
		}
	}
	
	return nil
}

// GetBlockByIndex returns a block by its index
func (bc *Blockchain) GetBlockByIndex(index int64) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	for _, block := range bc.Blocks {
		if block.Index == index {
			return block
		}
	}
	
	return nil
}

// GetDocumentByID returns a document by its ID
func (bc *Blockchain) GetDocumentByID(id string) *Document {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	// Check pending documents
	for _, doc := range bc.PendingDocs {
		if doc.ID == id {
			return doc
		}
	}
	
	// Check documents in blocks
	for _, block := range bc.Blocks {
		for _, doc := range block.Documents {
			if doc.ID == id {
				return doc
			}
		}
	}
	
	return nil
}