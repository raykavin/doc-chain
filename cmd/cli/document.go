package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newDocumentCmd() *cobra.Command {
	documentCmd := &cobra.Command{
		Use:   "document",
		Short: "Manage documents",
		Long:  `Sign, verify, and manage documents in the blockchain.`,
	}

	documentCmd.AddCommand(newDocumentSignCmd())
	documentCmd.AddCommand(newDocumentGetCmd())
	documentCmd.AddCommand(newDocumentVerifyCmd())

	return documentCmd
}

func newDocumentSignCmd() *cobra.Command {
	var (
		content     string
		contentFile string
		keyFile     string
	)

	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign a document",
		Long:  `Sign a document and add it to the blockchain.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if at least one content source is provided
			if content == "" && contentFile == "" {
				return fmt.Errorf("either --content or --content-file must be provided")
			}

			// Check if key file is provided
			if keyFile == "" {
				return fmt.Errorf("--key-file is required")
			}

			// Read the document content
			var documentContent string
			if contentFile != "" {
				fileContent, err := ioutil.ReadFile(contentFile)
				if err != nil {
					return fmt.Errorf("failed to read content file: %w", err)
				}
				documentContent = string(fileContent)
			} else {
				documentContent = content
			}

			// Read the private key
			keyData, err := ioutil.ReadFile(keyFile)
			if err != nil {
				return fmt.Errorf("failed to read key file: %w", err)
			}

			// Parse the key data
			var wallet map[string]interface{}
			if err := json.Unmarshal(keyData, &wallet); err != nil {
				return fmt.Errorf("failed to parse key data: %w", err)
			}

			// Check if the wallet contains a private key
			privateKey, ok := wallet["privateKey"].(string)
			if !ok {
				return fmt.Errorf("wallet data does not contain a private key")
			}

			// Prepare the request data
			reqData := map[string]string{
				"content": documentContent,
				"key":     privateKey,
			}

			reqBody, err := json.Marshal(reqData)
			if err != nil {
				return fmt.Errorf("failed to marshal request data: %w", err)
			}

			// Make the API request
			apiURL := viper.GetString("api-url")
			url := fmt.Sprintf("%s/document/sign", apiURL)

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				return fmt.Errorf("failed to sign document: %w", err)
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

			// Print the document details
			fmt.Println("Document signed successfully!")
			fmt.Printf("Document ID: %s\n", data["documentId"])
			fmt.Printf("Signed By: %s\n", data["signedBy"])

			// Convert timestamp to a readable format
			if timestampStr, ok := data["timestamp"].(string); ok {
				timestamp, err := time.Parse(time.RFC3339, timestampStr)
				if err == nil {
					fmt.Printf("Timestamp: %s\n", timestamp.Format(time.RFC3339))
				} else {
					fmt.Printf("Timestamp: %s\n", timestampStr)
				}
			} else {
				fmt.Printf("Timestamp: %v\n", data["timestamp"])
			}

			fmt.Println("\nNote: Document is now in pending state. Mine a new block to include it in the blockchain.")

			return nil
		},
	}

	cmd.Flags().StringVar(&content, "content", "", "document content")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "file containing document content")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "file containing the wallet private key")
	cmd.MarkFlagRequired("key-file")

	return cmd
}

func newDocumentGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [document-id]",
		Short: "Get document information",
		Long:  `Get information about a document by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			documentID := args[0]
			apiURL := viper.GetString("api-url")
			url := fmt.Sprintf("%s/document/%s", apiURL, documentID)

			// Make the API request
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("failed to get document: %w", err)
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

			// Print the document details
			fmt.Printf("Document ID: %s\n", data["id"])
			fmt.Printf("Content Hash: %s\n", data["content"])
			fmt.Printf("Signed By: %s\n", data["signedBy"])

			// Get timestamp as int or string and format it
			if timestamp, ok := data["timestamp"].(float64); ok {
				t := time.Unix(int64(timestamp), 0)
				fmt.Printf("Timestamp: %s\n", t.Format(time.RFC3339))
			} else if timestampStr, ok := data["timestamp"].(string); ok {
				fmt.Printf("Timestamp: %s\n", timestampStr)
			} else {
				fmt.Printf("Timestamp: %v\n", data["timestamp"])
			}

			// Display signature information if available
			if signature, ok := data["signature"].(map[string]interface{}); ok {
				fmt.Println("\nSignature:")
				fmt.Printf("  R: %v\n", signature["r"])
				fmt.Printf("  S: %v\n", signature["s"])
			}

			return nil
		},
	}
}

func newDocumentVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify [document-id]",
		Short: "Verify a document's signature",
		Long:  `Verify the signature of a document by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			documentID := args[0]
			apiURL := viper.GetString("api-url")
			url := fmt.Sprintf("%s/document/verify/%s", apiURL, documentID)

			// Make the API request
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("failed to verify document: %w", err)
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

			// Print the verification result
			isValid, _ := data["isValid"].(bool)

			if isValid {
				fmt.Println("✅ Document signature is VALID")
			} else {
				fmt.Println("❌ Document signature is INVALID")
			}

			fmt.Printf("Document ID: %s\n", data["documentId"])
			fmt.Printf("Signed By: %s\n", data["signedBy"])

			// Get timestamp as int or string and format it
			if timestamp, ok := data["timestamp"].(float64); ok {
				t := time.Unix(int64(timestamp), 0)
				fmt.Printf("Timestamp: %s\n", t.Format(time.RFC3339))
			} else if timestampStr, ok := data["timestamp"].(string); ok {
				fmt.Printf("Timestamp: %s\n", timestampStr)
			} else {
				fmt.Printf("Timestamp: %v\n", data["timestamp"])
			}

			return nil
		},
	}
}
