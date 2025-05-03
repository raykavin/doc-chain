package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/raykavin/doc-chain/internal/blockchain"
	"github.com/raykavin/doc-chain/internal/crypto"
	"github.com/raykavin/doc-chain/internal/database"
)

// Server represents the blockchain API server
type Server struct {
	Blockchain *blockchain.Blockchain
	Wallets    map[string]*crypto.Wallet
	Port       string
	Router     *mux.Router
	DB         *database.Database

	// Repositories
	WalletRepo   *database.WalletRepository
	DocumentRepo *database.DocumentRepository
}

// NewServer creates a new blockchain API server
func NewServer(port string, difficulty int, dbPath string) (*Server, error) {
	// Create the database connection
	db, err := database.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Create the repositories
	walletRepo, documentRepo, err := db.GetRepositories()
	if err != nil {
		return nil, fmt.Errorf("failed to create repositories: %w", err)
	}

	return &Server{
		Blockchain:   blockchain.NewBlockchain(difficulty),
		Wallets:      make(map[string]*crypto.Wallet),
		Port:         port,
		Router:       mux.NewRouter(),
		DB:           db,
		WalletRepo:   walletRepo,
		DocumentRepo: documentRepo,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.Port)
	log.Printf("Server starting on port %s\n", s.Port)
	return http.ListenAndServe(addr, s.Router)
}

// Configure sets up the server routes and middleware
func (s *Server) Configure() {
	// Setup middleware
	s.setupMiddleware()

	// Setup API routes FIRST - this is critical for routing order
	apiRouter := s.Router.PathPrefix("/api").Subrouter()
	s.setupWalletRoutes(apiRouter)
	s.setupDocumentRoutes(apiRouter)
	s.setupBlockchainRoutes(apiRouter)
	s.setupICPBrasilRoutes(apiRouter)

	// Setup static file server AFTER API routes
	s.setupStaticFileServer()
}

// setupMiddleware configures middleware for the server
func (s *Server) setupMiddleware() {
	s.Router.Use(LoggingMiddleware)
	s.Router.Use(CORSMiddleware)
	s.Router.Use(RecoverMiddleware)
}

// setupStaticFileServer configures static file serving
func (s *Server) setupStaticFileServer() {
	// Create file server
	fs := http.FileServer(http.Dir("./static"))

	// Handle the root path specifically
	s.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// Handle all other non-API static files
	// This should come AFTER the API routes in the Configure method
	s.Router.PathPrefix("/").Handler(fs)
}

// setupWalletRoutes configures wallet-related API routes
func (s *Server) setupWalletRoutes(apiRouter *mux.Router) {
	walletHandlers := &WalletHandlersDB{server: s}
	walletRouter := apiRouter.PathPrefix("/wallet").Subrouter()

	walletRouter.HandleFunc("/new", walletHandlers.CreateWallet).Methods(http.MethodPost)
	walletRouter.HandleFunc("/{id}", walletHandlers.GetWallet).Methods(http.MethodGet)
	walletRouter.HandleFunc("/list", walletHandlers.ListWallets).Methods(http.MethodGet)

	// Note: createCertificate handler moved to the proper ICP section
}

// setupDocumentRoutes configures document-related API routes
func (s *Server) setupDocumentRoutes(apiRouter *mux.Router) {
	documentHandlers := &DocumentHandlersDB{server: s}
	documentRouter := apiRouter.PathPrefix("/document").Subrouter()

	documentRouter.HandleFunc("/sign", documentHandlers.SignDocument).Methods(http.MethodPost)
	documentRouter.HandleFunc("/{id}", documentHandlers.GetDocument).Methods(http.MethodGet)
	documentRouter.HandleFunc("/verify/{id}", documentHandlers.VerifyDocument).Methods(http.MethodGet)
	documentRouter.HandleFunc("/list", documentHandlers.ListDocuments).Methods(http.MethodGet)
	documentRouter.HandleFunc("/upload", documentHandlers.UploadDocument).Methods(http.MethodPost)
	documentRouter.HandleFunc("/download/{id}", documentHandlers.DownloadDocument).Methods(http.MethodGet)
}

// setupBlockchainRoutes configures blockchain-related API routes
func (s *Server) setupBlockchainRoutes(apiRouter *mux.Router) {
	blockchainHandlers := &BlockchainHandlers{server: s}
	blockchainRouter := apiRouter.PathPrefix("/blockchain").Subrouter()

	blockchainRouter.HandleFunc("", blockchainHandlers.GetBlockchainInfo).Methods(http.MethodGet)
	blockchainRouter.HandleFunc("/mine", blockchainHandlers.MineBlock).Methods(http.MethodPost)
	blockchainRouter.HandleFunc("/blocks", blockchainHandlers.GetBlocks).Methods(http.MethodGet)
	blockchainRouter.HandleFunc("/block/{hash}", blockchainHandlers.GetBlockByHash).Methods(http.MethodGet)
	blockchainRouter.HandleFunc("/pending", blockchainHandlers.GetPendingDocuments).Methods(http.MethodGet)
	blockchainRouter.HandleFunc("/validate", blockchainHandlers.ValidateChain).Methods(http.MethodGet)
}

// setupICPBrasilRoutes configures ICP-Brasil specific API routes
func (s *Server) setupICPBrasilRoutes(apiRouter *mux.Router) {
	icpbrasilHandlers := &ICPBrasilHandlers{server: s}
	walletHandlers := &WalletHandlersDB{server: s}
	icpbrasilRouter := apiRouter.PathPrefix("/icpbrasil").Subrouter()

	icpbrasilRouter.HandleFunc("/sign", icpbrasilHandlers.SignDocument).Methods(http.MethodPost)
	icpbrasilRouter.HandleFunc("/verify", icpbrasilHandlers.VerifySignature).Methods(http.MethodPost)
	icpbrasilRouter.HandleFunc("/cert-info", icpbrasilHandlers.GetCertificateInfo).Methods(http.MethodPost)
	icpbrasilRouter.HandleFunc("/check-revocation", icpbrasilHandlers.CheckRevocation).Methods(http.MethodPost)
	icpbrasilRouter.HandleFunc("/document/{id}", icpbrasilHandlers.GetDocumentInfo).Methods(http.MethodGet)
	icpbrasilRouter.HandleFunc("/create-certificate", walletHandlers.CreateCertificate).Methods(http.MethodPost)
}
