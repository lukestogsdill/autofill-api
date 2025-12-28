package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"autofill-api/internal/generator"
	"autofill-api/internal/input"
	"autofill-api/internal/llm"

	"github.com/joho/godotenv"
)

func main() {
	start := time.Now()
	log.Println("Starting resume generation...")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY not set in .env file")
	}

	// Create LLM client
	ctx := context.Background()
	llmClient, err := llm.NewClient(ctx, apiKey)
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}
	defer llmClient.Close()

	// Step 1: Load experience.md
	log.Println("Loading experience.md...")
	experienceMd, err := input.ParseExperience("experience.md")
	if err != nil {
		log.Fatalf("Failed to load experience: %v", err)
	}
	log.Printf("Loaded %d chars from experience.md", len(experienceMd))

	// Step 2: Parse job description
	log.Println("Loading job description...")
	jobDescText, err := os.ReadFile("job-description.txt")
	if err != nil {
		log.Fatalf("Failed to load job description: %v", err)
	}

	jobInfo, err := input.ParseJobDescription(string(jobDescText))
	if err != nil {
		log.Fatalf("Failed to parse job description: %v", err)
	}
	log.Printf("Targeting: %s at %s", jobInfo.Title, jobInfo.Company)

	// Step 3: Generate resume JSON with LLM
	log.Println("Generating resume with LLM...")

	// Save the prompt first
	jobInfoForLLM := llm.JobInfo{
		Company:     jobInfo.Company,
		Title:       jobInfo.Title,
		Description: jobInfo.Description,
		URL:         jobInfo.URL,
	}
	prompt := llm.ResumeGenerationPrompt(experienceMd, jobInfoForLLM)

	resume, err := generator.GenerateResumeJSON(experienceMd, jobInfoForLLM, llmClient)
	if err != nil {
		log.Fatalf("Failed to generate resume: %v", err)
	}

	// Step 4: Save outputs
	log.Println("Saving outputs...")
	if err := os.MkdirAll("generated", 0605); err != nil {
		log.Fatalf("Failed to create generated directory: %v", err)
	}

	// Save prompt
	promptPath := "generated/prompt.txt"
	if err := os.WriteFile(promptPath, []byte(prompt), 0644); err != nil {
		log.Fatalf("Failed to write prompt file: %v", err)
	}
	log.Printf("Prompt saved to: %s", promptPath)

	// Save JSON
	jsonOutput, err := json.MarshalIndent(resume, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal resume JSON: %v", err)
	}

	jsonPath := "generated/resume.json"
	if err := os.WriteFile(jsonPath, jsonOutput, 0644); err != nil {
		log.Fatalf("Failed to write JSON file: %v", err)
	}

	// Step 5: Generate PDF
	log.Println("Generating PDF...")
	pdfPath := "generated/resume.pdf"
	if err := generator.GeneratePDF(resume, pdfPath); err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	// Show usage stats
	usage := llm.GetUsage()
	log.Printf("\n=== Generation Complete ===")
	log.Printf("Prompt saved to: %s", promptPath)
	log.Printf("JSON saved to: %s", jsonPath)
	log.Printf("PDF saved to: %s", pdfPath)
	log.Printf("Duration: %v", time.Since(start))
	log.Printf("LLM Stats:")
	log.Printf("  - Requests: %d", usage.RequestCount)
	log.Printf("  - Input tokens: %d", usage.InputTokens)
	log.Printf("  - Output tokens: %d", usage.OutputTokens)
	log.Printf("  - Total tokens: %d", usage.InputTokens+usage.OutputTokens)

	fmt.Printf("\n✓ Resume generated successfully!\n")
}
