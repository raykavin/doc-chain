package database

import (
	"database/sql"
	"fmt"
	"time"
)

// BlockchainRepository handles database operations for blockchain
type BlockchainRepository struct {
	db *sql.DB
}

// Block represents a block entity
type Block struct {
	Index       int       `json:"index"`
	Hash        string    `json:"hash"`
	PrevHash    string    `json:"prevHash"`
	MerkleRoot  string    `json:"merkleRoot"`
	Timestamp   time.Time `json:"timestamp"`
	Nonce       int       `json:"nonce"`
	Difficulty  int       `json:"difficulty"`
	Documents   []*Document `json:"documents,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// NewBlockchainRepository creates a new blockchain repository
func NewBlockchainRepository(database *Database) *BlockchainRepository {
	return &BlockchainRepository{
		db: database.GetDB(),
	}
}

// CreateBlock creates a new block
func (r *BlockchainRepository) CreateBlock(index int, hash, prevHash, merkleRoot string, timestamp time.Time, nonce, difficulty int) (*Block, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	_, err = tx.Exec(
		"INSERT INTO blocks (index, hash, prev_hash, merkle_root, timestamp, nonce, difficulty, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		index, hash, prevHash, merkleRoot, timestamp, nonce, difficulty, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create block: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &Block{
		Index:      index,
		Hash:       hash,
		PrevHash:   prevHash,
		MerkleRoot: merkleRoot,
		Timestamp:  timestamp,
		Nonce:      nonce,
		Difficulty: difficulty,
		CreatedAt:  now,
	}, nil
}

// GetBlock gets a block by index
func (r *BlockchainRepository) GetBlock(index int) (*Block, error) {
	var block Block
	var timestamp, createdAt string

	err := r.db.QueryRow(
		"SELECT index, hash, prev_hash, merkle_root, timestamp, nonce, difficulty, created_at FROM blocks WHERE index = ?",
		index,
	).Scan(&block.Index, &block.Hash, &block.PrevHash, &block.MerkleRoot, &timestamp, &block.Nonce, &block.Difficulty, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("block not found: %d", index)
		}
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	// Parse timestamps
	block.Timestamp, err = time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	block.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	// Get documents in this block
	documents, err := r.GetBlockDocuments(index)
	if err != nil {
		return nil, fmt.Errorf("failed to get block documents: %w", err)
	}
	block.Documents = documents

	return &block, nil
}

// GetBlocks gets all blocks
func (r *BlockchainRepository) GetBlocks() ([]*Block, error) {
	rows, err := r.db.Query(
		"SELECT index, hash, prev_hash, merkle_root, timestamp, nonce, difficulty, created_at FROM blocks ORDER BY index DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks: %w", err)
	}
	defer rows.Close()

	blocks := []*Block{}
	for rows.Next() {
		var block Block
		var timestamp, createdAt string

		err := rows.Scan(&block.Index, &block.Hash, &block.PrevHash, &block.MerkleRoot, &timestamp, &block.Nonce, &block.Difficulty, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan block: %w", err)
		}

		// Parse timestamps
		block.Timestamp, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}

		block.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		blocks = append(blocks, &block)
	}

	return blocks, nil
}

// GetLatestBlock gets the latest block
func (r *BlockchainRepository) GetLatestBlock() (*Block, error) {
	var block Block
	var timestamp, createdAt string

	err := r.db.QueryRow(
		"SELECT index, hash, prev_hash, merkle_root, timestamp, nonce, difficulty, created_at FROM blocks ORDER BY index DESC LIMIT 1",
	).Scan(&block.Index, &block.Hash, &block.PrevHash, &block.MerkleRoot, &timestamp, &block.Nonce, &block.Difficulty, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no blocks found")
		}
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	// Parse timestamps
	block.Timestamp, err = time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	block.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return &block, nil
}

// AddDocumentToBlock adds a document to a block
func (r *BlockchainRepository) AddDocumentToBlock(blockIndex int, documentID string) error {
	_, err := r.db.Exec(
		"INSERT INTO block_documents (block_index, document_id) VALUES (?, ?)",
		blockIndex, documentID,
	)
	if err != nil {
		return fmt.Errorf("failed to add document to block: %w", err)
	}
	return nil
}

// GetBlockDocuments gets all documents in a block
func (r *BlockchainRepository) GetBlockDocuments(blockIndex int) ([]*Document, error) {
	rows, err := r.db.Query(`
		SELECT d.id, d.title, d.content_type, d.status, d.created_at
		FROM documents d
		JOIN block_documents bd ON d.id = bd.document_id
		WHERE bd.block_index = ?
		ORDER BY d.created_at DESC
	`, blockIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get block documents: %w", err)
	}
	defer rows.Close()

	documents := []*Document{}
	for rows.Next() {
		var doc Document
		var createdAt string

		err := rows.Scan(&doc.ID, &doc.Title, &doc.ContentType, &doc.Status, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}

		// Parse created_at
		doc.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		documents = append(documents, &doc)
	}

	return documents, nil
}

// GetBlockCount gets the total number of blocks
func (r *BlockchainRepository) GetBlockCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get block count: %w", err)
	}
	return count, nil
}

// GetPendingDocumentsCount gets the count of documents that are not yet in a block
func (r *BlockchainRepository) GetPendingDocumentsCount() (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(DISTINCT s.document_id)
		FROM signatures s
		WHERE s.block_index IS NULL
	`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending documents count: %w", err)
	}
	return count, nil
}
