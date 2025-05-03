package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/raykavin/doc-chain/internal/blockchain"
	"github.com/raykavin/doc-chain/internal/crypto"
	"github.com/raykavin/doc-chain/internal/dto"
)

// WalletHandlers contains handlers for wallet-related endpoints
type WalletHandlers struct {
	server *Server
}

// CreateWallet creates a new wallet
func (h *WalletHandlers) CreateWallet(w http.ResponseWriter, r *http.Request) {
	wallet, err := crypto.NewWallet()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Falha ao criar carteira: "+err.Error())
		return
	}

	id := wallet.GetPublicKeyAsString()
	h.server.Wallets[id] = wallet

	privateKeyPEM, err := wallet.ExportPrivateKey()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Falha ao exportar chave privada: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Carteira criada com sucesso",
		"data": map[string]string{
			"id":         id,
			"publicKey":  id,
			"privateKey": privateKeyPEM,
		},
	})
}

// GetWallet gets a wallet by ID
func (h *WalletHandlers) GetWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	wallet, exists := h.server.Wallets[id]
	if !exists {
		RespondWithError(w, http.StatusNotFound, "Carteira não encontrada")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Carteira encontrada",
		"data": map[string]string{
			"id":        id,
			"publicKey": wallet.GetPublicKeyAsString(),
		},
	})
}

// DocumentHandlers contains handlers for document-related endpoints
type DocumentHandlers struct {
	server *Server
}

// SignDocument signs a document and adds it to the blockchain
func (h *DocumentHandlers) SignDocument(w http.ResponseWriter, r *http.Request) {
	var req dto.DocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Requisição inválida: "+err.Error())
		return
	}

	// Import the private key
	privateKey, err := crypto.ImportPrivateKey(req.Key)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Chave privada inválida: "+err.Error())
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
		RespondWithError(w, http.StatusInternalServerError, "Falha ao assinar documento: "+err.Error())
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
		RespondWithError(w, http.StatusInternalServerError, "Falha ao adicionar documento à blockchain")
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Documento assinado e adicionado aos documentos pendentes",
		"data": map[string]string{
			"documentId": doc.ID,
			"signedBy":   doc.SignedBy,
			"timestamp":  strconv.FormatInt(doc.Timestamp, 10),
		},
	})
}

// GetDocument gets a document by ID
func (h *DocumentHandlers) GetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	doc := h.server.Blockchain.GetDocumentByID(id)
	if doc == nil {
		RespondWithError(w, http.StatusNotFound, "Documento não encontrado")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Documento encontrado",
		"data":    doc,
	})
}

// VerifyDocument verifies a document's signature
func (h *DocumentHandlers) VerifyDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	doc := h.server.Blockchain.GetDocumentByID(id)
	if doc == nil {
		RespondWithError(w, http.StatusNotFound, "Documento não encontrado")
		return
	}

	isValid := blockchain.VerifyDocument(doc)

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Verificação de documento concluída",
		"data": map[string]interface{}{
			"documentId": doc.ID,
			"isValid":    isValid,
			"signedBy":   doc.SignedBy,
			"timestamp":  doc.Timestamp,
		},
	})
}

// BlockchainHandlers contains handlers for blockchain-related endpoints
type BlockchainHandlers struct {
	server *Server
}

// GetBlockchainInfo gets general information about the blockchain
func (h *BlockchainHandlers) GetBlockchainInfo(w http.ResponseWriter, r *http.Request) {
	bc := h.server.Blockchain

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Informações da blockchain recuperadas",
		"data": map[string]interface{}{
			"blocks":      len(bc.Blocks),
			"difficulty":  bc.Difficulty,
			"pendingDocs": len(bc.PendingDocs),
			"isValid":     bc.IsChainValid(),
		},
	})
}

// MineBlock mines a new block
func (h *BlockchainHandlers) MineBlock(w http.ResponseWriter, r *http.Request) {
	block, err := h.server.Blockchain.MineBlock()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Bloco minerado com sucesso",
		"data": map[string]interface{}{
			"index":      block.Index,
			"hash":       block.Hash,
			"documents":  len(block.Documents),
			"timestamp":  block.Timestamp,
			"merkleRoot": block.MerkleRoot,
			"nonce":      block.Nonce,
		},
	})
}

// GetBlocks gets all blocks in the blockchain
func (h *BlockchainHandlers) GetBlocks(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Blocos recuperados",
		"data":    h.server.Blockchain.Blocks,
	})
}

// GetBlockByHash gets a block by its hash
func (h *BlockchainHandlers) GetBlockByHash(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	block := h.server.Blockchain.GetBlockByHash(hash)
	if block == nil {
		RespondWithError(w, http.StatusNotFound, "Bloco não encontrado")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Bloco encontrado",
		"data":    block,
	})
}

// GetPendingDocuments gets all pending documents
func (h *BlockchainHandlers) GetPendingDocuments(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Documentos pendentes recuperados",
		"data":    h.server.Blockchain.PendingDocs,
	})
}

// ValidateChain validates the entire blockchain
func (h *BlockchainHandlers) ValidateChain(w http.ResponseWriter, r *http.Request) {
	isValid := h.server.Blockchain.IsChainValid()

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Validação da blockchain concluída",
		"data": map[string]bool{
			"isValid": isValid,
		},
	})
}
