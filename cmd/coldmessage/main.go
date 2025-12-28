package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"autofill-api/internal/constants"
	"autofill-api/internal/input"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Generating cold message...")

	// Load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Step 1: Load constants
	log.Println("Loading constants...")
	consts, err := constants.LoadConstants()
	if err != nil {
		log.Fatalf("Failed to load constants: %v", err)
	}

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

	// Step 3: Get values from constants
	name := getConstant(consts, "name", "Your Name")
	email := getConstant(consts, "email", "your@email.com")
	portfolio := getConstant(consts, "portfolio", "")
	currentRole := getConstant(consts, "current_role", "a software engineer")
	keyTech := getConstant(consts, "key_tech", "modern web technologies")
	keyAchievement := getConstant(consts, "key_achievement", "I've built scalable applications that serve thousands of users.")

	// Step 4: Build message with fmt.Sprintf
	log.Println("Generating message...")
	coldMessage := fmt.Sprintf(`Hi there,

I came across the %s position at %s and was immediately interested. I'm currently %s with experience in %s, and I think I'd be a great fit for this role.

%s

I'd love to chat more about how my background aligns with what you're building at %s. Would you have 15 minutes this week for a quick call?

Looking forward to hearing from you!

%s
%s%s`,
		jobInfo.Title,
		jobInfo.Company,
		currentRole,
		keyTech,
		keyAchievement,
		jobInfo.Company,
		name,
		email,
		formatOptional(portfolio),
	)

	// Step 5: Save output
	log.Println("Saving output...")
	if err := os.MkdirAll("generated", 0605); err != nil {
		log.Fatalf("Failed to create generated directory: %v", err)
	}

	coldMessagePath := "generated/coldmessage.txt"
	if err := os.WriteFile(coldMessagePath, []byte(coldMessage), 0644); err != nil {
		log.Fatalf("Failed to write cold message file: %v", err)
	}

	log.Printf("\n=== Generation Complete ===")
	log.Printf("Cold message saved to: %s", coldMessagePath)

	fmt.Printf("\n✓ Cold message generated successfully!\n")
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("%s\n", coldMessage)
	fmt.Printf("%s\n", strings.Repeat("=", 80))
}

// getConstant retrieves a constant value or returns a default
func getConstant(consts map[string]string, key, defaultValue string) string {
	if value, ok := consts[key]; ok && value != "" {
		return value
	}
	return defaultValue
}

// formatOptional adds a newline before the value if it's not empty
func formatOptional(value string) string {
	if value != "" {
		return "\n" + value
	}
	return ""
}
