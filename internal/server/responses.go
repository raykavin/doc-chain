package server

import (
	"encoding/json"
	"net/http"

	"github.com/raykavin/doc-chain/internal/dto"
)

// SendJSONResponse sends a JSON response with the given status code and data
func SendJSONResponse(w http.ResponseWriter, statusCode int, response *dto.BlockchainResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// SendErrorResponse sends an error response with the given status code and message
func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := &dto.BlockchainResponse{
		Success: false,
		Message: message,
	}

	SendJSONResponse(w, statusCode, response)
}

// SendSuccessResponse sends a success response with the given message and data
func SendSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := &dto.BlockchainResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	SendJSONResponse(w, statusCode, response)
}
