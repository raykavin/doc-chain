package icpbrasil

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/digitorus/timestamp"
	"go.mozilla.org/pkcs7"
)

// SignatureOptions contains options for creating a digital signature
type SignatureOptions struct {
	// Detached indicates whether to create a detached signature
	Detached bool

	// IncludeChain indicates whether to include the certificate chain
	IncludeChain bool

	// AddTimestamp indicates whether to add a timestamp
	AddTimestamp bool

	// TSAUrl is the URL of the Time Stamping Authority
	TSAUrl string
}

// DefaultSignatureOptions returns the default signature options
func DefaultSignatureOptions() *SignatureOptions {
	return &SignatureOptions{
		Detached:     true,
		IncludeChain: true,
		AddTimestamp: false,
		TSAUrl:       "http://timestamping.iti.gov.br/tsa", // Default ICP-Brasil TSA
	}
}

// SignedDocument represents a signed document
type SignedDocument struct {
	OriginalContent []byte
	Signature       []byte
	Certificate     *Certificate
	SignatureTime   time.Time
	TimestampTime   time.Time
	HasTimestamp    bool
}

// SignData signs data using CMS/PKCS#7
func (c *Certificate) SignData(data []byte, options *SignatureOptions) (*SignedDocument, error) {
	// Create a SignedData object
	sd, err := pkcs7.NewSignedData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed data: %w", err)
	}

	// Add the signer
	if err := sd.AddSigner(c.Cert, c.PrivateKey, pkcs7.SignerInfoConfig{}); err != nil {
		return nil, fmt.Errorf("failed to add signer: %w", err)
	}

	// Include the certificate chain if requested
	if options.IncludeChain {
		for i, cert := range c.Chain {
			if i > 0 { // Skip the leaf certificate (already added as signer)
				sd.AddCertificate(cert)
			}
		}
	}

	// Finalize the signature
	var signature []byte
	signature, err = sd.Finish()

	if err != nil {
		return nil, fmt.Errorf("failed to finalize signature: %w", err)
	}

	// Create the signed document
	signedDoc := &SignedDocument{
		OriginalContent: data,
		Signature:       signature,
		Certificate:     c,
		SignatureTime:   time.Now(),
		HasTimestamp:    false,
	}

	// Add timestamp if requested
	if options.AddTimestamp && options.TSAUrl != "" {
		if err := signedDoc.AddTimestamp(options.TSAUrl); err != nil {
			return signedDoc, fmt.Errorf("signature created but timestamp failed: %w", err)
		}
	}

	return signedDoc, nil
}

// SignFile signs a file using CMS/PKCS#7
func (c *Certificate) SignFile(filename string, options *SignatureOptions) (*SignedDocument, error) {
	// Read the file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return c.SignData(data, options)
}

// AddTimestamp adds a timestamp to the signature
func (sd *SignedDocument) AddTimestamp(tsaURL string) error {
	// Calculate the hash of the signature
	hash := sha256.Sum256(sd.Signature)

	// Create a timestamp request
	req, err := timestamp.CreateRequest(bytes.NewReader(hash[:]), &timestamp.RequestOptions{
		Certificates: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create timestamp request: %w", err)
	}

	// Send the request to the TSA
	resp, err := http.Post(tsaURL, "application/timestamp-query", bytes.NewReader(req))
	if err != nil {
		return fmt.Errorf("failed to send timestamp request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	tsResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read timestamp response: %w", err)
	}

	// Parse the timestamp response
	ts, err := timestamp.ParseResponse(tsResp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp response: %w", err)
	}

	// Update the signed document
	sd.TimestampTime = ts.Time
	sd.HasTimestamp = true

	return nil
}

// VerifySignature verifies a CMS/PKCS#7 signature
func VerifySignature(data, signature []byte) (*VerificationResult, error) {
	// Parse the signature
	p7, err := pkcs7.Parse(signature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature: %w", err)
	}

	// Create a verification result
	result := &VerificationResult{
		IsValid:       false,
		SignerCert:    nil,
		SignatureTime: time.Time{},
		Errors:        []string{},
	}

	// Extract the signer certificate
	if len(p7.Certificates) == 0 {
		result.Errors = append(result.Errors, "no certificates found in signature")
		return result, nil
	}

	// Get the signer certificate
	signerCert := p7.Certificates[0]
	result.SignerCert = signerCert

	// Create a certificate pool for verification
	pool := x509.NewCertPool()
	for _, cert := range p7.Certificates {
		pool.AddCert(cert)
	}

	// Verify the signature
	if err := p7.VerifyWithChain(pool); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("signature verification failed: %v", err))
		return result, nil
	}

	// For attached signatures, verify that the content matches
	if p7.Content != nil && !bytes.Equal(p7.Content, data) {
		result.Errors = append(result.Errors, "content in signature does not match provided data")
		return result, nil
	}

	// If we got here, the signature is valid
	result.IsValid = true

	return result, nil
}

