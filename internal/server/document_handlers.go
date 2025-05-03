package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/raykavin/doc-chain/internal/blockchain"
	"github.com/raykavin/doc-chain/internal/crypto"
)

// DocumentHandlersDB handles document-related HTTP requests with database storage
type DocumentHandlersDB struct {
	server *Server
}

// SignDocument handles document signing requests
func (h *DocumentHandlersDB) SignDocument(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		DocumentID       string `json:"documentId"`
		WalletID         string `json:"walletId"`
		Password         string `json:"password"`
		AddToBlockchain  bool   `json:"addToBlockchain"`
		IncludeTimestamp bool   `json:"includeTimestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Erro ao analisar solicitação: "+err.Error())
		return
	}

	// Get wallet from database
	wallet, err := h.server.WalletRepo.GetWallet(req.WalletID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Carteira não encontrada: "+err.Error())
		return
	}

	// Get certificate from database
	cert, err := h.server.WalletRepo.GetCertificate(req.WalletID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Certificado não encontrado: "+err.Error())
		return
	}

	// Check if document exists
	_, err = h.server.DocumentRepo.GetDocument(req.DocumentID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Documento não encontrado: "+err.Error())
		return
	}

	// Get document content
	content, err := h.server.DocumentRepo.GetDocumentContent(req.DocumentID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao obter conteúdo do documento: "+err.Error())
		return
	}

	// Verify certificate password
	// In a real implementation, this would verify the password against the stored hash
	// For simplicity, we're just checking if a password was provided
	if req.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Senha do certificado é obrigatória")
		return
	}

	// Import the private key
	privateKey, err := crypto.ImportPrivateKey(wallet.PrivateKey)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao importar chave privada: "+err.Error())
		return
	}

	// Create a wallet instance from the private key
	walletInstance := &crypto.Wallet{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}

	// Sign document content with wallet's private key
	signatureObj, err := walletInstance.SignData(content)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao assinar documento: "+err.Error())
		return
	}

	// Convert signature to bytes
	rBytes := signatureObj.R.Bytes()
	sBytes := signatureObj.S.Bytes()
	signature := append(rBytes, sBytes...)

	// Calculate signature hash
	hash := sha256.Sum256(signature)
	signatureHash := hex.EncodeToString(hash[:])

	// Store signature in database
	signatureID, err := h.server.DocumentRepo.SignDocument(
		req.DocumentID,
		req.WalletID,
		signatureHash,
		cert.ID,
		signature,
		req.IncludeTimestamp,
	)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao salvar assinatura: "+err.Error())
		return
	}

	// Add to blockchain if requested
	if req.AddToBlockchain {
		// Create a blockchain document
		contentStr := hex.EncodeToString(content)
		bcDoc := blockchain.NewDocument(contentStr, signatureObj, wallet.PublicKey)

		// Add document to blockchain
		success := h.server.Blockchain.AddDocument(bcDoc)
		if !success {
			RespondWithError(w, http.StatusInternalServerError, "Erro ao adicionar documento à blockchain")
			return
		}
	}

	// Get updated document
	updatedDoc, err := h.server.DocumentRepo.GetDocument(req.DocumentID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao obter documento atualizado: "+err.Error())
		return
	}

	// Respond with success
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"message":     "Documento assinado com sucesso",
		"signatureId": signatureID,
		"document":    updatedDoc,
	})
}

// GetDocument handles document retrieval requests
func (h *DocumentHandlersDB) GetDocument(w http.ResponseWriter, r *http.Request) {
	// Get document ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get document from database
	doc, err := h.server.DocumentRepo.GetDocument(id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Documento não encontrado: "+err.Error())
		return
	}

	// Respond with document
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"document": doc,
	})
}

// VerifyDocument handles document verification requests
func (h *DocumentHandlersDB) VerifyDocument(w http.ResponseWriter, r *http.Request) {
	// Get document ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get document from database
	doc, err := h.server.DocumentRepo.GetDocument(id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Documento não encontrado: "+err.Error())
		return
	}

	// Check if document is signed
	if doc.Status != "signed" {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"valid":        false,
			"message":      "Documento não está assinado",
			"inBlockchain": false,
		})
		return
	}

	// Check if document has content
	_, err = h.server.DocumentRepo.GetDocumentContent(id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao obter conteúdo do documento: "+err.Error())
		return
	}

	// Get wallet from database
	wallet, err := h.server.WalletRepo.GetWallet(doc.SignedBy)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao obter carteira: "+err.Error())
		return
	}

	// Verify signature
	// In a real implementation, this would verify the signature against the document content
	// For simplicity, we're just checking if the document is marked as signed
	valid := doc.Status == "signed"

	// Check if document is in blockchain
	inBlockchain := doc.BlockIndex != nil

	// Respond with verification result
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"valid":        valid,
		"inBlockchain": inBlockchain,
		"document":     doc,
		"wallet":       wallet.Name,
	})
}

// ListDocuments handles document listing requests
func (h *DocumentHandlersDB) ListDocuments(w http.ResponseWriter, r *http.Request) {
	// Get documents from database
	docs, err := h.server.DocumentRepo.GetDocuments()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao listar documentos: "+err.Error())
		return
	}

	// Respond with documents
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"documents": docs,
	})
}

// UploadDocument handles document upload requests
func (h *DocumentHandlersDB) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Erro ao analisar formulário: "+err.Error())
		return
	}

	// Get title from form
	title := r.FormValue("title")
	if title == "" {
		RespondWithError(w, http.StatusBadRequest, "Título do documento é obrigatório")
		return
	}

	// Get file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Erro ao obter arquivo: "+err.Error())
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao ler arquivo: "+err.Error())
		return
	}

	// Determine content type from file extension
	contentType := filepath.Ext(handler.Filename)
	if contentType != "" && contentType[0] == '.' {
		contentType = contentType[1:] // Remove leading dot
	}

	// Create document in database
	doc, err := h.server.DocumentRepo.CreateDocument(title, contentType, content)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao criar documento: "+err.Error())
		return
	}

	// Respond with success
	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success":  true,
		"message":  "Documento enviado com sucesso",
		"document": doc,
	})
}

// DownloadDocument handles document download requests
func (h *DocumentHandlersDB) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	// Get document ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get document from database
	doc, err := h.server.DocumentRepo.GetDocument(id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Documento não encontrado: "+err.Error())
		return
	}

	// Get document content
	content, err := h.server.DocumentRepo.GetDocumentContent(id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao obter conteúdo do documento: "+err.Error())
		return
	}

	// Set content type header
	contentType := "application/octet-stream"
	switch doc.ContentType {
	case "pdf":
		contentType = "application/pdf"
	case "txt":
		contentType = "text/plain"
	case "html":
		contentType = "text/html"
	case "json":
		contentType = "application/json"
	}

	// Set headers for file download
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+doc.Title+"."+doc.ContentType)
	w.Header().Set("Content-Length", fmt.Sprint(len(content)))

	// Write content to response
	_, err = w.Write(content)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao enviar arquivo: "+err.Error())
		return
	}
}
