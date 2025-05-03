package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/raykavin/doc-chain/internal/crypto"
)

// Document is a signed document in the blockchain
type Document struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Signature *crypto.Signature `json:"signature"`
	SignedBy  string            `json:"signedBy"` // Public key of the signer
	Timestamp int64             `json:"timestamp"`
}

// NewDocument creates a new document with the provided content
func NewDocument(content string, signature *crypto.Signature, signedBy string) *Document {
	contentHash := sha256.Sum256([]byte(content))

	return &Document{
		ID:        generateDocumentID(content),
		Content:   hex.EncodeToString(contentHash[:]),
		Signature: signature,
		SignedBy:  signedBy,
		Timestamp: time.Now().Unix(),
	}
}

// VerifyDocument verifies if the document's signature is valid
func VerifyDocument(doc *Document) bool {
	if doc == nil || doc.Signature == nil {
		return false
	}
	return crypto.VerifySignatureHex(doc.Content, doc.Signature, doc.SignedBy)
}

// generateDocumentID creates a unique ID for a document
func generateDocumentID(content string) string {
	return crypto.GenerateID(content)
}

// GetContentHash returns the document's content as a byte array
func (d *Document) GetContentHash() ([]byte, error) {
	return hex.DecodeString(d.Content)
}

// GetContentHashString returns the document's content as a string
func (d *Document) GetContentHashString() string {
	return d.Content
}

// CreateGenesisDocument creates a genesis document for the blockchain
func CreateGenesisDocument() *Document {
	return &Document{
		ID:      "genesis",
		Content: "genesis_document",
		Signature: &crypto.Signature{
			R: crypto.NewBigInt(0),
			S: crypto.NewBigInt(0),
		},
		SignedBy:  "system",
		Timestamp: time.Now().Unix(),
	}
}