// VerificationResult contains the result of a signature verification
type VerificationResult struct {
	IsValid       bool
	SignerCert    *x509.Certificate
	SignatureTime time.Time
	Errors        []string
}

// SaveSignature saves the signature to a file
func (sd *SignedDocument) SaveSignature(filename string) error {
	return ioutil.WriteFile(filename, sd.Signature, 0644)
}

// GetSignatureBase64 returns the signature as a base64-encoded string
func (sd *SignedDocument) GetSignatureBase64() string {
	return base64.StdEncoding.EncodeToString(sd.Signature)
}

// GetDocumentHash returns the SHA-256 hash of the original document
func (sd *SignedDocument) GetDocumentHash() string {
	hash := sha256.Sum256(sd.OriginalContent)
	return hex.EncodeToString(hash[:])
}

// GetSignatureHash returns the SHA-256 hash of the signature
func (sd *SignedDocument) GetSignatureHash() string {
	hash := sha256.Sum256(sd.Signature)
	return hex.EncodeToString(hash[:])
}

// GetSignerCertificateHash returns the SHA-256 hash of the signer's certificate
func (sd *SignedDocument) GetSignerCertificateHash() string {
	hash := sha256.Sum256(sd.Certificate.Cert.Raw)
	return hex.EncodeToString(hash[:])
}

// GetSignatureInfo returns information about the signature
func (sd *SignedDocument) GetSignatureInfo() map[string]string {
	info := make(map[string]string)
	info["DocumentHash"] = sd.GetDocumentHash()
	info["SignatureHash"] = sd.GetSignatureHash()
	info["SignerCertificateHash"] = sd.GetSignerCertificateHash()
	info["SignatureTime"] = sd.SignatureTime.Format(time.RFC3339)
	info["HasTimestamp"] = fmt.Sprintf("%t", sd.HasTimestamp)
	if sd.HasTimestamp {
		info["TimestampTime"] = sd.TimestampTime.Format(time.RFC3339)
	}
	return info
}

// VerifyDetachedSignature verifies a detached CMS/PKCS#7 signature
func VerifyDetachedSignature(data, signature []byte) (*VerificationResult, error) {
	// Parse the signature
	p7, err := pkcs7.Parse(signature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature: %w", err)
	}

	// For detached signatures, the content should be nil
	if p7.Content != nil {
		return nil, errors.New("not a detached signature")
	}

	return VerifySignature(data, signature)
}

// VerifyAttachedSignature verifies an attached CMS/PKCS#7 signature
func VerifyAttachedSignature(signature []byte) (*VerificationResult, error) {
	// Parse the signature
	p7, err := pkcs7.Parse(signature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature: %w", err)
	}

	// For attached signatures, the content should not be nil
	if p7.Content == nil {
		return nil, errors.New("not an attached signature")
	}

	return VerifySignature(p7.Content, signature)
}
