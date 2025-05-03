package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents a database connection
type Database struct {
	db *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*Database, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create database instance
	database := &Database{
		db: db,
	}

	// Initialize database
	if err := database.initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// GetDB returns the underlying database connection
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// GetRepositories returns the repositories for the database
func (d *Database) GetRepositories() (*WalletRepository, *DocumentRepository, error) {
	walletRepo := NewWalletRepository(d)
	documentRepo := NewDocumentRepository(d)
	return walletRepo, documentRepo, nil
}

// initialize creates the database tables if they don't exist
func (d *Database) initialize() error {
	// Create wallets table
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS wallets (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			public_key TEXT NOT NULL,
			private_key TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create wallets table: %w", err)
	}

	// Create certificates table
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS certificates (
			id TEXT PRIMARY KEY,
			wallet_id TEXT NOT NULL,
			cert_type TEXT NOT NULL,
			cert_data BLOB NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (wallet_id) REFERENCES wallets (id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create certificates table: %w", err)
	}

	// Create documents table
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content_type TEXT NOT NULL,
			content BLOB NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create documents table: %w", err)
	}

	// Create signatures table
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS signatures (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			wallet_id TEXT NOT NULL,
			signature BLOB NOT NULL,
			signature_hash TEXT NOT NULL,
			certificate_hash TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			has_timestamp BOOLEAN NOT NULL,
			block_index INTEGER,
			created_at TEXT NOT NULL,
			FOREIGN KEY (document_id) REFERENCES documents (id),
			FOREIGN KEY (wallet_id) REFERENCES wallets (id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create signatures table: %w", err)
	}

	// Create templates table
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS templates (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			content_type TEXT NOT NULL,
			content BLOB NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create templates table: %w", err)
	}

	return nil
}
