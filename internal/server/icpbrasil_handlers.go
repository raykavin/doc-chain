package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/raykavin/doc-chain/internal/icpbrasil"
)

// ICPBrasilHandlers contains handlers for ICP-Brasil-related endpoints
type ICPBrasilHandlers struct {
	server *Server
}

// ICPBrasilSignRequest represents a request to sign a document with an ICP-Brasil certificate
type ICPBrasilSignRequest struct {
	// Document content (base64 encoded)
	Document string `json:"document"`

	// Certificate content (base64 encoded)
	Certificate string `json:"certificate"`

	// Certificate password
	Password string `json:"password"`

	// Signature options
	Detached     bool   `json:"detached"`
	IncludeChain bool   `json:"includeChain"`
	AddTimestamp bool   `json:"addTimestamp"`
	TSAUrl       string `json:"tsaUrl"`

	// Blockchain options
	AddToBlockchain bool `json:"addToBlockchain"`
	MineBlock       bool `json:"mineBlock"`
}

// ICPBrasilVerifyRequest represents a request to verify a document signature
type ICPBrasilVerifyRequest struct {
	// Document content (base64 encoded)
	Document string `json:"document"`

	// Signature content (base64 encoded)
	Signature string `json:"signature"`

	// Blockchain options
	CheckBlockchain bool `json:"checkBlockchain"`
}

// ICPBrasilCertInfoRequest represents a request to get information about a certificate
type ICPBrasilCertInfoRequest struct {
	// Certificate content (base64 encoded)
	Certificate string `json:"certificate"`

	// Certificate password
	Password string `json:"password"`
}

// ICPBrasilCheckRevocationRequest represents a request to check if a certificate is revoked
type ICPBrasilCheckRevocationRequest struct {
	// Certificate content (base64 encoded)
	Certificate string `json:"certificate"`

	// Certificate password
	Password string `json:"password"`
}

// SignDocument signs a document with an ICP-Brasil certificate
func (h *ICPBrasilHandlers) SignDocument(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var req ICPBrasilSignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate the request
	if req.Document == "" {
		SendErrorResponse(w, http.StatusBadRequest, "Document is required")
		return
	}
	if req.Certificate == "" {
		SendErrorResponse(w, http.StatusBadRequest, "Certificate is required")
		return
	}

	// Create a temporary directory for the files
	tempDir, err := ioutil.TempDir("", "icpbrasil-sign-")
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create temporary directory: "+err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	// Decode the document
	documentData, err := base64.StdEncoding.DecodeString(req.Document)
	if err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid document encoding: "+err.Error())
		return
	}

	// Write the document to a temporary file
	documentPath := filepath.Join(tempDir, "document")
	if err := ioutil.WriteFile(documentPath, documentData, 0644); err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to write document: "+err.Error())
		return
	}

	// Decode the certificate
	certificateData, err := base64.StdEncoding.DecodeString(req.Certificate)
	if err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid certificate encoding: "+err.Error())
		return
	}

	// Write the certificate to a temporary file
	certificatePath := filepath.Join(tempDir, "certificate.p12")
	if err := ioutil.WriteFile(certificatePath, certificateData, 0644); err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to write certificate: "+err.Error())
		return
	}

	// Set default TSA URL if not provided
	if req.TSAUrl == "" {
		req.TSAUrl = "http://timestamping.iti.gov.br/tsa"
	}

	// Create signature options
	options := &icpbrasil.SignatureOptions{
		Detached:     req.Detached,
		IncludeChain: req.IncludeChain,
		AddTimestamp: req.AddTimestamp,
		TSAUrl:       req.TSAUrl,
	}

	// Create a document manager
	dm := icpbrasil.NewDocumentManager(h.server.Blockchain)

	var signedDoc *icpbrasil.SignedDocument
	var bcDoc *icpbrasil.BlockchainDocument
	var signErr error

	if req.AddToBlockchain {
		// Sign and add to blockchain
		bcDoc, signErr = dm.SignAndAddToBlockchain(
			documentPath,
			certificatePath,
			req.Password,
			options,
		)
		if signErr != nil {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to sign and add to blockchain: "+signErr.Error())
			return
		}
		signedDoc = bcDoc.SignedDocument
	} else {
		// Just sign the document
		signedDoc, signErr = dm.SignDocumentWithCertificate(
			documentPath,
			certificatePath,
			req.Password,
			options,
		)
		if signErr != nil {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to sign document: "+signErr.Error())
			return
		}
	}

	// Get signature information
	info := signedDoc.GetSignatureInfo()

	// Prepare the response
	response := map[string]interface{}{
		"documentHash":          info["DocumentHash"],
		"signatureHash":         info["SignatureHash"],
		"signerCertificateHash": info["SignerCertificateHash"],
		"signatureTime":         info["SignatureTime"],
		"hasTimestamp":          info["HasTimestamp"] == "true",
		"signature":             base64.StdEncoding.EncodeToString(signedDoc.Signature),
		"addedToBlockchain":     req.AddToBlockchain,
	}

	if info["HasTimestamp"] == "true" {
		response["timestampTime"] = info["TimestampTime"]
	}

	// Mine a block if requested
	if req.AddToBlockchain && req.MineBlock {
		block, err := h.server.Blockchain.MineBlock()
		if err != nil {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to mine block: "+err.Error())
			return
		}
		response["blockMined"] = true
		response["blockIndex"] = block.Index
		response["blockHash"] = block.Hash
	} else {
		response["blockMined"] = false
	}

	SendSuccessResponse(w, http.StatusOK, "Document signed successfully", response)
}

