package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

// Wallet represents a user's wallet for signing documents
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

// NewWallet creates a new wallet with a key pair
func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	
	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// GetPublicKeyAsString returns the public key as a hexadecimal string
func (w *Wallet) GetPublicKeyAsString() string {
	x := w.PublicKey.X.Bytes()
	y := w.PublicKey.Y.Bytes()
	return hex.EncodeToString(append(x, y...))
}

// ExportPrivateKey exports the private key in PEM format
func (w *Wallet) ExportPrivateKey() (string, error) {
	privateKeyBytes, err := x509.MarshalECPrivateKey(w.PrivateKey)
	if err != nil {
		return "", err
	}
	
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	
	return string(privateKeyPEM), nil
}

// ImportPrivateKey imports a private key from PEM format
func ImportPrivateKey(pemString string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	
	return privateKey, nil
}

// SignData signs data with the user's private key
func (w *Wallet) SignData(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, data)
	if err != nil {
		return nil, err
	}
	
	return &Signature{
		R: r,
		S: s,
	}, nil
}

// SignString signs a string with the user's private key
func (w *Wallet) SignString(content string) (*Signature, error) {
	contentHash := sha256.Sum256([]byte(content))
	return w.SignData(contentHash[:])
}

// ParsePublicKeyHex reconstructs a public key from a hexadecimal string
func ParsePublicKeyHex(publicKeyHex string) (*ecdsa.PublicKey, error) {
	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, err
	}
	
	if len(publicKeyBytes) == 0 {
		return nil, errors.New("empty public key")
	}
	
	// Reconstruct public key
	keyLength := len(publicKeyBytes) / 2
	if keyLength == 0 {
		return nil, errors.New("invalid public key format")
	}
	
	x := new(big.Int).SetBytes(publicKeyBytes[:keyLength])
	y := new(big.Int).SetBytes(publicKeyBytes[keyLength:])
	
	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	
	return pubKey, nil
}