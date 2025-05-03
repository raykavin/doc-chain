package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// WalletHandlersDB handles wallet-related HTTP requests with database storage
type WalletHandlersDB struct {
	server *Server
}

// CreateWallet handles wallet creation requests
func (h *WalletHandlersDB) CreateWallet(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Erro ao analisar solicitação: "+err.Error())
		return
	}

	// Validate request
	if req.Name == "" {
		RespondWithError(w, http.StatusBadRequest, "Nome da carteira é obrigatório")
		return
	}

	// Store wallet in database
	dbWallet, err := h.server.WalletRepo.CreateWallet(req.Name)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao salvar carteira: "+err.Error())
		return
	}

	// Respond with success
	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Carteira criada com sucesso",
		"wallet":  dbWallet,
	})
}

// GetWallet handles wallet retrieval requests
func (h *WalletHandlersDB) GetWallet(w http.ResponseWriter, r *http.Request) {
	// Get wallet ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get wallet from database
	wallet, err := h.server.WalletRepo.GetWallet(id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Carteira não encontrada: "+err.Error())
		return
	}

	// Respond with wallet
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"wallet": wallet,
	})
}

// ListWallets handles wallet listing requests
func (h *WalletHandlersDB) ListWallets(w http.ResponseWriter, r *http.Request) {
	// Get wallets from database
	wallets, err := h.server.WalletRepo.GetWallets()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao listar carteiras: "+err.Error())
		return
	}

	// Respond with wallets
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"wallets": wallets,
	})
}

// CreateCertificate handles certificate creation requests
func (h *WalletHandlersDB) CreateCertificate(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		WalletID string `json:"walletId"`
		CertType string `json:"certType"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Erro ao analisar solicitação: "+err.Error())
		return
	}

	// Validate request
	if req.WalletID == "" {
		RespondWithError(w, http.StatusBadRequest, "ID da carteira é obrigatório")
		return
	}

	if req.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Senha do certificado é obrigatória")
		return
	}

	// Get wallet from database
	wallet, err := h.server.WalletRepo.GetWallet(req.WalletID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Carteira não encontrada: "+err.Error())
		return
	}

	// Generate certificate data
	// In a real implementation, this would generate a proper X.509 certificate
	// For simplicity, we're just using the wallet's public key as the certificate data
	certData := []byte(wallet.PublicKey)

	// Hash the password
	// In a real implementation, this would use a proper password hashing algorithm
	passwordHash := req.Password // Simplified for demo

	// Store certificate in database
	cert, err := h.server.WalletRepo.CreateCertificate(req.WalletID, req.CertType, certData, passwordHash)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Erro ao criar certificado: "+err.Error())
		return
	}

	// Respond with success
	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success":     true,
		"message":     "Certificado criado com sucesso",
		"certificate": cert,
	})
}
