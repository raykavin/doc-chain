package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
)

// Signature represents a digital signature
type Signature struct {
	R *big.Int
	S *big.Int
}

// VerifySignature verifies if the signature is valid for a document
func VerifySignature(content string, signature *Signature, publicKey *ecdsa.PublicKey) bool {
	contentHash := sha256.Sum256([]byte(content))
	return ecdsa.Verify(publicKey, contentHash[:], signature.R, signature.S)
}

// VerifySignatureHex verifies if the signature is valid using hex strings
func VerifySignatureHex(contentHex string, signature *Signature, publicKeyHex string) bool {
	contentBytes, err := hex.DecodeString(contentHex)
	if err != nil {
		return false
	}
	
	publicKey, err := ParsePublicKeyHex(publicKeyHex)
	if err != nil {
		return false
	}
	
	return ecdsa.Verify(publicKey, contentBytes, signature.R, signature.S)
}

// NewBigInt creates a new big.Int from an int64
func NewBigInt(value int64) *big.Int {
	return big.NewInt(value)
}