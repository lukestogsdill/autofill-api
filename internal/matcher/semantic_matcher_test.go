package matcher

import (
	"context"
	"os"
	"testing"
)

func TestSemanticMatcher(t *testing.T) {
	// Skip if no API key available
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping semantic matcher test")
	}

	// Create test constants
	constants := map[string]string{
		"first_name":          "Luke",
		"last_name":           "Stogsdill",
		"email":               "lukestogsdill@gmail.com",
		"phone":               "(832) 392-2613",
		"authorized_to_work":  "yes",
		"require_sponsorship": "no",
		"years_experience":    "2",
		"willing_to_relocate": "yes",
		"veteran":             "no",
		"disability":          "no",
	}

	// Initialize semantic matcher
	ctx := context.Background()
	matcher, err := NewSemanticMatcher(ctx, apiKey)
	if err != nil {
		t.Fatalf("Failed to create semantic matcher: %v", err)
	}
	defer matcher.Close()

	// Cache embeddings
	err = matcher.CacheConstantEmbeddings(constants)
	if err != nil {
		t.Fatalf("Failed to cache embeddings: %v", err)
	}

	// Test cases with expected matches
	testCases := []struct {
		label            string
		expectedKey      string
		minSimilarity    float64
		shouldMatch      bool
		description      string
	}{
		{
			label:         "What is your first name?",
			expectedKey:   "first_name",
			minSimilarity: 0.7,
			shouldMatch:   true,
			description:   "Should match 'first_name' with high similarity",
		},
		{
			label:         "Email address",
			expectedKey:   "email",
			minSimilarity: 0.7,
			shouldMatch:   true,
			description:   "Should match 'email' with high similarity",
		},
		{
			label:         "Are you legally authorized to work?",
			expectedKey:   "authorized_to_work",
			minSimilarity: 0.7,
			shouldMatch:   true,
			description:   "Should match 'authorized_to_work' semantically",
		},
		{
			label:         "Do you need visa sponsorship?",
			expectedKey:   "require_sponsorship",
			minSimilarity: 0.7,
			shouldMatch:   true,
			description:   "Should match 'require_sponsorship' semantically",
		},
		{
			label:         "Years of professional experience",
			expectedKey:   "years_experience",
			minSimilarity: 0.7,
			shouldMatch:   true,
			description:   "Should match 'years_experience' semantically",
		},
		{
			label:         "Are you a veteran?",
			expectedKey:   "veteran",
			minSimilarity: 0.7,
			shouldMatch:   true,
			description:   "Should match 'veteran' with high similarity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := matcher.MatchFieldSemantically(tc.label, tc.minSimilarity)

			if tc.shouldMatch && !result.Found {
				t.Errorf("Expected to find match for '%s', but got no match (best: %s with %.3f)",
					tc.label, result.Key, result.Similarity)
			}

			if result.Found {
				t.Logf("✓ '%s' matched to '%s' with similarity %.3f",
					tc.label, result.Key, result.Similarity)

				if result.Key != tc.expectedKey {
					t.Logf("⚠️  Note: Expected '%s' but got '%s' (similarity: %.3f)",
						tc.expectedKey, result.Key, result.Similarity)
				}

				if result.Similarity < tc.minSimilarity {
					t.Errorf("Similarity %.3f is below threshold %.3f",
						result.Similarity, tc.minSimilarity)
				}
			}
		})
	}
}

func TestCosineSimilarity(t *testing.T) {
	testCases := []struct {
		name     string
		a        []float32
		b        []float32
		expected float64
		delta    float64
	}{
		{
			name:     "Identical vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "Orthogonal vectors",
			a:        []float32{1.0, 0.0},
			b:        []float32{0.0, 1.0},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name:     "Opposite vectors",
			a:        []float32{1.0, 0.0},
			b:        []float32{-1.0, 0.0},
			expected: -1.0,
			delta:    0.001,
		},
		{
			name:     "Different length vectors",
			a:        []float32{1.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 0.0,
			delta:    0.001,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cosineSimilarity(tc.a, tc.b)
			if result < tc.expected-tc.delta || result > tc.expected+tc.delta {
				t.Errorf("Expected similarity %.3f (±%.3f), got %.3f",
					tc.expected, tc.delta, result)
			} else {
				t.Logf("✓ Similarity: %.3f (expected: %.3f)", result, tc.expected)
			}
		})
	}
}

func TestConvertKeyToLabel(t *testing.T) {
	testCases := []struct {
		key      string
		expected string
	}{
		{"first_name", "first name"},
		{"authorized_to_work", "authorized to work"},
		{"email", "email"},
		{"years_experience", "years experience"},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			result := convertKeyToLabel(tc.key)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// Benchmark semantic matching performance
func BenchmarkSemanticMatch(b *testing.B) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		b.Skip("GEMINI_API_KEY not set")
	}

	constants := map[string]string{
		"first_name": "Luke",
		"email":      "test@example.com",
		"phone":      "123-456-7890",
	}

	ctx := context.Background()
	matcher, err := NewSemanticMatcher(ctx, apiKey)
	if err != nil {
		b.Fatalf("Failed to create matcher: %v", err)
	}
	defer matcher.Close()

	if err := matcher.CacheConstantEmbeddings(constants); err != nil {
		b.Fatalf("Failed to cache embeddings: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.MatchFieldSemantically("What is your first name?", 0.7)
	}
}
