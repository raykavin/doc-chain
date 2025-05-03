package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DocumentRepository handles document operations with the database
type DocumentRepository struct {
	db *Database
}

// Document represents a document in the database
type Document struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	ContentType string     `json:"contentType"`
	Status      string     `json:"status"`
	SignedBy    string     `json:"signedBy,omitempty"`
	BlockIndex  *int64     `json:"blockIndex,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	SignedAt    *time.Time `json:"signedAt,omitempty"`
}

// Signature represents a signature in the database
type Signature struct {
	ID             string    `json:"id"`
	DocumentID     string    `json:"documentId"`
	WalletID       string    `json:"walletId"`
	SignatureHash  string    `json:"signatureHash"`
	CertificateID  string    `json:"certificateId"`
	HasTimestamp   bool      `json:"hasTimestamp"`
	BlockIndex     *int64    `json:"blockIndex,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *Database) *DocumentRepository {
	return &DocumentRepository{
		db: db,
	}
}

// CreateDocument creates a new document in the database
func (r *DocumentRepository) CreateDocument(title, contentType string, content []byte) (*Document, error) {
	// Generate a unique ID
	id := uuid.New().String()

	// Get the current time
	now := time.Now()

	// Insert the document into the database
	_, err := r.db.GetDB().Exec(
		"INSERT INTO documents (id, title, content_type, content, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, title, contentType, content, "pending", now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao inserir documento no banco de dados: %w", err)
	}

	// Return the document
	return &Document{
		ID:          id,
		Title:       title,
		ContentType: contentType,
		Status:      "pending",
		CreatedAt:   now,
	}, nil
}

// GetDocument gets a document from the database
func (r *DocumentRepository) GetDocument(id string) (*Document, error) {
	// Query the database
	row := r.db.GetDB().QueryRow(
		`SELECT d.id, d.title, d.content_type, d.status, s.wallet_id, s.block_index, d.created_at, s.created_at
		FROM documents d
		LEFT JOIN signatures s ON d.id = s.document_id
		WHERE d.id = ?`,
		id,
	)

	// Scan the result
	var doc Document
	var createdAtStr string
	var signedByStr, signedAtStr *string
	var blockIndex *int64
	err := row.Scan(&doc.ID, &doc.Title, &doc.ContentType, &doc.Status, &signedByStr, &blockIndex, &createdAtStr, &signedAtStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter documento: %w", err)
	}

	// Parse the created at time
	doc.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao analisar data de criação: %w", err)
	}

	// Set signed by if available
	if signedByStr != nil {
		doc.SignedBy = *signedByStr
	}

	// Set block index if available
	if blockIndex != nil {
		doc.BlockIndex = blockIndex
	}

	// Parse the signed at time if available
	if signedAtStr != nil {
		signedAt, err := time.Parse(time.RFC3339, *signedAtStr)
		if err != nil {
			return nil, fmt.Errorf("falha ao analisar data de assinatura: %w", err)
		}
		doc.SignedAt = &signedAt
	}

	return &doc, nil
}

