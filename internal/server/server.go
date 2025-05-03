package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/raykavin/docchain/internal/blockchain"
	"github.com/raykavin/docchain/internal/crypto"
)

// Server represents the blockchain API server
type Server struct {
	Blockchain *blockchain.Blockchain
	Wallets    map[string]*crypto.Wallet
	Port       string
	Router     *mux.Router
}

// NewServer creates a new blockchain API server
func NewServer(port string, difficulty int) *Server {
	return &Server{
		Blockchain: blockchain.NewBlockchain(difficulty),
		Wallets:    make(map[string]*crypto.Wallet),
		Port:       port,
		Router:     mux.NewRouter(),
	}
}

// Configure sets up the server routes and middleware
func (s *Server) Configure() {
	// Create handlers
	walletHandlers := &WalletHandlers{server: s}
	documentHandlers := &DocumentHandlers{server: s}
	blockchainHandlers := &BlockchainHandlers{server: s}
	
	// Create a subrouter for API endpoints
	apiRouter := s.Router.PathPrefix("/api").Subrouter()
	
	// Wallet routes
	walletRouter := apiRouter.PathPrefix("/wallet").Subrouter()
	walletRouter.HandleFunc("/new", walletHandlers.CreateWallet).Methods("POST")
	walletRouter.HandleFunc("/{id}", walletHandlers.GetWallet).Methods("GET")
	
	// Document routes
	documentRouter := apiRouter.PathPrefix("/document").Subrouter()
	documentRouter.HandleFunc("/sign", documentHandlers.SignDocument).Methods("POST")
	documentRouter.HandleFunc("/{id}", documentHandlers.GetDocument).Methods("GET")
	documentRouter.HandleFunc("/verify/{id}", documentHandlers.VerifyDocument).Methods("GET")
	
	// Blockchain routes
	blockchainRouter := apiRouter.PathPrefix("/blockchain").Subrouter()
	blockchainRouter.HandleFunc("", blockchainHandlers.GetBlockchainInfo).Methods("GET")
	blockchainRouter.HandleFunc("/mine", blockchainHandlers.MineBlock).Methods("POST")
	blockchainRouter.HandleFunc("/blocks", blockchainHandlers.GetBlocks).Methods("GET")
	blockchainRouter.HandleFunc("/block/{hash}", blockchainHandlers.GetBlockByHash).Methods("GET")
	blockchainRouter.HandleFunc("/pending", blockchainHandlers.GetPendingDocuments).Methods("GET")
	blockchainRouter.HandleFunc("/validate", blockchainHandlers.ValidateChain).Methods("GET")
	
	// Apply middleware
	s.Router.Use(LoggingMiddleware)
	s.Router.Use(CORSMiddleware)
	s.Router.Use(RecoverMiddleware)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.Port)
	log.Printf("Server starting on port %s\n", s.Port)
	return http.ListenAndServe(addr, s.Router)
}