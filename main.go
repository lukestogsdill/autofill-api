package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"autofill-api/internal/constants"
	"autofill-api/internal/llm"
	"autofill-api/internal/matcher"

	"github.com/joho/godotenv"
)

// Global job context storage (in-memory for now)
var currentJobContext = struct {
	Title   string `json:"title"`
	Company string `json:"company"`
	URL     string `json:"url"`
}{
	Title:   "Software Engineer",
	Company: "Default Company",
	URL:     "",
}

// loadJobContextFromFile loads job context from job-description.txt
// Line 1: Company name
// Line 2: Job title
// Line 3+: Job description
func loadJobContextFromFile() {
	data, err := os.ReadFile("job-description.txt")
	if err != nil {
		log.Println("Warning: No job-description.txt found, using default job context")
		return
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) >= 1 {
		currentJobContext.Company = strings.TrimSpace(lines[0])
	}

	if len(lines) >= 2 {
		currentJobContext.Title = strings.TrimSpace(lines[1])
	}

	log.Printf("Loaded job context from job-description.txt: %s at %s", currentJobContext.Title, currentJobContext.Company)
}

func main() {
	// Load .env - log but don't crash if missing
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using environment variables or defaults")
	}

	// Load job context from company-info.txt
	loadJobContextFromFile()

	// Initialize semantic matcher with cached embeddings
	log.Println("🚀 Initializing semantic matcher...")
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey != "" {
		consts, err := constants.LoadConstants()
		if err != nil {
			log.Printf("⚠️  Warning: Failed to load constants for semantic matcher: %v", err)
		} else {
			ctx := context.Background()
			if err := matcher.InitSemanticMatcher(ctx, apiKey, consts); err != nil {
				log.Printf("⚠️  Warning: Failed to initialize semantic matcher: %v", err)
				log.Println("    Falling back to traditional fuzzy matching")
			}
		}
	} else {
		log.Println("⚠️  GEMINI_API_KEY not set, semantic matching disabled")
		log.Println("    Using traditional fuzzy matching only")
	}

	http.HandleFunc("/api/fill", handleFill)
	http.HandleFunc("/api/fill-constants", handleFillConstants)
	http.HandleFunc("/api/fill-llm", handleFillLLM)
	http.HandleFunc("/api/context", handleContext)
	http.HandleFunc("/api/constants", handleConstants)
	http.HandleFunc("/api/recent", handleRecent)
	http.HandleFunc("/script.js", serveScript)
	http.HandleFunc("/autofill.user.js", serveUserscript)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Println("PORT not set, defaulting to 8000")
	}

	ip := os.Getenv("IP")
	if ip == "" {
		ip = "localhost"
	}

	addr := "0.0.0.0:" + port
	log.Printf("Server starting on %s (HTTPS)", addr)
	log.Printf("API endpoint: https://%s:%s/api/fill", ip, port)
	log.Printf("Script endpoint: https://%s:%s/script.js", ip, port)

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func serveScript(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != "GET" {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-cache") // Don't cache during dev

	ip := os.Getenv("IP")
	if ip == "" {
		ip = "localhost"
		log.Printf("Warning: IP not set, using default: %s", ip)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Printf("Warning: PORT not set, using default: %s", port)
	}

	apiURL := fmt.Sprintf("https://%s:%s/api/fill", ip, port)

	script, err := os.ReadFile("public/script.js")
	if err != nil {
		log.Printf("Error reading public/script.js: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Script file not found")
		return
	}

	// Replace the API_URL placeholder in the script
	scriptContent := string(script)
	scriptContent = fmt.Sprintf("const API_URL = '%s';\n%s", apiURL, scriptContent)

	if _, err := w.Write([]byte(scriptContent)); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func serveUserscript(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != "GET" {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-cache")

	ip := os.Getenv("IP")
	if ip == "" {
		ip = "localhost"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	apiURL := fmt.Sprintf("https://%s:%s/api/fill", ip, port)

	userscript, err := os.ReadFile("public/autofill.user.js")
	if err != nil {
		log.Printf("Error reading public/autofill.user.js: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Userscript file not found")
		return
	}

	// Replace placeholders in the userscript
	scriptContent := string(userscript)
	scriptContent = strings.ReplaceAll(scriptContent, "YOUR_SERVER_IP:PORT", fmt.Sprintf("%s:%s", ip, port))
	scriptContent = strings.ReplaceAll(scriptContent, "const API_URL = 'https://YOUR_SERVER_IP:PORT/api/fill';", fmt.Sprintf("const API_URL = '%s';", apiURL))

	if _, err := w.Write([]byte(scriptContent)); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

type FillRequest struct {
	Fields        []matcher.Field `json:"fields"`
	JobContext    struct {
		Title   string `json:"title"`
		Company string `json:"company"`
		URL     string `json:"url"`
	} `json:"job_context"`
	ConstantsOnly bool `json:"constants_only"`
}

type FillResponse struct {
	Fields   map[string]string `json:"fields"`
	Metadata struct {
		ConstantMatches int `json:"constant_matches"`
		LLMMatches      int `json:"llm_matches"`
		TotalFields     int `json:"total_fields"`
	} `json:"metadata"`
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

	// Parse JSON request
	var req FillRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	log.Println("========================================")
	log.Printf("📋 Processing %d fields for job: %s at %s", len(req.Fields), req.JobContext.Title, req.JobContext.Company)
	log.Println("========================================")

	// Load constants
	consts, err := constants.LoadConstants()
	if err != nil {
		log.Printf("Error loading constants: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to load constants")
		return
	}

	// Load company info
	companyInfo := ""
	if companyData, err := os.ReadFile("company-info.txt"); err == nil {
		companyInfo = string(companyData)
		log.Printf("📄 Loaded company-info.txt (%d characters)", len(companyInfo))
	} else {
		log.Printf("⚠️  No company-info.txt found, continuing without it")
	}

	// Initialize response
	response := FillResponse{
		Fields: make(map[string]string),
	}

	// Track unmatched fields for LLM
	var unmatchedFields []matcher.Field

	// Step 1: Try to match fields with constants
	log.Println("\n🔍 Step 1: Matching fields with constants...")
	for _, field := range req.Fields {
		result := matcher.MatchField(field.Label, field.Name, field.Placeholder, consts)
		if result.Found {
			response.Fields[field.ID] = result.Value
			response.Metadata.ConstantMatches++
			log.Printf("  ✓ '%s' → %s", field.Label, result.Value)
		} else {
			unmatchedFields = append(unmatchedFields, field)
		}
	}
	log.Printf("📊 Constants matched: %d/%d fields", response.Metadata.ConstantMatches, len(req.Fields))

	// Step 2: Use LLM for unmatched fields (unless constants_only is true)
	if len(unmatchedFields) > 0 && !req.ConstantsOnly {
		log.Printf("\n🤖 Step 2: Using LLM to fill %d unmatched fields...", len(unmatchedFields))

		// Get Gemini API key
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			log.Println("⚠️  Warning: GEMINI_API_KEY not set, skipping LLM fallback")
		} else {
			// Create LLM client
			ctx := context.Background()
			llmClient, err := llm.NewClient(ctx, apiKey)
			if err != nil {
				log.Printf("❌ Error creating LLM client: %v", err)
			} else {
				defer llmClient.Close()

				// Reset usage tracking
				llm.ResetUsage()

				// Build context for LLM (including company info)
				fieldContext := matcher.FieldContext{
					JobTitle:      req.JobContext.Title,
					Company:       req.JobContext.Company,
					ExperienceMD:  companyInfo, // Using company info for now
					Constants:     consts,
				}

				// Fill each unmatched field with LLM
				for _, field := range unmatchedFields {
					value, err := matcher.FillFieldWithLLM(field, fieldContext, llmClient)
					if err != nil {
						log.Printf("  ❌ '%s': %v", field.Label, err)
						continue
					}
					response.Fields[field.ID] = value
					response.Metadata.LLMMatches++
					log.Printf("  ✓ '%s' → %s", field.Label, value)
				}

				// Log token usage
				usage := llm.GetUsage()
				log.Printf("\n📈 LLM Token Usage:")
				log.Printf("  Input tokens:  %d", usage.InputTokens)
				log.Printf("  Output tokens: %d", usage.OutputTokens)
				log.Printf("  Total tokens:  %d", usage.InputTokens+usage.OutputTokens)
				log.Printf("  Requests made: %d", usage.RequestCount)
			}
		}
	}

	response.Metadata.TotalFields = len(req.Fields)

	log.Println("\n========================================")
	log.Printf("✅ Fill Complete!")
	log.Printf("  Constant matches: %d", response.Metadata.ConstantMatches)
	log.Printf("  LLM matches:      %d", response.Metadata.LLMMatches)
	log.Printf("  Total fields:     %d", response.Metadata.TotalFields)
	log.Println("========================================\n")

	// Save response to file
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("responses/response_%s.json", timestamp)

	// Create responses directory if it doesn't exist
	if err := os.MkdirAll("responses", 0605); err != nil {
		log.Printf("⚠️  Failed to create responses directory: %v", err)
	} else {
		responseJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			log.Printf("⚠️  Failed to marshal response: %v", err)
		} else {
			if err := os.WriteFile(filename, responseJSON, 0644); err != nil {
				log.Printf("⚠️  Failed to save response to file: %v", err)
			} else {
				log.Printf("💾 Response saved to: %s\n", filename)
			}
		}
	}

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

// ConstantsFillRequest for /api/fill-constants endpoint
type ConstantsFillRequest struct {
	Fields []matcher.Field `json:"fields"`
}

// ConstantsFillResponse for /api/fill-constants endpoint
type ConstantsFillResponse struct {
	Fields   map[string]string `json:"fields"`
	Metadata struct {
		Matched     int `json:"matched"`
		TotalFields int `json:"total_fields"`
	} `json:"metadata"`
}

func handleFillConstants(w http.ResponseWriter, r *http.Request) {
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

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB max

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		respondWithError(w, http.StatusBadRequest, "Empty request body")
		return
	}

	// Parse JSON request
	var req ConstantsFillRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	log.Println("========================================")
	log.Printf("📝 Fill Constants: Processing %d fields", len(req.Fields))
	log.Println("========================================")

	// Load constants
	consts, err := constants.LoadConstants()
	if err != nil {
		log.Printf("Error loading constants: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to load constants")
		return
	}

	// Initialize response
	response := ConstantsFillResponse{
		Fields: make(map[string]string),
	}

	// Match fields with constants
	log.Println("\n🔍 Matching fields with constants...")
	for _, field := range req.Fields {
		result := matcher.MatchField(field.Label, field.Name, field.Placeholder, consts)
		if result.Found {
			response.Fields[field.ID] = result.Value
			response.Metadata.Matched++
			log.Printf("  ✓ '%s' → %s", field.Label, result.Value)
		}
	}

	response.Metadata.TotalFields = len(req.Fields)

	log.Println("\n========================================")
	log.Printf("✅ Constants Fill Complete!")
	log.Printf("  Matched: %d/%d fields", response.Metadata.Matched, response.Metadata.TotalFields)
	log.Println("========================================\n")

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

// LLMFillRequest for /api/fill-llm endpoint
type LLMFillRequest struct {
	Fields     []matcher.Field `json:"fields"`
	JobContext struct {
		Title   string `json:"title"`
		Company string `json:"company"`
		URL     string `json:"url"`
	} `json:"job_context"`
}

// LLMFillResponse for /api/fill-llm endpoint
type LLMFillResponse struct {
	Fields   map[string]string `json:"fields"`
	Metadata struct {
		LLMMatches  int `json:"llm_matches"`
		TotalFields int `json:"total_fields"`
	} `json:"metadata"`
}

func handleFillLLM(w http.ResponseWriter, r *http.Request) {
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

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB max

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		respondWithError(w, http.StatusBadRequest, "Empty request body")
		return
	}

	// Parse JSON request
	var req LLMFillRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	log.Println("========================================")
	log.Printf("🤖 Fill LLM: Processing %d fields for %s at %s", len(req.Fields), req.JobContext.Title, req.JobContext.Company)
	log.Println("========================================")

	// Get Gemini API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("❌ Error: GEMINI_API_KEY not set")
		respondWithError(w, http.StatusInternalServerError, "LLM API key not configured")
		return
	}

	// Load constants (for context)
	consts, err := constants.LoadConstants()
	if err != nil {
		log.Printf("Error loading constants: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to load constants")
		return
	}

	// Load company info
	companyInfo := ""
	if companyData, err := os.ReadFile("company-info.txt"); err == nil {
		companyInfo = string(companyData)
		log.Printf("📄 Loaded company-info.txt (%d characters)", len(companyInfo))
	} else {
		log.Printf("⚠️  No company-info.txt found")
	}

	// Load experience
	experience := ""
	if expData, err := os.ReadFile("experience.txt"); err == nil {
		experience = string(expData)
		log.Printf("📄 Loaded experience.txt (%d characters)", len(experience))
	} else {
		log.Printf("⚠️  No experience.txt found")
	}

	// Combine company info and experience for context
	contextMD := companyInfo
	if experience != "" {
		if contextMD != "" {
			contextMD += "\n\n"
		}
		contextMD += experience
	}

	// Initialize response
	response := LLMFillResponse{
		Fields: make(map[string]string),
	}

	// Create LLM client
	ctx := context.Background()
	llmClient, err := llm.NewClient(ctx, apiKey)
	if err != nil {
		log.Printf("❌ Error creating LLM client: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create LLM client")
		return
	}
	defer llmClient.Close()

	// Reset usage tracking
	llm.ResetUsage()

	// Build context for LLM
	fieldContext := matcher.FieldContext{
		JobTitle:     req.JobContext.Title,
		Company:      req.JobContext.Company,
		ExperienceMD: contextMD,
		Constants:    consts,
	}

	// Fill each field with LLM
	log.Println("\n🤖 Generating LLM responses...")
	for _, field := range req.Fields {
		value, err := matcher.FillFieldWithLLM(field, fieldContext, llmClient)
		if err != nil {
			log.Printf("  ❌ '%s': %v", field.Label, err)
			continue
		}
		response.Fields[field.ID] = value
		response.Metadata.LLMMatches++
		log.Printf("  ✓ '%s' → %s", field.Label, value)
	}

	response.Metadata.TotalFields = len(req.Fields)

	// Log token usage
	usage := llm.GetUsage()
	log.Printf("\n📈 LLM Token Usage:")
	log.Printf("  Input tokens:  %d", usage.InputTokens)
	log.Printf("  Output tokens: %d", usage.OutputTokens)
	log.Printf("  Total tokens:  %d", usage.InputTokens+usage.OutputTokens)
	log.Printf("  Requests made: %d", usage.RequestCount)

	log.Println("\n========================================")
	log.Printf("✅ LLM Fill Complete!")
	log.Printf("  Filled: %d/%d fields", response.Metadata.LLMMatches, response.Metadata.TotalFields)
	log.Println("========================================\n")

	// Save response to file
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("responses/response_llm_%s.json", timestamp)

	if err := os.MkdirAll("responses", 0605); err != nil {
		log.Printf("⚠️  Failed to create responses directory: %v", err)
	} else {
		responseJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			log.Printf("⚠️  Failed to marshal response: %v", err)
		} else {
			if err := os.WriteFile(filename, responseJSON, 0644); err != nil {
				log.Printf("⚠️  Failed to save response to file: %v", err)
			} else {
				log.Printf("💾 Response saved to: %s\n", filename)
			}
		}
	}

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

func handleContext(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET
	if r.Method != "GET" {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Return current job context
	if err := json.NewEncoder(w).Encode(currentJobContext); err != nil {
		log.Printf("Error encoding context response: %v", err)
		return
	}

	log.Printf("Returned job context: %s at %s", currentJobContext.Title, currentJobContext.Company)
}

func handleConstants(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		// Return current constants
		consts, err := constants.LoadConstants()
		if err != nil {
			log.Printf("Error loading constants: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to load constants")
			return
		}

		if err := json.NewEncoder(w).Encode(consts); err != nil {
			log.Printf("Error encoding constants response: %v", err)
			return
		}

		log.Println("Returned constants to client")

	case "POST":
		// Update constants
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			respondWithError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}
		defer r.Body.Close()

		var updatedConstants map[string]string
		if err := json.Unmarshal(body, &updatedConstants); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
			return
		}

		// Save to constants.json
		constantsJSON, err := json.MarshalIndent(updatedConstants, "", "  ")
		if err != nil {
			log.Printf("Error marshaling constants: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to format constants")
			return
		}

		if err := os.WriteFile("constants.json", constantsJSON, 0644); err != nil {
			log.Printf("Error writing constants.json: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to save constants")
			return
		}

		// Reload constants in memory
		if err := constants.ReloadConstants(); err != nil {
			log.Printf("Error reloading constants: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to reload constants")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

		log.Printf("Updated constants (%d fields)", len(updatedConstants))

	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleRecent returns the most recent response from /responses directory
func handleRecent(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Read responses directory
	entries, err := os.ReadDir("responses")
	if err != nil {
		log.Printf("Error reading responses directory: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to read responses directory")
		return
	}

	// Filter for JSON files and find the most recent
	var mostRecentFile string
	var mostRecentTime time.Time

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if mostRecentFile == "" || info.ModTime().After(mostRecentTime) {
			mostRecentFile = entry.Name()
			mostRecentTime = info.ModTime()
		}
	}

	if mostRecentFile == "" {
		respondWithError(w, http.StatusNotFound, "No response files found")
		return
	}

	// Read the most recent file
	filePath := fmt.Sprintf("responses/%s", mostRecentFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v", filePath, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to read response file")
		return
	}

	// Parse and return the JSON
	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		log.Printf("Error parsing JSON from %s: %v", filePath, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to parse response file")
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}

	log.Printf("Returned most recent response: %s", mostRecentFile)
}

// Helper function for consistent error responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	response := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}
