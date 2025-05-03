package database

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raykavin/doc-chain/internal/crypto"
)

// WalletRepository handles wallet operations with the database
type WalletRepository struct {
	db *Database
}

// Wallet represents a wallet in the database
type Wallet struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
	PrivateKey string `json:"-"` // Not returned in JSON
	CreatedAt time.Time `json:"createdAt"`
}

// Certificate represents a certificate in the database
type Certificate struct {
	ID           string    `json:"id"`
	WalletID     string    `json:"walletId"`
	Type         string    `json:"type"`
	Data         []byte    `json:"-"` // Not returned in JSON
	PasswordHash string    `json:"-"` // Not returned in JSON
	CreatedAt    time.Time `json:"createdAt"`
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *Database) *WalletRepository {
	return &WalletRepository{
		db: db,
	}
}

// CreateWallet creates a new wallet in the database
func (r *WalletRepository) CreateWallet(name string) (*Wallet, error) {
	// Create a new wallet
	wallet, err := crypto.NewWallet()
	if err != nil {
		return nil, fmt.Errorf("falha ao criar carteira: %w", err)
	}

	// Export the private key
	privateKeyPEM, err := wallet.ExportPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("falha ao exportar chave privada: %w", err)
	}

	// Generate a unique ID
	id := uuid.New().String()

	// Get the public key as string
	publicKey := wallet.GetPublicKeyAsString()

	// Get the current time
	now := time.Now()

	// Insert the wallet into the database
	_, err = r.db.GetDB().Exec(
		"INSERT INTO wallets (id, name, public_key, private_key, created_at) VALUES (?, ?, ?, ?, ?)",
		id, name, publicKey, privateKeyPEM, now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao inserir carteira no banco de dados: %w", err)
	}

	// Return the wallet
	return &Wallet{
		ID:        id,
		Name:      name,
		PublicKey: publicKey,
		PrivateKey: privateKeyPEM,
		CreatedAt: now,
	}, nil
}

// GetWallet gets a wallet from the database
func (r *WalletRepository) GetWallet(id string) (*Wallet, error) {
	// Query the database
	row := r.db.GetDB().QueryRow(
		"SELECT id, name, public_key, private_key, created_at FROM wallets WHERE id = ?",
		id,
	)

	// Scan the result
	var wallet Wallet
	var createdAtStr string
	err := row.Scan(&wallet.ID, &wallet.Name, &wallet.PublicKey, &wallet.PrivateKey, &createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter carteira: %w", err)
	}

	// Parse the created at time
	wallet.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao analisar data de criação: %w", err)
	}

	return &wallet, nil
}

// GetWallets gets all wallets from the database
func (r *WalletRepository) GetWallets() ([]*Wallet, error) {
	// Query the database
	rows, err := r.db.GetDB().Query(
		"SELECT id, name, public_key, created_at FROM wallets",
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter carteiras: %w", err)
	}
	defer rows.Close()

	// Scan the results
	var wallets []*Wallet
	for rows.Next() {
		var wallet Wallet
		var createdAtStr string
		err := rows.Scan(&wallet.ID, &wallet.Name, &wallet.PublicKey, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("falha ao analisar carteira: %w", err)
		}

		// Parse the created at time
		wallet.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("falha ao analisar data de criação: %w", err)
		}

		wallets = append(wallets, &wallet)
	}

	return wallets, nil
}

// CreateCertificate creates a new certificate in the database
func (r *WalletRepository) CreateCertificate(walletID, certType string, certData []byte, password string) (*Certificate, error) {
	// Generate a unique ID
	id := uuid.New().String()

	// Hash the password
	passwordHash := hashPassword(password)

	// Get the current time
	now := time.Now()

	// Insert the certificate into the database
	_, err := r.db.GetDB().Exec(
		"INSERT INTO certificates (id, wallet_id, cert_type, cert_data, password_hash, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, walletID, certType, certData, passwordHash, now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao inserir certificado no banco de dados: %w", err)
	}

	// Return the certificate
	return &Certificate{
		ID:           id,
		WalletID:     walletID,
		Type:         certType,
		Data:         certData,
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}, nil
}

// GetCertificate gets a certificate from the database
func (r *WalletRepository) GetCertificate(walletID string) (*Certificate, error) {
	// Query the database
	row := r.db.GetDB().QueryRow(
		"SELECT id, wallet_id, cert_type, cert_data, password_hash, created_at FROM certificates WHERE wallet_id = ?",
		walletID,
	)

	// Scan the result
	var cert Certificate
	var createdAtStr string
	err := row.Scan(&cert.ID, &cert.WalletID, &cert.Type, &cert.Data, &cert.PasswordHash, &createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter certificado: %w", err)
	}

	// Parse the created at time
	cert.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao analisar data de criação: %w", err)
	}

	return &cert, nil
}

// VerifyCertificatePassword verifies a certificate password
func (r *WalletRepository) VerifyCertificatePassword(certID, password string) (bool, error) {
	// Query the database
	row := r.db.GetDB().QueryRow(
		"SELECT password_hash FROM certificates WHERE id = ?",
		certID,
	)

	// Scan the result
	var passwordHash string
	err := row.Scan(&passwordHash)
	if err != nil {
		return false, fmt.Errorf("falha ao obter hash da senha: %w", err)
	}

	// Verify the password
	return passwordHash == hashPassword(password), nil
}

// hashPassword hashes a password
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
