package dto


// BlockchainResponse generic response structure
type BlockchainResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// DocumentRequest request to add a document
type DocumentRequest struct {
	Content  string `json:"content"`
	SignedBy string `json:"signedBy"`
	Key      string `json:"key"` // PEM encoded private key
}