// GetDocuments gets all documents from the database
func (r *DocumentRepository) GetDocuments() ([]*Document, error) {
	// Query the database
	rows, err := r.db.GetDB().Query(
		`SELECT d.id, d.title, d.content_type, d.status, s.wallet_id, s.block_index, d.created_at, s.created_at
		FROM documents d
		LEFT JOIN signatures s ON d.id = s.document_id
		ORDER BY d.created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter documentos: %w", err)
	}
	defer rows.Close()

	// Scan the results
	var docs []*Document
	for rows.Next() {
		var doc Document
		var createdAtStr string
		var signedByStr, signedAtStr *string
		var blockIndex *int64
		err := rows.Scan(&doc.ID, &doc.Title, &doc.ContentType, &doc.Status, &signedByStr, &blockIndex, &createdAtStr, &signedAtStr)
		if err != nil {
			return nil, fmt.Errorf("falha ao analisar documento: %w", err)
		}

		// Parse the created at time
		doc.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("falha ao analisar data de criação: %w", err)
		}

		// Set signed by if available
		if signedByStr != nil {
			doc.SignedBy = *signedByStr
		}

		// Set block index if available
		if blockIndex != nil {
			doc.BlockIndex = blockIndex
		}

		// Parse the signed at time if available
		if signedAtStr != nil {
			signedAt, err := time.Parse(time.RFC3339, *signedAtStr)
			if err != nil {
				return nil, fmt.Errorf("falha ao analisar data de assinatura: %w", err)
			}
			doc.SignedAt = &signedAt
		}

		docs = append(docs, &doc)
	}

	return docs, nil
}

// GetDocumentContent gets a document's content from the database
func (r *DocumentRepository) GetDocumentContent(id string) ([]byte, error) {
	// Query the database
	row := r.db.GetDB().QueryRow(
		"SELECT content FROM documents WHERE id = ?",
		id,
	)

	// Scan the result
	var content []byte
	err := row.Scan(&content)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter conteúdo do documento: %w", err)
	}

	return content, nil
}

// SignDocument signs a document in the database
func (r *DocumentRepository) SignDocument(documentID, walletID, signatureHash, certificateID string, signature []byte, includeTimestamp bool) (string, error) {
	// Generate a unique ID
	id := uuid.New().String()

	// Get the current time
	now := time.Now()

	// Start a transaction
	tx, err := r.db.GetDB().Begin()
	if err != nil {
		return "", fmt.Errorf("falha ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Insert the signature into the database
	_, err = tx.Exec(
		`INSERT INTO signatures (id, document_id, wallet_id, signature, signature_hash, certificate_hash, timestamp, has_timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, documentID, walletID, signature, signatureHash, certificateID, now.Format(time.RFC3339), includeTimestamp, now.Format(time.RFC3339),
	)
	if err != nil {
		return "", fmt.Errorf("falha ao inserir assinatura no banco de dados: %w", err)
	}

	// Update the document status
	_, err = tx.Exec(
		"UPDATE documents SET status = ? WHERE id = ?",
		"signed", documentID,
	)
	if err != nil {
		return "", fmt.Errorf("falha ao atualizar status do documento: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("falha ao confirmar transação: %w", err)
	}

	return id, nil
}

// UpdateDocumentBlockIndex updates a document's block index in the database
func (r *DocumentRepository) UpdateDocumentBlockIndex(documentID string, blockIndex int64) error {
	// Update the signature's block index
	_, err := r.db.GetDB().Exec(
		"UPDATE signatures SET block_index = ? WHERE document_id = ?",
		blockIndex, documentID,
	)
	if err != nil {
		return fmt.Errorf("falha ao atualizar índice de bloco da assinatura: %w", err)
	}

	return nil
}

// GetSignature gets a signature from the database
func (r *DocumentRepository) GetSignature(id string) (*Signature, error) {
	// Query the database
	row := r.db.GetDB().QueryRow(
		`SELECT id, document_id, wallet_id, signature_hash, certificate_hash, has_timestamp, block_index, created_at
		FROM signatures
		WHERE id = ?`,
		id,
	)

	// Scan the result
	var sig Signature
	var createdAtStr string
	err := row.Scan(&sig.ID, &sig.DocumentID, &sig.WalletID, &sig.SignatureHash, &sig.CertificateID, &sig.HasTimestamp, &sig.BlockIndex, &createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter assinatura: %w", err)
	}

	// Parse the created at time
	sig.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao analisar data de criação: %w", err)
	}

	return &sig, nil
}

// GetSignaturesByDocument gets all signatures for a document from the database
func (r *DocumentRepository) GetSignaturesByDocument(documentID string) ([]*Signature, error) {
	// Query the database
	rows, err := r.db.GetDB().Query(
		`SELECT id, document_id, wallet_id, signature_hash, certificate_hash, has_timestamp, block_index, created_at
		FROM signatures
		WHERE document_id = ?
		ORDER BY created_at DESC`,
		documentID,
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter assinaturas: %w", err)
	}
	defer rows.Close()

	// Scan the results
	var sigs []*Signature
	for rows.Next() {
		var sig Signature
		var createdAtStr string
		err := rows.Scan(&sig.ID, &sig.DocumentID, &sig.WalletID, &sig.SignatureHash, &sig.CertificateID, &sig.HasTimestamp, &sig.BlockIndex, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("falha ao analisar assinatura: %w", err)
		}

		// Parse the created at time
		sig.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("falha ao analisar data de criação: %w", err)
		}

		sigs = append(sigs, &sig)
	}

	return sigs, nil
}
