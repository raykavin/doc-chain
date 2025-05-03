package icpbrasil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/raykavin/doc-chain/internal/blockchain"
)

// DocumentManager manages ICP-Brasil document operations
type DocumentManager struct {
	Blockchain *blockchain.Blockchain
}

// NewDocumentManager creates a new document manager
func NewDocumentManager(bc *blockchain.Blockchain) *DocumentManager {
	return &DocumentManager{
		Blockchain: bc,
	}
}

// SignDocumentWithCertificate signs a document with an ICP-Brasil certificate
func (dm *DocumentManager) SignDocumentWithCertificate(
	documentPath string,
	certificatePath string,
	certificatePassword string,
	options *SignatureOptions,
) (*SignedDocument, error) {
	// Load the certificate
	cert, err := LoadPFXFromFile(certificatePath, certificatePassword)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	// Verify the certificate
	if err := cert.VerifyCertificate(); err != nil {
		return nil, fmt.Errorf("certificate verification failed: %w", err)
	}

	// Check if it's an ICP-Brasil certificate
	if !cert.IsICPBrasilCertificate() {
		return nil, fmt.Errorf("not an ICP-Brasil certificate")
	}

	// Sign the document
	signedDoc, err := cert.SignFile(documentPath, options)
	if err != nil {
		return nil, fmt.Errorf("failed to sign document: %w", err)
	}

	return signedDoc, nil
}

// SignAndAddToBlockchain signs a document and adds it to the blockchain
func (dm *DocumentManager) SignAndAddToBlockchain(
	documentPath string,
	certificatePath string,
	certificatePassword string,
	options *SignatureOptions,
) (*BlockchainDocument, error) {
	// Sign the document
	signedDoc, err := dm.SignDocumentWithCertificate(
		documentPath,
		certificatePath,
		certificatePassword,
		options,
	)
	if err != nil {
		return nil, err
	}

	// Create a blockchain document
	bcDoc := NewBlockchainDocument(signedDoc)

	// Add to blockchain
	_, err = bcDoc.AddToBlockchain(dm.Blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to add document to blockchain: %w", err)
	}

	return bcDoc, nil
}

// VerifySignature verifies a document signature
func (dm *DocumentManager) VerifySignature(
	documentPath string,
	signaturePath string,
) (*VerificationResult, error) {
	// Read the document
	document, err := ioutil.ReadFile(documentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read document: %w", err)
	}

	// Read the signature
	signature, err := ioutil.ReadFile(signaturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signature: %w", err)
	}

	// Verify the signature
	return VerifyDetachedSignature(document, signature)
}

// VerifySignatureAndBlockchain verifies a document signature and checks the blockchain
func (dm *DocumentManager) VerifySignatureAndBlockchain(
	documentPath string,
	signaturePath string,
) (bool, *VerificationResult, error) {
	// Verify the signature
	result, err := dm.VerifySignature(documentPath, signaturePath)
	if err != nil {
		return false, nil, err
	}

	// If the signature is not valid, return early
	if !result.IsValid {
		return false, result, nil
	}

	// Read the document
	document, err := ioutil.ReadFile(documentPath)
	if err != nil {
		return false, result, fmt.Errorf("failed to read document: %w", err)
	}

	// Read the signature
	signature, err := ioutil.ReadFile(signaturePath)
	if err != nil {
		return false, result, fmt.Errorf("failed to read signature: %w", err)
	}

	// Create a signed document
	signedDoc := &SignedDocument{
		OriginalContent: document,
		Signature:       signature,
		Certificate: &Certificate{
			Cert: result.SignerCert,
		},
		SignatureTime: result.SignatureTime,
	}

	// Verify against the blockchain
	bcValid, err := VerifyBlockchainDocument(dm.Blockchain, signedDoc)
	if err != nil {
		// Document might be valid but not in the blockchain yet
		return false, result, fmt.Errorf("blockchain verification failed: %w", err)
	}

	return bcValid, result, nil
}

// SaveSignedDocument saves a signed document and its signature
func (dm *DocumentManager) SaveSignedDocument(
	signedDoc *SignedDocument,
	outputDir string,
	baseFilename string,
) (string, string, error) {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save the original document
	documentPath := filepath.Join(outputDir, baseFilename)
	if err := ioutil.WriteFile(documentPath, signedDoc.OriginalContent, 0644); err != nil {
		return "", "", fmt.Errorf("failed to save document: %w", err)
	}

	// Save the signature
	signaturePath := filepath.Join(outputDir, baseFilename+".p7s")
	if err := ioutil.WriteFile(signaturePath, signedDoc.Signature, 0644); err != nil {
		return "", "", fmt.Errorf("failed to save signature: %w", err)
	}

	return documentPath, signaturePath, nil
}

// GetDocumentInfo returns information about a signed document
func (dm *DocumentManager) GetDocumentInfo(
	documentPath string,
	signaturePath string,
) (map[string]string, error) {
	// Verify the signature
	result, err := dm.VerifySignature(documentPath, signaturePath)
	if err != nil {
		return nil, err
	}

	// Read the document
	document, err := ioutil.ReadFile(documentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read document: %w", err)
	}

	// Read the signature
	signature, err := ioutil.ReadFile(signaturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signature: %w", err)
	}

	// Create a signed document
	signedDoc := &SignedDocument{
		OriginalContent: document,
		Signature:       signature,
		Certificate: &Certificate{
			Cert: result.SignerCert,
		},
		SignatureTime: result.SignatureTime,
	}

	// Get document information
	info := make(map[string]string)
	info["DocumentHash"] = signedDoc.GetDocumentHash()
	info["SignatureHash"] = signedDoc.GetSignatureHash()
	info["SignerCertificateHash"] = signedDoc.GetSignerCertificateHash()
	info["SignatureValid"] = fmt.Sprintf("%t", result.IsValid)
	info["SignerName"] = result.SignerCert.Subject.CommonName
	info["SignerIssuer"] = result.SignerCert.Issuer.CommonName
	info["SignatureTime"] = result.SignatureTime.Format(time.RFC3339)

	// Try to get blockchain information
	bcInfo, err := GetBlockchainInfo(dm.Blockchain, info["DocumentHash"])
	if err == nil {
		info["InBlockchain"] = "true"
		if blockInfo, ok := bcInfo["Block"]; ok {
			info["BlockInfo"] = blockInfo
		} else {
			info["BlockInfo"] = "Pending (not yet in a block)"
		}
	} else {
		info["InBlockchain"] = "false"
	}

	return info, nil
}

// CheckRevocation checks if a certificate is revoked
func (dm *DocumentManager) CheckRevocation(
	certificatePath string,
	certificatePassword string,
) (bool, error) {
	// Load the certificate
	cert, err := LoadPFXFromFile(certificatePath, certificatePassword)
	if err != nil {
		return false, fmt.Errorf("failed to load certificate: %w", err)
	}

	// Try CRL first
	crlErr := cert.CheckRevocationCRL()
	if crlErr == nil {
		// Certificate is not revoked according to CRL
		return false, nil
	}

	// Try OCSP if CRL failed
	ocspErr := cert.CheckRevocationOCSP()
	if ocspErr == nil {
		// Certificate is not revoked according to OCSP
		return false, nil
	}

	// Check if either error indicates that the certificate is revoked
	if crlErr.Error()[:22] == "certificate is revoked" ||
		ocspErr.Error()[:22] == "certificate is revoked" {
		return true, nil
	}
	// Both methods failed for other reasons
	return false, nil

}