// VerifySignature verifies a document signature
func (h *ICPBrasilHandlers) VerifySignature(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var req ICPBrasilVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate the request
	if req.Document == "" {
		SendErrorResponse(w, http.StatusBadRequest, "Document is required")
		return
	}
	if req.Signature == "" {
		SendErrorResponse(w, http.StatusBadRequest, "Signature is required")
		return
	}

	// Create a temporary directory for the files
	tempDir, err := ioutil.TempDir("", "icpbrasil-verify-")
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create temporary directory: "+err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	// Decode the document
	documentData, err := base64.StdEncoding.DecodeString(req.Document)
	if err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid document encoding: "+err.Error())
		return
	}

	// Write the document to a temporary file
	documentPath := filepath.Join(tempDir, "document")
	if err := ioutil.WriteFile(documentPath, documentData, 0644); err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to write document: "+err.Error())
		return
	}

	// Decode the signature
	signatureData, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid signature encoding: "+err.Error())
		return
	}

	// Write the signature to a temporary file
	signaturePath := filepath.Join(tempDir, "signature.p7s")
	if err := ioutil.WriteFile(signaturePath, signatureData, 0644); err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to write signature: "+err.Error())
		return
	}

	// Create a document manager
	dm := icpbrasil.NewDocumentManager(h.server.Blockchain)

	var result *icpbrasil.VerificationResult
	var bcValid bool
	var verifyErr error

	if req.CheckBlockchain {
		// Verify signature and check blockchain
		bcValid, result, verifyErr = dm.VerifySignatureAndBlockchain(documentPath, signaturePath)
		if verifyErr != nil && !strings.HasPrefix(verifyErr.Error(), "blockchain verification") {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to verify signature: "+verifyErr.Error())
			return
		}
	} else {
		// Just verify the signature
		result, verifyErr = dm.VerifySignature(documentPath, signaturePath)
		if verifyErr != nil {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to verify signature: "+verifyErr.Error())
			return
		}
	}

	// Prepare the response
	response := map[string]interface{}{
		"isValid": result.IsValid,
		"errors":  result.Errors,
	}

	// Add signer information if available
	if result.SignerCert != nil {
		response["signer"] = result.SignerCert.Subject.CommonName
		response["issuer"] = result.SignerCert.Issuer.CommonName
		response["serialNumber"] = result.SignerCert.SerialNumber.String()
		response["validFrom"] = result.SignerCert.NotBefore.Format(time.RFC3339)
		response["validTo"] = result.SignerCert.NotAfter.Format(time.RFC3339)
	}

	// Add blockchain verification result if requested
	if req.CheckBlockchain {
		response["inBlockchain"] = bcValid
		if verifyErr != nil {
			response["blockchainError"] = verifyErr.Error()
		}
	}

	SendSuccessResponse(w, http.StatusOK, "Signature verification completed", response)
}

