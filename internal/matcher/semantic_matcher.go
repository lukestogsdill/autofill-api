package matcher

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// SemanticMatcher handles embedding-based field matching
type SemanticMatcher struct {
	client      *genai.Client
	embedModel  *genai.EmbeddingModel
	ctx         context.Context
	vectorCache map[string][]float32
	mu          sync.RWMutex
}

// SemanticMatchResult contains the result of a semantic match
type SemanticMatchResult struct {
	Key        string
	Value      string
	Similarity float64
	Found      bool
}

var (
	globalSemanticMatcher *SemanticMatcher
	semanticMatcherOnce   sync.Once
)

// InitSemanticMatcher initializes the global semantic matcher with cached embeddings
func InitSemanticMatcher(ctx context.Context, apiKey string, constants map[string]string) error {
	var initErr error

	semanticMatcherOnce.Do(func() {
		matcher, err := NewSemanticMatcher(ctx, apiKey)
		if err != nil {
			initErr = fmt.Errorf("failed to create semantic matcher: %w", err)
			return
		}

		// Batch embed all constant keys
		if err := matcher.CacheConstantEmbeddings(constants); err != nil {
			matcher.Close()
			initErr = fmt.Errorf("failed to cache embeddings: %w", err)
			return
		}

		globalSemanticMatcher = matcher
		log.Printf("✅ Semantic matcher initialized with %d cached embeddings", len(matcher.vectorCache))
	})

	return initErr
}

// GetSemanticMatcher returns the global semantic matcher instance
func GetSemanticMatcher() *SemanticMatcher {
	return globalSemanticMatcher
}

// NewSemanticMatcher creates a new semantic matcher
func NewSemanticMatcher(ctx context.Context, apiKey string) (*SemanticMatcher, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Use gemini-embedding-004 for embeddings
	embedModel := client.EmbeddingModel("gemini-embedding-004")

	// Set task type for retrieval/similarity matching
	embedModel.TaskType = genai.TaskTypeRetrievalQuery

	return &SemanticMatcher{
		client:      client,
		embedModel:  embedModel,
		ctx:         ctx,
		vectorCache: make(map[string][]float32),
	}, nil
}

// CacheConstantEmbeddings batch-embeds all constant keys and caches them
func (sm *SemanticMatcher) CacheConstantEmbeddings(constants map[string]string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Extract unique keys to embed
	keys := make([]string, 0, len(constants))
	for key := range constants {
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil
	}

	log.Printf("🔄 Batch embedding %d constant keys...", len(keys))

	// Batch embed all keys at once
	batch := sm.embedModel.NewBatch()
	for _, key := range keys {
		// Convert underscore-separated keys to human-readable labels
		humanReadable := convertKeyToLabel(key)
		batch.AddContent(genai.Text(humanReadable))
	}

	embeddings, err := sm.embedModel.BatchEmbedContents(sm.ctx, batch)
	if err != nil {
		return fmt.Errorf("failed to batch embed constants: %w", err)
	}

	// Cache embeddings
	for i, key := range keys {
		if i < len(embeddings.Embeddings) {
			sm.vectorCache[key] = embeddings.Embeddings[i].Values
		}
	}

	log.Printf("✅ Cached %d constant embeddings", len(sm.vectorCache))
	return nil
}

// MatchFieldSemantically finds the best matching constant using semantic similarity
func (sm *SemanticMatcher) MatchFieldSemantically(label string, threshold float64) SemanticMatchResult {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.vectorCache) == 0 {
		return SemanticMatchResult{Found: false}
	}

	// Embed the incoming label
	embedding, err := sm.embedText(label)
	if err != nil {
		log.Printf("⚠️  Failed to embed label '%s': %v", label, err)
		return SemanticMatchResult{Found: false}
	}

	// Find best match using cosine similarity
	var bestKey string
	var bestSimilarity float64

	for key, cachedVector := range sm.vectorCache {
		similarity := cosineSimilarity(embedding, cachedVector)
		if similarity > bestSimilarity {
			bestSimilarity = similarity
			bestKey = key
		}
	}

	// Check if similarity meets threshold
	if bestSimilarity >= threshold {
		return SemanticMatchResult{
			Key:        bestKey,
			Similarity: bestSimilarity,
			Found:      true,
		}
	}

	return SemanticMatchResult{
		Key:        bestKey,
		Similarity: bestSimilarity,
		Found:      false,
	}
}

// embedText embeds a single text string
func (sm *SemanticMatcher) embedText(text string) ([]float32, error) {
	result, err := sm.embedModel.EmbedContent(sm.ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	if result == nil || result.Embedding == nil {
		return nil, fmt.Errorf("empty embedding result")
	}

	return result.Embedding.Values, nil
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (normA * normB)
}

// convertKeyToLabel converts underscore-separated keys to human-readable labels
// e.g., "first_name" -> "first name", "authorized_to_work" -> "authorized to work"
func convertKeyToLabel(key string) string {
	// Simple conversion: replace underscores with spaces
	result := ""
	for _, char := range key {
		if char == '_' {
			result += " "
		} else {
			result += string(char)
		}
	}
	return result
}

// AddToCache manually adds an embedding to the cache
func (sm *SemanticMatcher) AddToCache(key string, embedding []float32) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.vectorCache[key] = embedding
}

// GetCachedEmbedding retrieves a cached embedding
func (sm *SemanticMatcher) GetCachedEmbedding(key string) ([]float32, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	embedding, exists := sm.vectorCache[key]
	return embedding, exists
}

// ClearCache clears all cached embeddings
func (sm *SemanticMatcher) ClearCache() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.vectorCache = make(map[string][]float32)
}

// Close closes the semantic matcher client
func (sm *SemanticMatcher) Close() error {
	if sm.client != nil {
		return sm.client.Close()
	}
	return nil
}
