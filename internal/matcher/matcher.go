package matcher

import (
	"regexp"
	"strings"
)

// FieldMatchResult represents the result of a field matching operation
type FieldMatchResult struct {
	Value  string
	Source string // "constant" or "llm"
	Found  bool
}

// NormalizeLabel normalizes a field label for matching
func NormalizeLabel(label string) string {
	// Convert to lowercase
	normalized := strings.ToLower(label)

	// Remove special characters except underscores and spaces
	reg := regexp.MustCompile(`[^a-z0-9\s_]`)
	normalized = reg.ReplaceAllString(normalized, "")

	// Replace multiple spaces with single space
	normalized = strings.Join(strings.Fields(normalized), " ")

	// Trim whitespace
	return strings.TrimSpace(normalized)
}

// detectNegation checks if a label contains negation words
func detectNegation(label string) bool {
	negationPatterns := []string{
		"not", "don't", "do not", "doesn't", "does not",
		"aren't", "are not", "isn't", "is not", "won't",
		"will not", "can't", "cannot", "haven't", "have not",
	}

	normalized := strings.ToLower(label)
	for _, pattern := range negationPatterns {
		if strings.Contains(normalized, pattern) {
			return true
		}
	}
	return false
}

// invertBoolean inverts yes/no values
func invertBoolean(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	yesVariants := []string{"yes", "y", "true", "1"}
	noVariants := []string{"no", "n", "false", "0"}

	for _, v := range yesVariants {
		if normalized == v {
			return "no"
		}
	}
	for _, v := range noVariants {
		if normalized == v {
			return "yes"
		}
	}
	return value // Return as-is if not a boolean
}

// MatchField attempts to match a field label to a constant
func MatchField(label, fieldName, placeholder string, constants map[string]string) FieldMatchResult {
	hasNegation := detectNegation(label)

	// Try exact match on field name first
	if fieldName != "" {
		normalized := NormalizeLabel(fieldName)
		if value, exists := constants[normalized]; exists {
			if hasNegation {
				value = invertBoolean(value)
			}
			return FieldMatchResult{Value: value, Source: "constant", Found: true}
		}

		// Try field name with underscores replaced
		underscored := strings.ReplaceAll(normalized, " ", "_")
		if value, exists := constants[underscored]; exists {
			if hasNegation {
				value = invertBoolean(value)
			}
			return FieldMatchResult{Value: value, Source: "constant", Found: true}
		}
	}

	// Try label matching
	normalizedLabel := NormalizeLabel(label)
	if normalizedLabel != "" {
		// Direct match
		if value, exists := constants[normalizedLabel]; exists {
			if hasNegation {
				value = invertBoolean(value)
			}
			return FieldMatchResult{Value: value, Source: "constant", Found: true}
		}

		// Try with underscores
		underscored := strings.ReplaceAll(normalizedLabel, " ", "_")
		if value, exists := constants[underscored]; exists {
			if hasNegation {
				value = invertBoolean(value)
			}
			return FieldMatchResult{Value: value, Source: "constant", Found: true}
		}

		// Try semantic matching if available (threshold: 0.7)
		semanticMatcher := GetSemanticMatcher()
		if semanticMatcher != nil {
			semanticResult := semanticMatcher.MatchFieldSemantically(normalizedLabel, 0.7)
			if semanticResult.Found {
				if value, exists := constants[semanticResult.Key]; exists {
					if hasNegation {
						value = invertBoolean(value)
					}
					return FieldMatchResult{Value: value, Source: "constant", Found: true}
				}
			}
		}

		// Fallback to fuzzy matching against common patterns
		if match := fuzzyMatchLabel(normalizedLabel, constants); match.Found {
			if hasNegation {
				match.Value = invertBoolean(match.Value)
			}
			return match
		}
	}

	// Try placeholder if available
	if placeholder != "" {
		normalizedPlaceholder := NormalizeLabel(placeholder)
		if value, exists := constants[normalizedPlaceholder]; exists {
			if hasNegation {
				value = invertBoolean(value)
			}
			return FieldMatchResult{Value: value, Source: "constant", Found: true}
		}

		underscored := strings.ReplaceAll(normalizedPlaceholder, " ", "_")
		if value, exists := constants[underscored]; exists {
			if hasNegation {
				value = invertBoolean(value)
			}
			return FieldMatchResult{Value: value, Source: "constant", Found: true}
		}
	}

	return FieldMatchResult{Found: false}
}

