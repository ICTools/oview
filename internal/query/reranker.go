package query

import (
	"sort"
	"strings"
)

// SearchResult represents a search result with metadata
type SearchResult struct {
	ID         int
	Path       string
	Type       string
	Language   string
	Symbol     string
	Component  string
	Content    string
	Similarity float64
	Score      float64 // Final score after re-ranking
}

// Reranker re-ranks search results based on strategy and relevance
type Reranker struct {
	strategy Strategy
	config   *StrategyConfig
}

// NewReranker creates a new reranker
func NewReranker(strategy Strategy) *Reranker {
	return &Reranker{
		strategy: strategy,
		config:   GetStrategyConfig(strategy),
	}
}

// Rerank re-ranks results based on strategy preferences
func (r *Reranker) Rerank(results []SearchResult) []SearchResult {
	// Calculate scores
	for i := range results {
		results[i].Score = r.calculateScore(&results[i])
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply diversity if needed
	if r.config.DiversityFactor > 0 {
		results = r.applyDiversity(results)
	}

	return results
}

// calculateScore calculates relevance score for a result
func (r *Reranker) calculateScore(result *SearchResult) float64 {
	score := result.Similarity

	// Type preference boost
	typeBoost := r.getTypeBoost(result.Type)
	score *= (1.0 + typeBoost)

	// Symbol match boost (if searching for specific symbols)
	if result.Symbol != "" {
		score *= 1.1
	}

	// Content length penalty (very short chunks might be less useful)
	contentLength := float64(len(result.Content))
	if contentLength < 100 {
		score *= 0.9
	} else if contentLength > 2000 {
		score *= 0.95
	}

	// Component diversity (encourage results from different components)
	// This is handled separately in applyDiversity

	return score
}

// getTypeBoost returns a boost factor based on type preference
func (r *Reranker) getTypeBoost(typ string) float64 {
	for i, prefType := range r.config.PreferredTypes {
		if typ == prefType {
			// Higher boost for types earlier in the preference list
			return float64(len(r.config.PreferredTypes)-i) * 0.1
		}
	}
	return 0.0
}

// applyDiversity ensures results are diverse across components
func (r *Reranker) applyDiversity(results []SearchResult) []SearchResult {
	if len(results) <= 1 {
		return results
	}

	diversityFactor := r.config.DiversityFactor
	if diversityFactor <= 0 {
		return results
	}

	// Track components seen
	componentCount := make(map[string]int)

	// Apply diversity penalty
	for i := range results {
		comp := results[i].Component
		if comp == "" {
			comp = getComponentFromPath(results[i].Path)
		}

		count := componentCount[comp]
		if count > 0 {
			// Reduce score for repeated components
			penalty := 1.0 - (float64(count) * diversityFactor * 0.1)
			if penalty < 0.5 {
				penalty = 0.5 // Don't penalize too much
			}
			results[i].Score *= penalty
		}
		componentCount[comp]++
	}

	// Re-sort after diversity adjustment
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// getComponentFromPath extracts component name from path
func getComponentFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return "root"
}

// FilterByMinSimilarity filters results by minimum similarity threshold
func FilterByMinSimilarity(results []SearchResult, minSimilarity float64) []SearchResult {
	if minSimilarity <= 0 {
		return results
	}

	filtered := make([]SearchResult, 0, len(results))
	for _, r := range results {
		if r.Similarity >= minSimilarity {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// ExpandContext expands results with related chunks
func ExpandContext(results []SearchResult, allChunks []SearchResult, maxExpansion int) []SearchResult {
	if maxExpansion <= 0 {
		return results
	}

	// Build a map of existing results
	existing := make(map[int]bool)
	for _, r := range results {
		existing[r.ID] = true
	}

	// For each result, find related chunks from same file
	expanded := make([]SearchResult, 0, len(results)+maxExpansion)
	expanded = append(expanded, results...)

	for _, result := range results {
		related := 0
		for _, chunk := range allChunks {
			if related >= maxExpansion {
				break
			}
			if existing[chunk.ID] {
				continue
			}
			// Same file but different chunk
			if chunk.Path == result.Path && chunk.ID != result.ID {
				chunk.Score = result.Score * 0.8 // Lower score for expanded context
				expanded = append(expanded, chunk)
				existing[chunk.ID] = true
				related++
			}
		}
	}

	return expanded
}
