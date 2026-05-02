package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func startAPIServer(port string) {
	// Serve static files from the public directory
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	// Register API endpoint for summarization request
	http.HandleFunc("/api/summarize", handleSummarize)

	// Start the HTTP server
	log.Printf("Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleSummarize(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Type-Allowed-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONs" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ensure only request post requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request to Config struct

	var config Config
	err := json.NewDecoder(r.Body).Decode(&config)

	if err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate that text provided
	if config.Text == "" {
		respondWithError(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Process the summary
	response, err := summarizeText(config, false)

	if err != nil {
		respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(AppResponse{Error: message})
}
