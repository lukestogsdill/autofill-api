package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"autofill-api/internal/input"
	"autofill-api/internal/llm"

	"github.com/joho/godotenv"
)

func main() {
	start := time.Now()
	log.Println("Starting cover letter generation...")

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

	// Step 3: Extract key achievements from experience
	log.Println("Extracting key achievements...")
	keyAchievements := extractKeyAchievements(experienceMd)

	// Step 4: Generate cover letter with LLM
	log.Println("Generating cover letter with LLM...")
	llm.ResetUsage()

	jobInfoForLLM := llm.JobInfo{
		Company:     jobInfo.Company,
		Title:       jobInfo.Title,
		Description: jobInfo.Description,
		URL:         jobInfo.URL,
	}

	// Create a brief summary from experience
	resumeSummary := createSummaryFromExperience(experienceMd)

	prompt := llm.CoverLetterPrompt(resumeSummary, keyAchievements, jobInfoForLLM)

	coverLetter, err := llmClient.GenerateText(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to generate cover letter: %v", err)
	}

	// Step 5: Save outputs
	log.Println("Saving outputs...")
	if err := os.MkdirAll("generated", 0755); err != nil {
		log.Fatalf("Failed to create generated directory: %v", err)
	}

	// Save prompt
	promptPath := "generated/coverletter-prompt.txt"
	if err := os.WriteFile(promptPath, []byte(prompt), 0644); err != nil {
		log.Fatalf("Failed to write prompt file: %v", err)
	}
	log.Printf("Prompt saved to: %s", promptPath)

	// Save cover letter
	coverLetterPath := "generated/coverletter.txt"
	if err := os.WriteFile(coverLetterPath, []byte(coverLetter), 0644); err != nil {
		log.Fatalf("Failed to write cover letter file: %v", err)
	}

	// Show usage stats
	usage := llm.GetUsage()
	log.Printf("\n=== Generation Complete ===")
	log.Printf("Prompt saved to: %s", promptPath)
	log.Printf("Cover letter saved to: %s", coverLetterPath)
	log.Printf("Duration: %v", time.Since(start))
	log.Printf("LLM Stats:")
	log.Printf("  - Requests: %d", usage.RequestCount)
	log.Printf("  - Input tokens: %d", usage.InputTokens)
	log.Printf("  - Output tokens: %d", usage.OutputTokens)
	log.Printf("  - Total tokens: %d", usage.InputTokens+usage.OutputTokens)

	fmt.Printf("\n✓ Cover letter generated successfully!\n")
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("%s\n", coverLetter)
	fmt.Printf("%s\n", strings.Repeat("=", 80))
}

// extractKeyAchievements pulls out notable achievements from experience markdown
func extractKeyAchievements(experienceMd string) []string {
	achievements := []string{}
	lines := strings.Split(experienceMd, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for bullet points that are likely achievements
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			achievement := strings.TrimPrefix(strings.TrimPrefix(line, "-"), "*")
			achievement = strings.TrimSpace(achievement)
			if len(achievement) > 20 { // Skip short lines
				achievements = append(achievements, achievement)
			}
		}
	}

	// Limit to top 5-6 achievements
	if len(achievements) > 6 {
		achievements = achievements[:6]
	}

	return achievements
}

// createSummaryFromExperience creates a brief summary for the cover letter prompt
func createSummaryFromExperience(experienceMd string) string {
	// Take first 500 characters as a quick summary
	if len(experienceMd) > 500 {
		return experienceMd[:500] + "..."
	}
	return experienceMd
}
