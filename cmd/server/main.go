package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"document-blockchain/internal/server"
)

func main() {
	// Parse command line flags
	var (
		port       string
		difficulty int
	)
	
	flag.StringVar(&port, "port", getEnv("PORT", "8080"), "Server port")
	flag.IntVar(&difficulty, "difficulty", getEnvAsInt("DIFFICULTY", 4), "Mining difficulty")
	flag.Parse()
	
	// Create and configure the server
	srv := server.NewServer(port, difficulty)
	srv.Configure()
	
	// Start the server
	log.Printf("Starting Document Blockchain server on port %s with difficulty %d", port, difficulty)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	
	return value
}