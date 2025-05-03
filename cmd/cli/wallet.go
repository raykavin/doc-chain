package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"document-blockchain/internal/models"
)

func newWalletCmd() *cobra.Command {
	walletCmd := &cobra.Command{
		Use:   "wallet",
		Short: "Manage wallets",
		Long:  `Create and manage wallets for signing documents.`,
	}

	walletCmd.AddCommand(newWalletCreateCmd())
	walletCmd.AddCommand(newWalletInfoCmd())
	walletCmd.AddCommand(newWalletExportCmd())
	walletCmd.AddCommand(newWalletImportCmd())

	return walletCmd
}

func newWalletCreateCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new wallet",
		Long:  `Create a new wallet with a key pair for signing documents.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiURL := viper.GetString("api-url")
			url := fmt.Sprintf("%s/wallet/new", apiURL)

			// Make the API request
			resp, err := http.Post(url, "application/json", nil)
			if err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}
			defer resp.Body.Close()

			// Read the response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			// Parse the response
			var response models.BlockchainResponse
			if err := json.Unmarshal(body, &response); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Check if the request was successful
			if !response.Success {
				return fmt.Errorf("API error: %s", response.Message)
			}

			// Extract the data
			data, ok := response.Data.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			// Print the wallet details
			fmt.Println("Wallet created successfully!")
			fmt.Printf("ID: %s\n", data["id"])
			fmt.Printf("Public Key: %s\n", data["publicKey"])
			
			// Save the wallet to a file if requested
			if outputFile != "" {
				walletData, err := json.MarshalIndent(data, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal wallet data: %w", err)
				}

				if err := ioutil.WriteFile(outputFile, walletData, 0600); err != nil {
					return fmt.Errorf("failed to write wallet to file: %w", err)
				}

				fmt.Printf("Wallet saved to %s\n", outputFile)
				fmt.Println("WARNING: This file contains your private key. Keep it secure!")
			} else {
				fmt.Println("Private Key: ********** (not shown for security)")
				fmt.Println("Use the --output flag to save the wallet (including private key) to a file.")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "save wallet to file")

	return cmd
}

func newWalletInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [wallet-id]",
		Short: "Get wallet information",
		Long:  `Get information about a wallet by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			walletID := args[0]
			apiURL := viper.GetString("api-url")
			url := fmt.Sprintf("%s/wallet/%s", apiURL, walletID)

			// Make the API request
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("failed to get wallet: %w", err)
			}
			defer resp.Body.Close()

			// Read the response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			// Parse the response
			var response models.BlockchainResponse
			if err := json.Unmarshal(body, &response); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Check if the request was successful
			if !response.Success {
				return fmt.Errorf("API error: %s", response.Message)
			}

			// Extract the data
			data, ok := response.Data.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			// Print the wallet details
			fmt.Printf("Wallet ID: %s\n", data["id"])
			fmt.Printf("Public Key: %s\n", data["publicKey"])

			return nil
		},
	}
}

func newWalletExportCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "export [wallet-id]",
		Short: "Export a wallet",
		Long:  `Export a wallet's public key to a file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			walletID := args[0]
			apiURL := viper.GetString("api-url")
			url := fmt.Sprintf("%s/wallet/%s", apiURL, walletID)

			// Make the API request
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("failed to get wallet: %w", err)
			}
			defer resp.Body.Close()

			// Read the response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			// Parse the response
			var response models.BlockchainResponse
			if err := json.Unmarshal(body, &response); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Check if the request was successful
			if !response.Success {
				return fmt.Errorf("API error: %s", response.Message)
			}

			// Extract the data
			data, ok := response.Data.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			// Export the wallet
			walletData := map[string]string{
				"id":        data["id"].(string),
				"publicKey": data["publicKey"].(string),
			}

			// Marshal the data
			exportData, err := json.MarshalIndent(walletData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal wallet data: %w", err)
			}

			// Save to a file or print to stdout
			if outputFile != "" {
				if err := ioutil.WriteFile(outputFile, exportData, 0644); err != nil {
					return fmt.Errorf("failed to write wallet to file: %w", err)
				}
				fmt.Printf("Wallet exported to %s\n", outputFile)
			} else {
				fmt.Println(string(exportData))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file (default: stdout)")

	return cmd
}

func newWalletImportCmd() *cobra.Command {
	var inputFile string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a wallet",
		Long:  `Import a wallet from a file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFile == "" {
				return fmt.Errorf("input file is required")
			}

			// Read the wallet file
			walletData, err := ioutil.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read wallet file: %w", err)
			}

			// Parse the wallet data
			var wallet map[string]interface{}
			if err := json.Unmarshal(walletData, &wallet); err != nil {
				return fmt.Errorf("failed to parse wallet data: %w", err)
			}

			// Check if the wallet contains an ID
			id, ok := wallet["id"].(string)
			if !ok {
				return fmt.Errorf("wallet data does not contain an ID")
			}

			fmt.Printf("Wallet imported: %s\n", id)
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file")
	cmd.MarkFlagRequired("input")

	return cmd
}