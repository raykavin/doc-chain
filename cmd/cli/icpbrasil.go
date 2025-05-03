package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/raykavin/doc-chain/internal/blockchain"
	"github.com/raykavin/doc-chain/internal/icpbrasil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newICPBrasilCmd() *cobra.Command {
	icpbrasilCmd := &cobra.Command{
		Use:   "icpbrasil",
		Short: "ICP-Brasil digital signature operations",
		Long:  `Sign, verify, and manage documents using ICP-Brasil digital certificates.`,
	}

	icpbrasilCmd.AddCommand(newICPBrasilSignCmd())
	icpbrasilCmd.AddCommand(newICPBrasilVerifyCmd())
	icpbrasilCmd.AddCommand(newICPBrasilCertInfoCmd())
	icpbrasilCmd.AddCommand(newICPBrasilCheckRevocationCmd())
	icpbrasilCmd.AddCommand(newICPBrasilDocInfoCmd())

	return icpbrasilCmd
}

func newICPBrasilSignCmd() *cobra.Command {
	var (
		documentPath        string
		certificatePath     string
		certificatePassword string
		outputDir           string
		detached            bool
		includeChain        bool
		addTimestamp        bool
		tsaUrl              string
		addToBlockchain     bool
		mineBlock           bool
	)

	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign a document with an ICP-Brasil certificate",
		Long:  `Sign a document using an ICP-Brasil A1 certificate (.pfx or .p12).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate inputs
			if documentPath == "" {
				return fmt.Errorf("document path is required")
			}
			if certificatePath == "" {
				return fmt.Errorf("certificate path is required")
			}

			// Check if the document exists
			if _, err := os.Stat(documentPath); os.IsNotExist(err) {
				return fmt.Errorf("document not found: %s", documentPath)
			}

			// Check if the certificate exists
			if _, err := os.Stat(certificatePath); os.IsNotExist(err) {
				return fmt.Errorf("certificate not found: %s", certificatePath)
			}

			// Set default output directory if not provided
			if outputDir == "" {
				outputDir = filepath.Dir(documentPath)
			}

			// Create signature options
			options := &icpbrasil.SignatureOptions{
				Detached:     detached,
				IncludeChain: includeChain,
				AddTimestamp: addTimestamp,
				TSAUrl:       tsaUrl,
			}

			// Create a blockchain instance
			bc := getBlockchain()

			// Create a document manager
			dm := icpbrasil.NewDocumentManager(bc)

			var signedDoc *icpbrasil.SignedDocument
			var bcDoc *icpbrasil.BlockchainDocument
			var err error

			if addToBlockchain {
				// Sign and add to blockchain
				bcDoc, err = dm.SignAndAddToBlockchain(
					documentPath,
					certificatePath,
					certificatePassword,
					options,
				)
				if err != nil {
					return fmt.Errorf("failed to sign and add to blockchain: %w", err)
				}
				signedDoc = bcDoc.SignedDocument
			} else {
				// Just sign the document
				signedDoc, err = dm.SignDocumentWithCertificate(
					documentPath,
					certificatePath,
					certificatePassword,
					options,
				)
				if err != nil {
					return fmt.Errorf("failed to sign document: %w", err)
				}
			}

			// Save the signed document
			baseFilename := filepath.Base(documentPath)
			docPath, sigPath, err := dm.SaveSignedDocument(signedDoc, outputDir, baseFilename)
			if err != nil {
				return fmt.Errorf("failed to save signed document: %w", err)
			}

			// Print success message
			fmt.Println("Document signed successfully!")
			fmt.Printf("Document: %s\n", docPath)
			fmt.Printf("Signature: %s\n", sigPath)

			// Print signature information
			info := signedDoc.GetSignatureInfo()
			fmt.Println("\nSignature Information:")
			fmt.Printf("Document Hash: %s\n", info["DocumentHash"])
			fmt.Printf("Signature Hash: %s\n", info["SignatureHash"])
			fmt.Printf("Signer Certificate Hash: %s\n", info["SignerCertificateHash"])
			fmt.Printf("Signature Time: %s\n", info["SignatureTime"])
			fmt.Printf("Has Timestamp: %s\n", info["HasTimestamp"])
			if info["HasTimestamp"] == "true" {
				fmt.Printf("Timestamp Time: %s\n", info["TimestampTime"])
			}

			// Print blockchain information if added
			if addToBlockchain {
				fmt.Println("\nDocument added to blockchain.")
				fmt.Println("Status: Pending (not yet in a block)")

				// Mine a block if requested
				if mineBlock {
					fmt.Println("\nMining a new block...")
					block, err := bc.MineBlock()
					if err != nil {
						return fmt.Errorf("failed to mine block: %w", err)
					}
					fmt.Printf("Block mined successfully! Block #%d (Hash: %s)\n", block.Index, block.Hash)
				} else {
					fmt.Println("\nUse 'blockchain-cli blockchain mine' to include the document in a block.")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&documentPath, "document", "", "path to the document to sign")
	cmd.Flags().StringVar(&certificatePath, "certificate", "", "path to the ICP-Brasil certificate (.pfx or .p12)")
	cmd.Flags().StringVar(&certificatePassword, "password", "", "password for the certificate")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "directory to save the signed document and signature")
	cmd.Flags().BoolVar(&detached, "detached", true, "create a detached signature")
	cmd.Flags().BoolVar(&includeChain, "include-chain", true, "include the certificate chain in the signature")
	cmd.Flags().BoolVar(&addTimestamp, "timestamp", false, "add a timestamp to the signature")
	cmd.Flags().StringVar(&tsaUrl, "tsa-url", "http://timestamping.iti.gov.br/tsa", "URL of the Time Stamping Authority")
	cmd.Flags().BoolVar(&addToBlockchain, "blockchain", true, "add the document to the blockchain")
	cmd.Flags().BoolVar(&mineBlock, "mine", false, "mine a block after adding the document to the blockchain")

	cmd.MarkFlagRequired("document")
	cmd.MarkFlagRequired("certificate")
	cmd.MarkFlagRequired("password")

	return cmd
}

func newICPBrasilVerifyCmd() *cobra.Command {
	var (
		documentPath  string
		signaturePath string
		checkBlockchain bool
	)

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify a document signature",
		Long:  `Verify a document signature created with an ICP-Brasil certificate.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate inputs
			if documentPath == "" {
				return fmt.Errorf("document path is required")
			}
			if signaturePath == "" {
				return fmt.Errorf("signature path is required")
			}

			// Check if the document exists
			if _, err := os.Stat(documentPath); os.IsNotExist(err) {
				return fmt.Errorf("document not found: %s", documentPath)
			}

			// Check if the signature exists
			if _, err := os.Stat(signaturePath); os.IsNotExist(err) {
				return fmt.Errorf("signature not found: %s", signaturePath)
			}

			// Create a blockchain instance
			bc := getBlockchain()

			// Create a document manager
			dm := icpbrasil.NewDocumentManager(bc)

			var result *icpbrasil.VerificationResult
			var bcValid bool
			var err error

			if checkBlockchain {
				// Verify signature and check blockchain
				bcValid, result, err = dm.VerifySignatureAndBlockchain(documentPath, signaturePath)
				if err != nil {
					if strings.Contains(err.Error(), "blockchain verification failed") {
						// Document might be valid but not in the blockchain yet
						fmt.Println("⚠️ Document not found in blockchain")
					} else {
						return fmt.Errorf("verification failed: %w", err)
					}
				}
			} else {
				// Just verify the signature
				result, err = dm.VerifySignature(documentPath, signaturePath)
				if err != nil {
					return fmt.Errorf("verification failed: %w", err)
				}
			}

			// Print verification result
			if result.IsValid {
				fmt.Println("✅ Document signature is VALID")
			} else {
				fmt.Println("❌ Document signature is INVALID")
				for _, errMsg := range result.Errors {
					fmt.Printf("  - %s\n", errMsg)
				}
				return nil
			}

			// Print signer information
			if result.SignerCert != nil {
				fmt.Printf("\nSigner: %s\n", result.SignerCert.Subject.CommonName)
				fmt.Printf("Issuer: %s\n", result.SignerCert.Issuer.CommonName)
				fmt.Printf("Serial Number: %s\n", result.SignerCert.SerialNumber)
				fmt.Printf("Valid From: %s\n", result.SignerCert.NotBefore.Format(time.RFC3339))
				fmt.Printf("Valid To: %s\n", result.SignerCert.NotAfter.Format(time.RFC3339))
			}

			// Print blockchain verification result
			if checkBlockchain {
				if bcValid {
					fmt.Println("\n✅ Document is verified in the blockchain")
				} else {
					fmt.Println("\n❌ Document is not verified in the blockchain")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&documentPath, "document", "", "path to the document to verify")
	cmd.Flags().StringVar(&signaturePath, "signature", "", "path to the signature file")
	cmd.Flags().BoolVar(&checkBlockchain, "blockchain", false, "check if the document is in the blockchain")

	cmd.MarkFlagRequired("document")
	cmd.MarkFlagRequired("signature")

	return cmd
}

func newICPBrasilCertInfoCmd() *cobra.Command {
	var (
		certificatePath     string
		certificatePassword string
		outputFile          string
	)

	cmd := &cobra.Command{
		Use:   "cert-info",
		Short: "Get information about an ICP-Brasil certificate",
		Long:  `Display information about an ICP-Brasil A1 certificate (.pfx or .p12).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate inputs
			if certificatePath == "" {
				return fmt.Errorf("certificate path is required")
			}

			// Check if the certificate exists
			if _, err := os.Stat(certificatePath); os.IsNotExist(err) {
				return fmt.Errorf("certificate not found: %s", certificatePath)
			}

			// Load the certificate
			cert, err := icpbrasil.LoadPFXFromFile(certificatePath, certificatePassword)
			if err != nil {
				return fmt.Errorf("failed to load certificate: %w", err)
			}

			// Get certificate information
			info := cert.GetCertificateInfo()

			// Print certificate information
			fmt.Println("Certificate Information:")
			fmt.Printf("Subject: %s\n", info["Subject"])
			fmt.Printf("Issuer: %s\n", info["Issuer"])
			fmt.Printf("Serial Number: %s\n", info["SerialNumber"])
			fmt.Printf("Valid From: %s\n", info["NotBefore"])
			fmt.Printf("Valid To: %s\n", info["NotAfter"])
			fmt.Printf("Is ICP-Brasil: %s\n", info["IsICPBrasil"])

			// Save to file if requested
			if outputFile != "" {
				// Convert info to JSON
				jsonData, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to convert to JSON: %w", err)
				}

				// Write to file
				if err := ioutil.WriteFile(outputFile, jsonData, 0644); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}

				fmt.Printf("\nCertificate information saved to %s\n", outputFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&certificatePath, "certificate", "", "path to the ICP-Brasil certificate (.pfx or .p12)")
	cmd.Flags().StringVar(&certificatePassword, "password", "", "password for the certificate")
	cmd.Flags().StringVar(&outputFile, "output", "", "path to save the certificate information as JSON")

	cmd.MarkFlagRequired("certificate")
	cmd.MarkFlagRequired("password")

	return cmd
}

func newICPBrasilCheckRevocationCmd() *cobra.Command {
	var (
		certificatePath     string
		certificatePassword string
	)

	cmd := &cobra.Command{
		Use:   "check-revocation",
		Short: "Check if an ICP-Brasil certificate is revoked",
		Long:  `Check if an ICP-Brasil A1 certificate (.pfx or .p12) is revoked using CRL and OCSP.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate inputs
			if certificatePath == "" {
				return fmt.Errorf("certificate path is required")
			}

			// Check if the certificate exists
			if _, err := os.Stat(certificatePath); os.IsNotExist(err) {
				return fmt.Errorf("certificate not found: %s", certificatePath)
			}

			// Create a blockchain instance
			bc := getBlockchain()

			// Create a document manager
			dm := icpbrasil.NewDocumentManager(bc)

			// Check revocation
			isRevoked, err := dm.CheckRevocation(certificatePath, certificatePassword)
			if err != nil {
				return fmt.Errorf("failed to check revocation: %w", err)
			}

			// Print result
			if isRevoked {
				fmt.Println("❌ Certificate is REVOKED")
			} else {
				fmt.Println("✅ Certificate is NOT revoked")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&certificatePath, "certificate", "", "path to the ICP-Brasil certificate (.pfx or .p12)")
	cmd.Flags().StringVar(&certificatePassword, "password", "", "password for the certificate")

	cmd.MarkFlagRequired("certificate")
	cmd.MarkFlagRequired("password")

	return cmd
}

func newICPBrasilDocInfoCmd() *cobra.Command {
	var (
		documentPath  string
		signaturePath string
		outputFile    string
	)

	cmd := &cobra.Command{
		Use:   "doc-info",
		Short: "Get information about a signed document",
		Long:  `Display information about a document signed with an ICP-Brasil certificate.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate inputs
			if documentPath == "" {
				return fmt.Errorf("document path is required")
			}
			if signaturePath == "" {
				return fmt.Errorf("signature path is required")
			}

			// Check if the document exists
			if _, err := os.Stat(documentPath); os.IsNotExist(err) {
				return fmt.Errorf("document not found: %s", documentPath)
			}

			// Check if the signature exists
			if _, err := os.Stat(signaturePath); os.IsNotExist(err) {
				return fmt.Errorf("signature not found: %s", signaturePath)
			}

			// Create a blockchain instance
			bc := getBlockchain()

			// Create a document manager
			dm := icpbrasil.NewDocumentManager(bc)

			// Get document information
			info, err := dm.GetDocumentInfo(documentPath, signaturePath)
			if err != nil {
				return fmt.Errorf("failed to get document information: %w", err)
			}

			// Print document information
			fmt.Println("Document Information:")
			fmt.Printf("Document Hash: %s\n", info["DocumentHash"])
			fmt.Printf("Signature Hash: %s\n", info["SignatureHash"])
			fmt.Printf("Signer Certificate Hash: %s\n", info["SignerCertificateHash"])
			fmt.Printf("Signature Valid: %s\n", info["SignatureValid"])
			fmt.Printf("Signer Name: %s\n", info["SignerName"])
			fmt.Printf("Signer Issuer: %s\n", info["SignerIssuer"])
			fmt.Printf("Signature Time: %s\n", info["SignatureTime"])
			fmt.Printf("In Blockchain: %s\n", info["InBlockchain"])
			if info["InBlockchain"] == "true" {
				fmt.Printf("Block Info: %s\n", info["BlockInfo"])
			}

			// Save to file if requested
			if outputFile != "" {
				// Convert info to JSON
				jsonData, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to convert to JSON: %w", err)
				}

				// Write to file
				if err := ioutil.WriteFile(outputFile, jsonData, 0644); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}

				fmt.Printf("\nDocument information saved to %s\n", outputFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&documentPath, "document", "", "path to the document")
	cmd.Flags().StringVar(&signaturePath, "signature", "", "path to the signature file")
	cmd.Flags().StringVar(&outputFile, "output", "", "path to save the document information as JSON")

	cmd.MarkFlagRequired("document")
	cmd.MarkFlagRequired("signature")

	return cmd
}

// getBlockchain creates a blockchain instance
func getBlockchain() *blockchain.Blockchain {
	// Get blockchain difficulty from config
	difficulty := viper.GetInt("difficulty")
	if difficulty == 0 {
		difficulty = 4 // Default difficulty
	}

	// Create a new blockchain
	return blockchain.NewBlockchain(difficulty)
}
