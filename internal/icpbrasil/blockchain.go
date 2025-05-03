package icpbrasil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/raykavin/doc-chain/internal/blockchain"
	"github.com/raykavin/doc-chain/internal/crypto"
)

// BlockchainDocument represents a document to be stored in the blockchain
type BlockchainDocument struct {
	// Original document hash
	DocumentHash string

	// Signature hash
	SignatureHash string

	// Signer certificate hash
	CertificateHash string

	// Signer information
	SignerName string
	SignerID   string

	// Timestamp information
	SignatureTime time.Time
	TimestampTime time.Time
	HasTimestamp  bool

	// Original signed document
	SignedDocument *SignedDocument
}

// NewBlockchainDocument creates a new blockchain document from a signed document
func NewBlockchainDocument(signedDoc *SignedDocument) *BlockchainDocument {
	cert := signedDoc.Certificate.Cert
	
	// Extract signer information from the certificate
	signerName := cert.Subject.CommonName
	signerID := ""
	
	// Look for ICP-Brasil specific OIDs in the certificate
	// OID 2.16.76.1.3.1 contains the CPF (Brazilian tax ID)
	for _, ext := range cert.Extensions {
		if ext.Id.String() == "2.16.76.1.3.1" {
			// Extract CPF from the extension value
			// Format: XXXXXXXXXXXXXXX (first 11 digits are the CPF)
			if len(ext.Value) >= 11 {
				signerID = string(ext.Value[:11])
			}
			break
		}
	}

	return &BlockchainDocument{
		DocumentHash:    signedDoc.GetDocumentHash(),
		SignatureHash:   signedDoc.GetSignatureHash(),
		CertificateHash: signedDoc.GetSignerCertificateHash(),
		SignerName:      signerName,
		SignerID:        signerID,
		SignatureTime:   signedDoc.SignatureTime,
		TimestampTime:   signedDoc.TimestampTime,
		HasTimestamp:    signedDoc.HasTimestamp,
		SignedDocument:  signedDoc,
	}
}

// AddToBlockchain adds the document to the blockchain
func (bd *BlockchainDocument) AddToBlockchain(bc *blockchain.Blockchain) (*blockchain.Document, error) {
	// Create a content string with all the document information
	content := fmt.Sprintf("%s:%s:%s:%s:%s:%d",
		bd.DocumentHash,
		bd.SignatureHash,
		bd.CertificateHash,
		bd.SignerName,
		bd.SignerID,
		bd.SignatureTime.Unix(),
	)

	// Create a signature for the blockchain
	// This is different from the document signature
	// It's just a wrapper to fit into the existing blockchain model
	signature := &crypto.Signature{
		R: crypto.NewBigInt(0),
		S: crypto.NewBigInt(0),
	}

	// Create a blockchain document
	doc := blockchain.NewDocument(
		content,
		signature,
		bd.CertificateHash, // Use the certificate hash as the signer ID
	)

	// Add the document to the blockchain
	if !bc.AddDocument(doc) {
		return nil, fmt.Errorf("failed to add document to blockchain")
	}

	return doc, nil
}

// VerifyBlockchainDocument verifies a document against the blockchain
func VerifyBlockchainDocument(bc *blockchain.Blockchain, signedDoc *SignedDocument) (bool, error) {
	// Create a blockchain document from the signed document
	bd := NewBlockchainDocument(signedDoc)

	// Get the document hash
	docHash := bd.DocumentHash

	// Find the document in the blockchain
	bcDoc := bc.GetDocumentByID(docHash)
	if bcDoc == nil {
		return false, fmt.Errorf("document not found in blockchain")
	}

	// Extract the document information from the blockchain document
	content := bcDoc.Content
	
	// Verify that the content contains the expected hashes
	expectedContent := fmt.Sprintf("%s:%s:%s",
		bd.DocumentHash,
		bd.SignatureHash,
		bd.CertificateHash,
	)

	// Check if the blockchain content contains the expected content
	// We use Contains instead of exact match because the blockchain content
	// might have additional information
	if content[:len(expectedContent)] != expectedContent {
		return false, fmt.Errorf("document hash mismatch")
	}

	return true, nil
}

// GenerateDocumentID generates a unique ID for a document
func GenerateDocumentID(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// GetBlockchainInfo returns information about a document in the blockchain
func GetBlockchainInfo(bc *blockchain.Blockchain, documentID string) (map[string]string, error) {
	// Find the document in the blockchain
	doc := bc.GetDocumentByID(documentID)
	if doc == nil {
		return nil, fmt.Errorf("document not found in blockchain")
	}

	// Extract information from the document
	info := make(map[string]string)
	info["DocumentID"] = doc.ID
	info["Content"] = doc.Content
	info["SignedBy"] = doc.SignedBy
	info["Timestamp"] = fmt.Sprintf("%d", doc.Timestamp)

	// Find the block containing the document
	var blockInfo string
	for _, block := range bc.Blocks {
		for _, blockDoc := range block.Documents {
			if blockDoc.ID == documentID {
				blockInfo = fmt.Sprintf("Block #%d (Hash: %s)", block.Index, block.Hash)
				break
			}
		}
		if blockInfo != "" {
			break
		}
	}

	if blockInfo != "" {
		info["Block"] = blockInfo
	} else {
		info["Status"] = "Pending (not yet in a block)"
	}

	return info, nil
}
