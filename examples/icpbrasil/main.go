package main

import (
	"fmt"
	"log"
	"os"

	"github.com/raykavin/doc-chain/internal/blockchain"
	"github.com/raykavin/doc-chain/internal/icpbrasil"
)

func main() {
	// Check command line arguments
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run icpbrasil_example.go <document_path> <certificate_path> <certificate_password>")
		os.Exit(1)
	}

	documentPath := os.Args[1]
	certificatePath := os.Args[2]
	certificatePassword := os.Args[3]

	// Create a blockchain instance
	bc := blockchain.NewBlockchain(4) // Difficulty level 4

	// Create a document manager
	dm := icpbrasil.NewDocumentManager(bc)

	// Example 1: Sign a document
	fmt.Println("=== Example 1: Sign a document ===")
	signExample(dm, documentPath, certificatePath, certificatePassword)

	// Example 2: Verify a signature
	fmt.Println("\n=== Example 2: Verify a signature ===")
	verifyExample(dm, documentPath, "signature.p7s")

	// Example 3: Get certificate information
	fmt.Println("\n=== Example 3: Get certificate information ===")
	certInfoExample(certificatePath, certificatePassword)

	// Example 4: Check certificate revocation
	fmt.Println("\n=== Example 4: Check certificate revocation ===")
	checkRevocationExample(dm, certificatePath, certificatePassword)

	// Example 5: Sign and add to blockchain
	fmt.Println("\n=== Example 5: Sign and add to blockchain ===")
	signAndAddToBlockchainExample(dm, documentPath, certificatePath, certificatePassword)

	// Example 6: Mine a block
	fmt.Println("\n=== Example 6: Mine a block ===")
	mineBlockExample(bc)

	// Example 7: Verify document in blockchain
	fmt.Println("\n=== Example 7: Verify document in blockchain ===")
	verifyBlockchainExample(dm, documentPath, "signature.p7s")
}

// signExample demonstrates how to sign a document
func signExample(dm *icpbrasil.DocumentManager, documentPath, certificatePath, certificatePassword string) {
	// Create signature options
	options := icpbrasil.DefaultSignatureOptions()
	options.Detached = true
	options.IncludeChain = true
	options.AddTimestamp = false

	// Sign the document
	signedDoc, err := dm.SignDocumentWithCertificate(
		documentPath,
		certificatePath,
		certificatePassword,
		options,
	)
	if err != nil {
		log.Fatalf("Failed to sign document: %v", err)
	}

	// Save the signature
	if err := signedDoc.SaveSignature("signature.p7s"); err != nil {
		log.Fatalf("Failed to save signature: %v", err)
	}

	// Get signature information
	info := signedDoc.GetSignatureInfo()
	fmt.Println("Document signed successfully!")
	fmt.Printf("Document Hash: %s\n", info["DocumentHash"])
	fmt.Printf("Signature Hash: %s\n", info["SignatureHash"])
	fmt.Printf("Signer Certificate Hash: %s\n", info["SignerCertificateHash"])
	fmt.Printf("Signature Time: %s\n", info["SignatureTime"])
}

// verifyExample demonstrates how to verify a signature
func verifyExample(dm *icpbrasil.DocumentManager, documentPath, signaturePath string) {
	// Verify the signature
	result, err := dm.VerifySignature(documentPath, signaturePath)
	if err != nil {
		log.Fatalf("Failed to verify signature: %v", err)
	}

	// Print verification result
	if result.IsValid {
		fmt.Println("✅ Document signature is VALID")
	} else {
		fmt.Println("❌ Document signature is INVALID")
		for _, errMsg := range result.Errors {
			fmt.Printf("  - %s\n", errMsg)
		}
		return
	}

	// Print signer information
	if result.SignerCert != nil {
		fmt.Printf("Signer: %s\n", result.SignerCert.Subject.CommonName)
		fmt.Printf("Issuer: %s\n", result.SignerCert.Issuer.CommonName)
		fmt.Printf("Serial Number: %s\n", result.SignerCert.SerialNumber)
		fmt.Printf("Valid From: %s\n", result.SignerCert.NotBefore)
		fmt.Printf("Valid To: %s\n", result.SignerCert.NotAfter)
	}
}