// GetCertificateInfo gets information about a certificate
func (h *ICPBrasilHandlers) GetCertificateInfo(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var req ICPBrasilCertInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate the request
	if req.Certificate == "" {
		SendErrorResponse(w, http.StatusBadRequest, "Certificate is required")
		return
	}

	// Create a temporary directory for the files
	tempDir, err := ioutil.TempDir("", "icpbrasil-certinfo-")
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create temporary directory: "+err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	// Decode the certificate
	certificateData, err := base64.StdEncoding.DecodeString(req.Certificate)
	if err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid certificate encoding: "+err.Error())
		return
	}

	// Write the certificate to a temporary file
	certificatePath := filepath.Join(tempDir, "certificate.p12")
	if err := ioutil.WriteFile(certificatePath, certificateData, 0644); err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to write certificate: "+err.Error())
		return
	}

	// Load the certificate
	cert, err := icpbrasil.LoadPFXFromFile(certificatePath, req.Password)
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to load certificate: "+err.Error())
		return
	}

	// Get certificate information
	info := cert.GetCertificateInfo()

	SendSuccessResponse(w, http.StatusOK, "Certificate information retrieved", info)
}

// CheckRevocation checks if a certificate is revoked
func (h *ICPBrasilHandlers) CheckRevocation(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var req ICPBrasilCheckRevocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate the request
	if req.Certificate == "" {
		SendErrorResponse(w, http.StatusBadRequest, "Certificate is required")
		return
	}

	// Create a temporary directory for the files
	tempDir, err := ioutil.TempDir("", "icpbrasil-revocation-")
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create temporary directory: "+err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	// Decode the certificate
	certificateData, err := base64.StdEncoding.DecodeString(req.Certificate)
	if err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid certificate encoding: "+err.Error())
		return
	}

	// Write the certificate to a temporary file
	certificatePath := filepath.Join(tempDir, "certificate.p12")
	if err := ioutil.WriteFile(certificatePath, certificateData, 0644); err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to write certificate: "+err.Error())
		return
	}

	// Create a document manager
	dm := icpbrasil.NewDocumentManager(h.server.Blockchain)

	// Check revocation
	isRevoked, err := dm.CheckRevocation(certificatePath, req.Password)
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to check revocation: "+err.Error())
		return
	}

	SendSuccessResponse(w, http.StatusOK, "Revocation check completed", map[string]bool{
		"isRevoked": isRevoked,
	})
}

// GetDocumentInfo gets information about a signed document
func (h *ICPBrasilHandlers) GetDocumentInfo(w http.ResponseWriter, r *http.Request) {
	// Get document ID from URL
	vars := mux.Vars(r)
	documentID := vars["id"]

	// Get the document from the blockchain
	doc := h.server.Blockchain.GetDocumentByID(documentID)
	if doc == nil {
		SendErrorResponse(w, http.StatusNotFound, "Document not found")
		return
	}

	// Extract document information
	info := make(map[string]string)
	info["DocumentID"] = doc.ID
	info["Content"] = doc.Content
	info["SignedBy"] = doc.SignedBy
	info["Timestamp"] = fmt.Sprintf("%d", doc.Timestamp)

	// Find the block containing the document
	var blockInfo string
	for _, block := range h.server.Blockchain.Blocks {
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

	SendSuccessResponse(w, http.StatusOK, "Document information retrieved", info)
}
