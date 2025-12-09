package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	// Load .env
	godotenv.Load()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Println("Available models:")
	iter := client.ListModels(ctx)
	for {
		model, err := iter.Next()
		if err != nil {
			break
		}
		fmt.Printf("  - %s\n", model.Name)
		fmt.Printf("    Supported methods: %v\n", model.SupportedGenerationMethods)
	}
}