// certInfoExample demonstrates how to get certificate information
func certInfoExample(certificatePath, certificatePassword string) {
	// Load the certificate
	cert, err := icpbrasil.LoadPFXFromFile(certificatePath, certificatePassword)
	if err != nil {
		log.Fatalf("Failed to load certificate: %v", err)
	}

	// Get certificate information
	info := cert.GetCertificateInfo()
	fmt.Println("Certificate Information:")
	fmt.Printf("Subject: %s\n", info["Subject"])
	fmt.Printf("Issuer: %s\n", info["Issuer"])
	fmt.Printf("Serial Number: %s\n", info["SerialNumber"])
	fmt.Printf("Valid From: %s\n", info["NotBefore"])
	fmt.Printf("Valid To: %s\n", info["NotAfter"])
	fmt.Printf("Is ICP-Brasil: %s\n", info["IsICPBrasil"])
}

// checkRevocationExample demonstrates how to check certificate revocation
func checkRevocationExample(dm *icpbrasil.DocumentManager, certificatePath, certificatePassword string) {
	// Check revocation
	isRevoked, err := dm.CheckRevocation(certificatePath, certificatePassword)
	if err != nil {
		log.Printf("Failed to check revocation: %v", err)
		return
	}

	// Print result
	if isRevoked {
		fmt.Println("❌ Certificate is REVOKED")
	} else {
		fmt.Println("✅ Certificate is NOT revoked")
	}
}

// signAndAddToBlockchainExample demonstrates how to sign a document and add it to the blockchain
func signAndAddToBlockchainExample(dm *icpbrasil.DocumentManager, documentPath, certificatePath, certificatePassword string) {
	// Create signature options
	options := icpbrasil.DefaultSignatureOptions()
	options.Detached = true
	options.IncludeChain = true
	options.AddTimestamp = false

	// Sign and add to blockchain
	bcDoc, err := dm.SignAndAddToBlockchain(
		documentPath,
		certificatePath,
		certificatePassword,
		options,
	)
	if err != nil {
		log.Fatalf("Failed to sign and add to blockchain: %v", err)
	}

	// Save the signature
	if err := bcDoc.SignedDocument.SaveSignature("signature.p7s"); err != nil {
		log.Fatalf("Failed to save signature: %v", err)
	}

	fmt.Println("Document signed and added to blockchain successfully!")
	fmt.Printf("Document Hash: %s\n", bcDoc.DocumentHash)
	fmt.Printf("Signature Hash: %s\n", bcDoc.SignatureHash)
	fmt.Printf("Signer Certificate Hash: %s\n", bcDoc.CertificateHash)
	fmt.Printf("Signer Name: %s\n", bcDoc.SignerName)
	fmt.Printf("Signer ID: %s\n", bcDoc.SignerID)
	fmt.Printf("Signature Time: %s\n", bcDoc.SignatureTime)
	fmt.Println("Status: Pending (not yet in a block)")
}

// mineBlockExample demonstrates how to mine a block
func mineBlockExample(bc *blockchain.Blockchain) {
	// Mine a block
	block, err := bc.MineBlock()
	if err != nil {
		log.Fatalf("Failed to mine block: %v", err)
	}

	fmt.Println("Block mined successfully!")
	fmt.Printf("Block #%d\n", block.Index)
	fmt.Printf("Hash: %s\n", block.Hash)
	fmt.Printf("Previous Hash: %s\n", block.PrevHash)
	fmt.Printf("Merkle Root: %s\n", block.MerkleRoot)
	fmt.Printf("Timestamp: %d\n", block.Timestamp)
	fmt.Printf("Nonce: %d\n", block.Nonce)
	fmt.Printf("Documents: %d\n", len(block.Documents))
}

// verifyBlockchainExample demonstrates how to verify a document in the blockchain
func verifyBlockchainExample(dm *icpbrasil.DocumentManager, documentPath, signaturePath string) {
	// Verify signature and check blockchain
	bcValid, result, err := dm.VerifySignatureAndBlockchain(documentPath, signaturePath)
	if err != nil {
		if err.Error()[:22] == "blockchain verification" {
			fmt.Println("⚠️ Document not found in blockchain")
		} else {
			log.Fatalf("Failed to verify signature: %v", err)
		}
		return
	}

	// Print verification result
	if result.IsValid {
		fmt.Println("✅ Document signature is VALID")
	} else {
		fmt.Println("❌ Document signature is INVALID")
		for _, errMsg := range result.Errors {
			fmt.Printf("  - %s\n", errMsg)
		}
		return
	}

	// Print blockchain verification result
	if bcValid {
		fmt.Println("✅ Document is verified in the blockchain")
	} else {
		fmt.Println("❌ Document is not verified in the blockchain")
	}
}
