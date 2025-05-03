package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/raykavin/docchain/internal/blockchain"
	"github.com/raykavin/docchain/internal/crypto"
	"github.com/raykavin/docchain/internal/models"
)

// WalletHandlers contains handlers for wallet-related endpoints
type WalletHandlers struct {
	server *Server
}

// CreateWallet creates a new wallet
func (h *WalletHandlers) CreateWallet(w http.ResponseWriter, r *http.Request) {
	wallet, err := crypto.NewWallet()
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create wallet: "+err.Error())
		return
	}
	
	id := wallet.GetPublicKeyAsString()
	h.server.Wallets[id] = wallet
	
	privateKeyPEM, err := wallet.ExportPrivateKey()
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to export private key: "+err.Error())
		return
	}
	
	SendSuccessResponse(w, http.StatusCreated, "Wallet created successfully", map[string]string{
		"id":         id,
		"publicKey":  id,
		"privateKey": privateKeyPEM,
	})
}

// GetWallet gets a wallet by ID
func (h *WalletHandlers) GetWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	wallet, exists := h.server.Wallets[id]
	if !exists {
		SendErrorResponse(w, http.StatusNotFound, "Wallet not found")
		return
	}
	
	SendSuccessResponse(w, http.StatusOK, "Wallet found", map[string]string{
		"id":        id,
		"publicKey": wallet.GetPublicKeyAsString(),
	})
}

// DocumentHandlers contains handlers for document-related endpoints
type DocumentHandlers struct {
	server *Server
}

// SignDocument signs a document and adds it to the blockchain
func (h *DocumentHandlers) SignDocument(w http.ResponseWriter, r *http.Request) {
	var req models.DocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	// Import the private key
	privateKey, err := crypto.ImportPrivateKey(req.Key)
	if err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "Invalid private key: "+err.Error())
		return
	}
	
	// Create a temporary wallet with the imported private key
	tempWallet := &crypto.Wallet{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}
	
	// Hash the content
	contentHash := crypto.CalculateHashFromString(req.Content)
	
	// Sign the document hash
	signature, err := tempWallet.SignString(contentHash)
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to sign document: "+err.Error())
		return
	}
	
	// Create a new document
	doc := blockchain.NewDocument(
		req.Content,
		signature,
		tempWallet.GetPublicKeyAsString(),
	)
	
	// Add the document to the blockchain
	if !h.server.Blockchain.AddDocument(doc) {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to add document to blockchain")
		return
	}
	
	SendSuccessResponse(w, http.StatusCreated, "Document signed and added to pending documents", map[string]string{
		"documentId": doc.ID,
		"signedBy":   doc.SignedBy,
		"timestamp":  strconv.FormatInt(doc.Timestamp, 10),
	})
}

// GetDocument gets a document by ID
func (h *DocumentHandlers) GetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	doc := h.server.Blockchain.GetDocumentByID(id)
	if doc == nil {
		SendErrorResponse(w, http.StatusNotFound, "Document not found")
		return
	}
	
	SendSuccessResponse(w, http.StatusOK, "Document found", doc)
}

// VerifyDocument verifies a document's signature
func (h *DocumentHandlers) VerifyDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	doc := h.server.Blockchain.GetDocumentByID(id)
	if doc == nil {
		SendErrorResponse(w, http.StatusNotFound, "Document not found")
		return
	}
	
	isValid := blockchain.VerifyDocument(doc)
	
	SendSuccessResponse(w, http.StatusOK, "Document verification completed", map[string]interface{}{
		"documentId": doc.ID,
		"isValid":    isValid,
		"signedBy":   doc.SignedBy,
		"timestamp":  doc.Timestamp,
	})
}

// BlockchainHandlers contains handlers for blockchain-related endpoints
type BlockchainHandlers struct {
	server *Server
}

// GetBlockchainInfo gets general information about the blockchain
func (h *BlockchainHandlers) GetBlockchainInfo(w http.ResponseWriter, r *http.Request) {
	bc := h.server.Blockchain
	
	SendSuccessResponse(w, http.StatusOK, "Blockchain info retrieved", map[string]interface{}{
		"blocks":      len(bc.Blocks),
		"difficulty":  bc.Difficulty,
		"pendingDocs": len(bc.PendingDocs),
		"isValid":     bc.IsChainValid(),
	})
}

// MineBlock mines a new block
func (h *BlockchainHandlers) MineBlock(w http.ResponseWriter, r *http.Request) {
	block, err := h.server.Blockchain.MineBlock()
	if err != nil {
		SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	
	SendSuccessResponse(w, http.StatusCreated, "Block mined successfully", map[string]interface{}{
		"index":      block.Index,
		"hash":       block.Hash,
		"documents":  len(block.Documents),
		"timestamp":  block.Timestamp,
		"merkleRoot": block.MerkleRoot,
		"nonce":      block.Nonce,
	})
}

// GetBlocks gets all blocks in the blockchain
func (h *BlockchainHandlers) GetBlocks(w http.ResponseWriter, r *http.Request) {
	SendSuccessResponse(w, http.StatusOK, "Blocks retrieved", h.server.Blockchain.Blocks)
}

// GetBlockByHash gets a block by its hash
func (h *BlockchainHandlers) GetBlockByHash(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	
	block := h.server.Blockchain.GetBlockByHash(hash)
	if block == nil {
		SendErrorResponse(w, http.StatusNotFound, "Block not found")
		return
	}
	
	SendSuccessResponse(w, http.StatusOK, "Block found", block)
}

// GetPendingDocuments gets all pending documents
func (h *BlockchainHandlers) GetPendingDocuments(w http.ResponseWriter, r *http.Request) {
	SendSuccessResponse(w, http.StatusOK, "Pending documents retrieved", h.server.Blockchain.PendingDocs)
}

// ValidateChain validates the entire blockchain
func (h *BlockchainHandlers) ValidateChain(w http.ResponseWriter, r *http.Request) {
	isValid := h.server.Blockchain.IsChainValid()
	
	SendSuccessResponse(w, http.StatusOK, "Blockchain validation completed", map[string]bool{
		"isValid": isValid,
	})
}