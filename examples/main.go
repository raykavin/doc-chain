package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Response structure to parse API responses
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// WalletResponse for wallet creation
type WalletResponse struct {
	ID         string `json:"id"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

func main() {
	baseURL := "http://localhost:8080/api"
	
	// 1. Create a new wallet
	wallet, err := createWallet(baseURL)
	if err != nil {
		fmt.Printf("Error creating wallet: %v\n", err)
		return
	}
	fmt.Printf("Created wallet with ID: %s\n", wallet.ID)
	
	// 2. Sign a document
	docID, err := signDocument(baseURL, "This is a document to be signed and stored on the blockchain", wallet.PrivateKey)
	if err != nil {
		fmt.Printf("Error signing document: %v\n", err)
		return
	}
	fmt.Printf("Document signed with ID: %s\n", docID)
	
	// 3. Get blockchain info
	err = getBlockchainInfo(baseURL)
	if err != nil {
		fmt.Printf("Error getting blockchain info: %v\n", err)
		return
	}
	
	// 4. Mine a block
	err = mineBlock(baseURL)
	if err != nil {
		fmt.Printf("Error mining block: %v\n", err)
		return
	}
	
	// 5. Verify the document
	err = verifyDocument(baseURL, docID)
	if err != nil {
		fmt.Printf("Error verifying document: %v\n", err)
		return
	}
}

// Create a new wallet
func createWallet(baseURL string) (*WalletResponse, error) {
	resp, err := http.Post(baseURL+"/wallet/new", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	
	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}
	
	// Parse the wallet data
	walletData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}
	
	var wallet WalletResponse
	if err := json.Unmarshal(walletData, &wallet); err != nil {
		return nil, err
	}
	
	return &wallet, nil
}

// Sign a document
func signDocument(baseURL, content, privateKey string) (string, error) {
	reqBody, err := json.Marshal(map[string]string{
		"content": content,
		"key":     privateKey,
	})
	if err != nil {
		return "", err
	}
	
	resp, err := http.Post(baseURL+"/document/sign", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	
	if !response.Success {
		return "", fmt.Errorf("API error: %s", response.Message)
	}
	
	// Extract document ID from response
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response data format")
	}
	
	docID, ok := data["documentId"].(string)
	if !ok {
		return "", fmt.Errorf("document ID not found in response")
	}
	
	return docID, nil
}

// Get blockchain info
func getBlockchainInfo(baseURL string) error {
	resp, err := http.Get(baseURL + "/blockchain")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	
	if !response.Success {
		return fmt.Errorf("API error: %s", response.Message)
	}
	
	fmt.Println("Blockchain Info:")
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response data format")
	}
	
	for k, v := range data {
		fmt.Printf("%s: %v\n", k, v)
	}
	
	return nil
}

// Mine a block
func mineBlock(baseURL string) error {
	resp, err := http.Post(baseURL+"/blockchain/mine", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	
	if !response.Success {
		return fmt.Errorf("API error: %s", response.Message)
	}
	
	fmt.Println("Block Mined Successfully:")
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response data format")
	}
	
	for k, v := range data {
		fmt.Printf("%s: %v\n", k, v)
	}
	
	return nil
}

// Verify a document
func verifyDocument(baseURL, docID string) error {
	resp, err := http.Get(fmt.Sprintf("%s/document/verify/%s", baseURL, docID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	
	if !response.Success {
		return fmt.Errorf("API error: %s", response.Message)
	}
	
	fmt.Println("Document Verification Result:")
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response data format")
	}
	
	isValid, ok := data["isValid"].(bool)
	if !ok {
		return fmt.Errorf("isValid not found in response")
	}
	
	if isValid {
		fmt.Printf("Document %s is valid and verified on the blockchain\n", docID)
	} else {
		fmt.Printf("Document %s verification failed\n", docID)
	}
	
	return nil
}