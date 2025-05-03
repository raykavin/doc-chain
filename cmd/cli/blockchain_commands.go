package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newBlockchainBlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "block [hash]",
		Short: "Get a block by hash",
		Long:  `Get detailed information about a block by its hash.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			blockHash := args[0]
			apiURL := viper.GetString("api-url")
			url := fmt.Sprintf("%s/blockchain/block/%s", apiURL, blockHash)

			// Make the API request
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("failed to get block: %w", err)
			}
			defer resp.Body.Close()

			// Read the response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			// Parse the response
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Check if the request was successful
			success, ok := response["success"].(bool)
			if !ok || !success {
				message, _ := response["message"].(string)
				return fmt.Errorf("API error: %s", message)
			}

			// Extract the data
			block, ok := response["data"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			// Print the block details
			fmt.Printf("Block Details:\n")
			fmt.Printf("Index: %v\n", block["index"])
			fmt.Printf("Hash: %s\n", block["hash"])
			fmt.Printf("Previous Hash: %s\n", block["prevHash"])
			
			// Format timestamp
			if timestamp, ok := block["timestamp"].(float64); ok {
				t := time.Unix(int64(timestamp), 0)
				fmt.Printf("Timestamp: %s\n", t.Format(time.RFC3339))
			} else {
				fmt.Printf("Timestamp: %v\n", block["timestamp"])
			}
			
			fmt.Printf("Merkle Root: %s\n", block["merkleRoot"])
			fmt.Printf("Nonce: %v\n", block["nonce"])
			fmt.Printf("Difficulty: %v\n", block["difficulty"])
			
			// Display documents
			if documents, ok := block["documents"].([]interface{}); ok {
				fmt.Printf("\nDocuments (%d):\n", len(documents))
				
				for i, docData := range documents {
					doc, ok := docData.(map[string]interface{})
					if !ok {
						continue
					}
					
					fmt.Printf("\nDocument #%d:\n", i+1)
					fmt.Printf("  ID: %s\n", doc["id"])
					fmt.Printf("  Content Hash: %s\n", doc["content"])
					fmt.Printf("  Signed By: %s\n", doc["signedBy"])
					
					// Format timestamp
					if timestamp, ok := doc["timestamp"].(float64); ok {
						t := time.Unix(int64(timestamp), 0)
						fmt.Printf("  Timestamp: %s\n", t.Format(time.RFC3339))
					} else {
						fmt.Printf("  Timestamp: %v\n", doc["timestamp"])
					}
				}
			}

			return nil
		},
	}
}

func newBlockchainPendingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pending",
		Short: "List pending documents",
		Long:  `List all pending documents that haven't been mined into a block yet.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiURL := viper.GetString("api-url")
			url := fmt.Sprintf("%s/blockchain/pending", apiURL)

			// Make the API request
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("failed to get pending documents: %w", err)
			}
			defer resp.Body.Close()

			// Read the response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			// Parse the response
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Check if the request was successful
			success, ok := response["success"].(bool)
			if !ok || !success {
				message, _ := response["message"].(string)
				return fmt.Errorf("API error: %s", message)
			}

			// Extract the data
			data, ok := response["data"].([]interface{})
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			// Print the pending documents
			fmt.Printf("Pending Documents: %d\n\n", len(data))
			
			if len(data) == 0 {
				fmt.Println("No pending documents.")
				fmt.Println("Use 'blockchain-cli document sign' to sign a document.")
				return nil
			}
			
			for i, docData := range data {
				doc, ok := docData.(map[string]interface{})
				if !ok {
					continue
				}
				
				fmt.Printf("Document #%d:\n", i+1)
				fmt.Printf("  ID: %s\n", doc["id"])
				fmt.Printf("  Content Hash: %s\n", doc["content"])
				fmt.Printf("  Signed By: %s\n", doc["signedBy"])
				
				// Format timestamp
				if timestamp, ok := doc["timestamp"].(float64); ok {
					t := time.Unix(int64(timestamp), 0)
					fmt.Printf("  Timestamp: %s\n", t.Format(time.RFC3339))
				} else {
					fmt.Printf("  Timestamp: %v\n", doc["timestamp"])
				}
				
				fmt.Println()
			}
			
			fmt.Println("Use 'blockchain-cli blockchain mine' to mine these documents into a new block.")

			return nil
		},
	}
}

func newBlockchainValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the blockchain",
		Long:  `Validate the integrity of the entire blockchain.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiURL := viper.GetString("api-url")
			url := fmt.Sprintf("%s/blockchain/validate", apiURL)

			// Make the API request
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("failed to validate blockchain: %w", err)
			}
			defer resp.Body.Close()

			// Read the response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			// Parse the response
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Check if the request was successful
			success, ok := response["success"].(bool)
			if !ok || !success {
				message, _ := response["message"].(string)
				return fmt.Errorf("API error: %s", message)
			}

			// Extract the data
			data, ok := response["data"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			// Print the validation result
			isValid, _ := data["isValid"].(bool)
			
			if isValid {
				fmt.Println("✅ Blockchain is valid!")
				fmt.Println("The blockchain integrity is intact. All blocks and signatures are valid.")
			} else {
				fmt.Println("❌ Blockchain is INVALID!")
				fmt.Println("The blockchain integrity has been compromised. Some blocks or signatures are invalid.")
			}

			return nil
		},
	}
}