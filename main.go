package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env - log but don't crash if missing
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using environment variables or defaults")
	}

	http.HandleFunc("/api/fill", handleFill)
	http.HandleFunc("/script.js", serveScript)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Println("PORT not set, defaulting to 8000")
	}

	addr := "0.0.0.0:" + port
	log.Printf("Server starting on %s", addr)
	log.Printf("API endpoint: http://localhost:%s/api/fill", port)
	log.Printf("Script endpoint: http://localhost:%s/script.js", port)

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func serveScript(w http.ResponseWriter, r *http.Request) {
	log.Printf("hello")
	// Only allow GET requests
	if r.Method != "GET" {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-cache") // Don't cache during dev

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = fmt.Sprintf("http://localhost:%s/api/fill", os.Getenv("PORT"))
		if os.Getenv("PORT") == "" {
			apiURL = "http://localhost:8000/api/fill"
		}
		log.Printf("Warning: API_URL not set, using default: %s", apiURL)
	}

	script, err := os.ReadFile("script.js")
	if err != nil {
		log.Printf("Error reading script.js: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Script file not found")
		return
	}

	scriptWithURL := fmt.Sprintf("const API_URL = '%s';\n%s", apiURL, string(script))
	
	if _, err := w.Write([]byte(scriptWithURL)); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func handleFill(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit request body size (prevent abuse)
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB max

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Validate we got some data
	if len(body) == 0 {
		respondWithError(w, http.StatusBadRequest, "Empty request body")
		return
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		log.Printf("Received body: %s", string(body))
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Log the received data (pretty print for debugging)
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Error formatting JSON for logging: %v", err)
		log.Printf("Raw data: %+v", data)
	} else {
		log.Printf("Received JSON:\n%s", string(prettyJSON))
	}

	// TODO: Process the form data here
	// For now, return dummy response
	response := map[string]string{
		"name":    "Luke Henderson",
		"email":   "luke@example.com",
		"phone":   "555-123-4567",
		"message": "Auto-filled from Go API!",
	}

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		// Can't use respondWithError here since we already started writing
		return
	}

	log.Println("Successfully processed and responded to request")
}

// Helper function for consistent error responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	response := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}
