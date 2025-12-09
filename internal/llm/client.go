package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Client wraps the Gemini API client with retry logic and token tracking
type Client struct {
	client *genai.Client
	model  *genai.GenerativeModel
	ctx    context.Context
}

// TokenUsage tracks API usage for monitoring
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	RequestCount int
}

var usage TokenUsage

// NewClient creates a new Gemini API client
func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Use Gemini Flash Lite Latest (cheapest, fastest)
	model := client.GenerativeModel("gemini-flash-lite-latest")
	model.SetTemperature(0.7)
	model.SetTopP(0.95)
	model.SetTopK(40)

	return &Client{
		client: client,
		model:  model,
		ctx:    ctx,
	}, nil
}

// GenerateJSON sends a prompt and expects JSON response using structured outputs
func (c *Client) GenerateJSON(prompt string, responseSchema interface{}, jsonSchema map[string]any) error {
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := c.generateJSONWithRetry(prompt, jsonSchema, attempt)
		if err != nil {
			lastErr = err
			continue
		}

		// Parse JSON response
		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			lastErr = fmt.Errorf("empty response from Gemini")
			continue
		}

		// Extract text from response
		text := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

		// Track token usage
		if resp.UsageMetadata != nil {
			usage.InputTokens += int(resp.UsageMetadata.PromptTokenCount)
			usage.OutputTokens += int(resp.UsageMetadata.CandidatesTokenCount)
			usage.RequestCount++
		}

		// Unmarshal JSON
		if err := json.Unmarshal([]byte(text), responseSchema); err != nil {
			lastErr = fmt.Errorf("failed to parse JSON response: %w", err)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// GenerateText sends a prompt and returns text response
func (c *Client) GenerateText(prompt string) (string, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := c.generateWithRetry(prompt, attempt)
		if err != nil {
			lastErr = err
			continue
		}

		// Extract text from response
		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			lastErr = fmt.Errorf("empty response from Gemini")
			continue
		}

		text := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

		// Track token usage
		if resp.UsageMetadata != nil {
			usage.InputTokens += int(resp.UsageMetadata.PromptTokenCount)
			usage.OutputTokens += int(resp.UsageMetadata.CandidatesTokenCount)
			usage.RequestCount++
		}

		return text, nil
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// generateWithRetry handles exponential backoff for rate limits
func (c *Client) generateWithRetry(prompt string, attempt int) (*genai.GenerateContentResponse, error) {
	if attempt > 1 {
		// Exponential backoff: 1s, 2s, 4s
		backoff := time.Duration(1<<uint(attempt-1)) * time.Second
		log.Printf("Retrying after %v (attempt %d)...", backoff, attempt)
		time.Sleep(backoff)
	}

	resp, err := c.model.GenerateContent(c.ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	return resp, nil
}

// generateJSONWithRetry handles exponential backoff for structured output requests
func (c *Client) generateJSONWithRetry(prompt string, jsonSchema map[string]any, attempt int) (*genai.GenerateContentResponse, error) {
	if attempt > 1 {
		// Exponential backoff: 1s, 2s, 4s
		backoff := time.Duration(1<<uint(attempt-1)) * time.Second
		log.Printf("Retrying after %v (attempt %d)...", backoff, attempt)
		time.Sleep(backoff)
	}

	// Configure model for structured output
	schema := convertToSchema(jsonSchema)
	c.model.ResponseMIMEType = "application/json"
	c.model.ResponseSchema = schema

	resp, err := c.model.GenerateContent(c.ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	return resp, nil
}

// convertToSchema converts map[string]any to genai.Schema
func convertToSchema(m map[string]any) *genai.Schema {
	schema := &genai.Schema{}

	if typeStr, ok := m["type"].(string); ok {
		schema.Type = stringToType(typeStr)
	}

	if desc, ok := m["description"].(string); ok {
		schema.Description = desc
	}

	if props, ok := m["properties"].(map[string]any); ok {
		schema.Properties = make(map[string]*genai.Schema)
		for k, v := range props {
			if propMap, ok := v.(map[string]any); ok {
				schema.Properties[k] = convertToSchema(propMap)
			}
		}
	}

	if items, ok := m["items"].(map[string]any); ok {
		schema.Items = convertToSchema(items)
	}

	if required, ok := m["required"].([]string); ok {
		schema.Required = required
	}

	return schema
}

// stringToType converts string type to genai.Type constant
func stringToType(typeStr string) genai.Type {
	switch typeStr {
	case "object":
		return genai.TypeObject
	case "array":
		return genai.TypeArray
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	default:
		return genai.TypeUnspecified
	}
}

// GetUsage returns current token usage statistics
func GetUsage() TokenUsage {
	return usage
}

// ResetUsage resets token usage counters
func ResetUsage() {
	usage = TokenUsage{}
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.client.Close()
}
