package query

import "fmt"

// Strategy defines how to execute a search
type Strategy string

const (
	// StrategyDefault is the standard semantic search
	StrategyDefault Strategy = "default"

	// StrategyAnalysis returns broader context for code analysis
	// - Higher result limit
	// - Includes related chunks
	// - Good for understanding architecture
	StrategyAnalysis Strategy = "analysis"

	// StrategyDebug focuses on specific symbols and errors
	// - Prioritizes code chunks
	// - Filters out documentation
	// - Good for finding bugs
	StrategyDebug Strategy = "debug"

	// StrategyExploration returns diverse results for codebase exploration
	// - Includes all types (code, doc, config)
	// - Lower similarity threshold
	// - Good for discovery
	StrategyExploration Strategy = "exploration"

	// StrategyDocumentation focuses on documentation and comments
	// - Prioritizes doc chunks
	// - Includes markdown and comments
	// - Good for understanding concepts
	StrategyDocumentation Strategy = "documentation"
)

// StrategyConfig defines configuration for a search strategy
type StrategyConfig struct {
	// Base limit for results
	Limit int

	// Minimum similarity threshold
	MinSimilarity float64

	// Preferred types (ordered by priority)
	PreferredTypes []string

	// Excluded types
	ExcludedTypes []string

	// Include related chunks (expand context)
	IncludeRelated bool

	// Diversity factor (0-1, higher = more diverse results)
	DiversityFactor float64

	// Description for documentation
	Description string
}

// GetStrategyConfig returns configuration for a strategy
func GetStrategyConfig(strategy Strategy) *StrategyConfig {
	switch strategy {
	case StrategyAnalysis:
		return &StrategyConfig{
			Limit:           10,
			MinSimilarity:   0.3,
			PreferredTypes:  []string{"code", "doc", "config"},
			IncludeRelated:  true,
			DiversityFactor: 0.5,
			Description:     "Broad context for architecture analysis and code understanding",
		}

	case StrategyDebug:
		return &StrategyConfig{
			Limit:          5,
			MinSimilarity:  0.4,
			PreferredTypes: []string{"code"},
			ExcludedTypes:  []string{"doc"},
			Description:    "Focused search for debugging and error investigation",
		}

	case StrategyExploration:
		return &StrategyConfig{
			Limit:           15,
			MinSimilarity:   0.2,
			PreferredTypes:  []string{"code", "doc", "config", "test"},
			DiversityFactor: 0.8,
			Description:     "Diverse results for codebase discovery and exploration",
		}

	case StrategyDocumentation:
		return &StrategyConfig{
			Limit:          8,
			MinSimilarity:  0.3,
			PreferredTypes: []string{"doc", "code"},
			Description:    "Documentation-focused search for learning and understanding",
		}

	default: // StrategyDefault
		return &StrategyConfig{
			Limit:          5,
			MinSimilarity:  0.0,
			PreferredTypes: []string{"code", "doc", "config"},
			Description:    "Standard semantic search with balanced results",
		}
	}
}

// ApplyStrategy applies a strategy to search filters
func ApplyStrategy(filters *SearchFilters, strategy Strategy) {
	config := GetStrategyConfig(strategy)

	// Override limit if not explicitly set
	if filters.Limit == 5 { // Default value
		filters.Limit = config.Limit
	}

	// Set minimum similarity
	if filters.MinSimilarity == 0.0 { // Default value
		filters.MinSimilarity = config.MinSimilarity
	}

	// Apply type preferences if no types specified
	if len(filters.Types) == 0 {
		if len(config.ExcludedTypes) > 0 {
			// Don't set types, will be handled by exclusion logic
		} else if len(config.PreferredTypes) > 0 {
			// Note: We don't enforce preferred types as hard filters
			// Instead, we'll use them for re-ranking (implemented separately)
		}
	}
}

// GetAvailableStrategies returns all available strategies with descriptions
func GetAvailableStrategies() map[Strategy]string {
	strategies := map[Strategy]string{
		StrategyDefault:       GetStrategyConfig(StrategyDefault).Description,
		StrategyAnalysis:      GetStrategyConfig(StrategyAnalysis).Description,
		StrategyDebug:         GetStrategyConfig(StrategyDebug).Description,
		StrategyExploration:   GetStrategyConfig(StrategyExploration).Description,
		StrategyDocumentation: GetStrategyConfig(StrategyDocumentation).Description,
	}
	return strategies
}

// ParseStrategy parses a strategy from string
func ParseStrategy(s string) (Strategy, error) {
	switch s {
	case "default", "":
		return StrategyDefault, nil
	case "analysis":
		return StrategyAnalysis, nil
	case "debug":
		return StrategyDebug, nil
	case "exploration":
		return StrategyExploration, nil
	case "documentation":
		return StrategyDocumentation, nil
	default:
		return "", fmt.Errorf("unknown strategy: %s", s)
	}
}
