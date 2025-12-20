package matcher

import (
	"autofill-api/internal/llm"
	"fmt"
	"strings"
)

// FieldContext provides context for LLM field filling
type FieldContext struct {
	JobTitle       string
	Company        string
	ExperienceMD   string
	ResumeSummary  string
	Skills         []string
	Constants      map[string]string
}

// Field represents a form field to be filled
type Field struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Type        string                   `json:"type"`
	Label       string                   `json:"label"`
	Placeholder string                   `json:"placeholder"`
	Required    bool                     `json:"required"`
	Value       string                   `json:"value"`
	Options     []map[string]interface{} `json:"options,omitempty"`
}

// FillFieldWithLLM uses LLM to generate an appropriate answer for a field
func FillFieldWithLLM(field Field, context FieldContext, llmClient *llm.Client) (string, error) {
	prompt := buildFieldPrompt(field, context)

	// Use GenerateText for simple text responses
	response, err := llmClient.GenerateText(prompt)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	// Clean up response
	response = strings.TrimSpace(response)

	// Handle select/radio fields - try to match with options
	if field.Type == "select" || field.Type == "radio" {
		if len(field.Options) > 0 {
			// Try to find best match from options
			matched := matchOptionFromLLM(response, field.Options)
			if matched != "" {
				return matched, nil
			}
		}
	}

	return response, nil
}

// buildFieldPrompt constructs a prompt for LLM field filling
func buildFieldPrompt(field Field, context FieldContext) string {
	var sb strings.Builder

	sb.WriteString("You are filling out a job application form.\n\n")

	// Field information
	sb.WriteString(fmt.Sprintf("Field Label: %s\n", field.Label))
	if field.Placeholder != "" {
		sb.WriteString(fmt.Sprintf("Field Placeholder: %s\n", field.Placeholder))
	}
	sb.WriteString(fmt.Sprintf("Field Type: %s\n", field.Type))
	if field.Required {
		sb.WriteString("This field is REQUIRED\n")
	}

	// Options for select/radio fields
	if len(field.Options) > 0 {
		sb.WriteString("\nAvailable Options:\n")
		for i, opt := range field.Options {
			if text, ok := opt["text"].(string); ok {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, text))
			} else if value, ok := opt["value"].(string); ok {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, value))
			}
		}
		sb.WriteString("\nYou must choose one of the options above. Respond with the exact text of your chosen option.\n")
	}

	sb.WriteString("\n")

	// Context about the job
	if context.JobTitle != "" || context.Company != "" {
		sb.WriteString("Job Application:\n")
		if context.JobTitle != "" {
			sb.WriteString(fmt.Sprintf("  Position: %s\n", context.JobTitle))
		}
		if context.Company != "" {
			sb.WriteString(fmt.Sprintf("  Company: %s\n", context.Company))
		}
		sb.WriteString("\n")
	}

	// Company information (from company-info.txt)
	if context.ExperienceMD != "" {
		companyInfo := context.ExperienceMD
		if len(companyInfo) > 800 {
			companyInfo = companyInfo[:800] + "..."
		}
		sb.WriteString(fmt.Sprintf("Company Information:\n%s\n\n", companyInfo))
	}

	// Resume/experience context (truncated to save tokens)
	if context.ResumeSummary != "" {
		summary := context.ResumeSummary
		if len(summary) > 500 {
			summary = summary[:500] + "..."
		}
		sb.WriteString(fmt.Sprintf("Your Background:\n%s\n\n", summary))
	}

	// Skills if relevant
	if len(context.Skills) > 0 {
		sb.WriteString(fmt.Sprintf("Your Skills: %s\n\n", strings.Join(context.Skills, ", ")))
	}

	// Add some basic info from constants
	if len(context.Constants) > 0 {
		sb.WriteString("Basic Information:\n")
		for key, value := range context.Constants {
			// Only include relevant constants (not all)
			if isRelevantConstant(key) {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
			}
		}
		sb.WriteString("\n")
	}

	// Instructions
	sb.WriteString("Instructions:\n")
	sb.WriteString("- Provide a concise, professional answer appropriate for this field\n")
	sb.WriteString("- Be honest and accurate based on the provided background\n")
	sb.WriteString("- Keep answers brief unless the field explicitly asks for detail\n")
	sb.WriteString("- For yes/no questions, respond with just 'yes' or 'no'\n")
	if len(field.Options) > 0 {
		sb.WriteString("- For multiple choice, respond with ONLY the exact option text from the list above\n")
	}
	sb.WriteString("\nYour Answer (text only, no explanation):")

	return sb.String()
}

// isRelevantConstant filters which constants to include in LLM context
func isRelevantConstant(key string) bool {
	irrelevantKeys := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"phone":      true,
		"city":       true,
		"state":      true,
		"zip":        true,
	}
	return !irrelevantKeys[key]
}

// matchOptionFromLLM tries to match LLM response with available options
func matchOptionFromLLM(llmResponse string, options []map[string]interface{}) string {
	normalized := strings.ToLower(strings.TrimSpace(llmResponse))

	for _, opt := range options {
		// Try matching against text
		if text, ok := opt["text"].(string); ok {
			if strings.ToLower(text) == normalized || strings.Contains(normalized, strings.ToLower(text)) {
				if value, ok := opt["value"].(string); ok {
					return value
				}
				return text
			}
		}

		// Try matching against value
		if value, ok := opt["value"].(string); ok {
			if strings.ToLower(value) == normalized || strings.Contains(normalized, strings.ToLower(value)) {
				return value
			}
		}
	}

	// If no match found, return the LLM response as-is
	return llmResponse
}

// BatchFillFieldsWithLLM fills multiple fields using LLM in batch
func BatchFillFieldsWithLLM(fields []Field, context FieldContext, llmClient *llm.Client) (map[string]string, error) {
	results := make(map[string]string)

	for _, field := range fields {
		value, err := FillFieldWithLLM(field, context, llmClient)
		if err != nil {
			// Log error but continue with other fields
			continue
		}
		results[field.ID] = value
	}

	return results, nil
}