// fuzzyMatchLabel performs fuzzy matching with common field patterns
func fuzzyMatchLabel(label string, constants map[string]string) FieldMatchResult {
	// Define common field label patterns and their corresponding constant keys
	patterns := map[string][]string{
		"first_name":           {"first name", "fname", "given name", "firstname"},
		"last_name":            {"last name", "lname", "surname", "family name", "lastname"},
		"name":                 {"full name", "your name", "complete name"},
		"email":                {"email address", "e mail", "email addr", "your email"},
		"phone":                {"phone number", "mobile", "telephone", "cell phone", "phone no", "contact number"},
		"city":                 {"city", "town"},
		"state":                {"state", "province", "region"},
		"zip":                  {"zip code", "postal code", "zipcode", "postcode"},
		"country":              {"country"},
		"linkedin_url":         {"linkedin", "linkedin profile", "linkedin url"},
		"github_url":           {"github", "github profile", "github url"},
		"website_url":          {"website", "personal website", "portfolio", "portfolio url"},
		"years_experience":     {"years of experience", "years experience", "experience years", "yoe"},
		"willing_to_relocate":  {"willing to relocate", "relocation", "relocate"},
		"require_sponsorship":  {"require sponsorship", "need sponsorship", "visa sponsorship", "work authorization"},
		"authorized_to_work":   {"authorized to work", "work authorization", "legally authorized"},
		"over_18":              {"over 18", "age 18", "18 years old"},
		"veteran":              {"veteran", "military", "armed forces", "service member"},
		"veteran_status":       {"veteran status"},
		"disability":           {"disability", "disabled", "impairment"},
		"disability_status":    {"disability status"},
		"convicted_felony":     {"convicted", "felony", "criminal", "crime"},
		"security_clearance":   {"security clearance", "clearance"},
	}

	for constantKey, labelPatterns := range patterns {
		for _, pattern := range labelPatterns {
			if strings.Contains(label, pattern) {
				if value, exists := constants[constantKey]; exists {
					return FieldMatchResult{Value: value, Source: "constant", Found: true}
				}
			}
		}
	}

	// Try partial matches for name-related fields
	if strings.Contains(label, "name") {
		// Check if it's asking for last name
		if strings.Contains(label, "last") || strings.Contains(label, "sur") || strings.Contains(label, "family") {
			if value, exists := constants["last_name"]; exists {
				return FieldMatchResult{Value: value, Source: "constant", Found: true}
			}
			// Fallback: split full name if available
			if fullName, exists := constants["name"]; exists {
				parts := strings.Fields(fullName)
				if len(parts) >= 2 {
					return FieldMatchResult{Value: parts[len(parts)-1], Source: "constant", Found: true}
				}
			}
		} else if strings.Contains(label, "first") || strings.Contains(label, "given") {
			// Asking for first name
			if value, exists := constants["first_name"]; exists {
				return FieldMatchResult{Value: value, Source: "constant", Found: true}
			}
			// Fallback: split full name if available
			if fullName, exists := constants["name"]; exists {
				parts := strings.Fields(fullName)
				if len(parts) >= 1 {
					return FieldMatchResult{Value: parts[0], Source: "constant", Found: true}
				}
			}
		} else {
			// Just "name" - use full name
			if value, exists := constants["name"]; exists {
				return FieldMatchResult{Value: value, Source: "constant", Found: true}
			}
			// Fallback: build from first + last
			if firstName, ok1 := constants["first_name"]; ok1 {
				if lastName, ok2 := constants["last_name"]; ok2 {
					return FieldMatchResult{Value: firstName + " " + lastName, Source: "constant", Found: true}
				}
			}
		}
	}

	// Try email variations
	if strings.Contains(label, "email") || strings.Contains(label, "mail") {
		if value, exists := constants["email"]; exists {
			return FieldMatchResult{Value: value, Source: "constant", Found: true}
		}
	}

	// Try phone variations
	if strings.Contains(label, "phone") || strings.Contains(label, "mobile") || strings.Contains(label, "cell") {
		if value, exists := constants["phone"]; exists {
			return FieldMatchResult{Value: value, Source: "constant", Found: true}
		}
	}

	return FieldMatchResult{Found: false}
}

// BuildFullName constructs a full name from constants
func BuildFullName(constants map[string]string) string {
	firstName, _ := constants["first_name"]
	lastName, _ := constants["last_name"]
	return strings.TrimSpace(firstName + " " + lastName)
}

// GetMatchStats returns statistics about field matching
type MatchStats struct {
	TotalFields     int
	ConstantMatches int
	LLMMatches      int
	UnmatchedFields int
}

// CalculateStats calculates matching statistics
func CalculateStats(results map[string]FieldMatchResult) MatchStats {
	stats := MatchStats{
		TotalFields: len(results),
	}

	for _, result := range results {
		if result.Found {
			if result.Source == "constant" {
				stats.ConstantMatches++
			} else if result.Source == "llm" {
				stats.LLMMatches++
			}
		} else {
			stats.UnmatchedFields++
		}
	}

	return stats
}
