package query

import (
	"fmt"
	"strings"
)

// SearchFilters defines filters for semantic search
type SearchFilters struct {
	// Language filters (e.g., "PHP", "JavaScript")
	Languages []string

	// Type filters (code, doc, config, test)
	Types []string

	// Path pattern (supports wildcards, e.g., "src/Service/*")
	PathPattern string

	// Component filter (directory/module name)
	Components []string

	// Symbol filter (function/class name pattern)
	SymbolPattern string

	// Minimum similarity threshold (0.0-1.0)
	MinSimilarity float64

	// Maximum results
	Limit int
}

// NewSearchFilters creates filters with sensible defaults
func NewSearchFilters() *SearchFilters {
	return &SearchFilters{
		MinSimilarity: 0.0, // No threshold by default
		Limit:         5,
	}
}

// BuildWhereClause builds a SQL WHERE clause from filters
func (f *SearchFilters) BuildWhereClause(projectID string, baseConditions []string) (string, []interface{}) {
	conditions := append([]string{}, baseConditions...)
	args := []interface{}{projectID}

	// Language filter
	if len(f.Languages) > 0 {
		placeholders := make([]string, len(f.Languages))
		for i, lang := range f.Languages {
			args = append(args, lang)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		conditions = append(conditions, fmt.Sprintf("language IN (%s)", strings.Join(placeholders, ",")))
	}

	// Type filter
	if len(f.Types) > 0 {
		placeholders := make([]string, len(f.Types))
		for i, typ := range f.Types {
			args = append(args, typ)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ",")))
	}

	// Path pattern filter
	if f.PathPattern != "" {
		// Convert glob pattern to SQL LIKE pattern
		likePattern := strings.ReplaceAll(f.PathPattern, "*", "%")
		args = append(args, likePattern)
		conditions = append(conditions, fmt.Sprintf("path LIKE $%d", len(args)))
	}

	// Component filter
	if len(f.Components) > 0 {
		placeholders := make([]string, len(f.Components))
		for i, comp := range f.Components {
			args = append(args, comp)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		conditions = append(conditions, fmt.Sprintf("component IN (%s)", strings.Join(placeholders, ",")))
	}

	// Symbol pattern filter
	if f.SymbolPattern != "" {
		likePattern := strings.ReplaceAll(f.SymbolPattern, "*", "%")
		args = append(args, likePattern)
		conditions = append(conditions, fmt.Sprintf("symbol LIKE $%d", len(args)))
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")
	return whereClause, args
}

// ParseFiltersFromArgs parses filters from MCP tool arguments
func ParseFiltersFromArgs(args map[string]interface{}) *SearchFilters {
	filters := NewSearchFilters()

	// Parse limit
	if l, ok := args["limit"].(float64); ok {
		filters.Limit = int(l)
		if filters.Limit > 20 {
			filters.Limit = 20
		}
	}

	// Parse languages
	if langs, ok := args["languages"].([]interface{}); ok {
		for _, lang := range langs {
			if l, ok := lang.(string); ok {
				filters.Languages = append(filters.Languages, l)
			}
		}
	} else if lang, ok := args["language"].(string); ok {
		filters.Languages = []string{lang}
	}

	// Parse types
	if types, ok := args["types"].([]interface{}); ok {
		for _, typ := range types {
			if t, ok := typ.(string); ok {
				filters.Types = append(filters.Types, t)
			}
		}
	} else if typ, ok := args["type"].(string); ok {
		filters.Types = []string{typ}
	}

	// Parse path pattern
	if pattern, ok := args["path_pattern"].(string); ok {
		filters.PathPattern = pattern
	}

	// Parse components
	if comps, ok := args["components"].([]interface{}); ok {
		for _, comp := range comps {
			if c, ok := comp.(string); ok {
				filters.Components = append(filters.Components, c)
			}
		}
	} else if comp, ok := args["component"].(string); ok {
		filters.Components = []string{comp}
	}

	// Parse symbol pattern
	if pattern, ok := args["symbol_pattern"].(string); ok {
		filters.SymbolPattern = pattern
	}

	// Parse minimum similarity
	if sim, ok := args["min_similarity"].(float64); ok {
		filters.MinSimilarity = sim
	}

	return filters
}

// Validate validates the filters
func (f *SearchFilters) Validate() error {
	if f.MinSimilarity < 0 || f.MinSimilarity > 1 {
		return fmt.Errorf("min_similarity must be between 0 and 1")
	}

	if f.Limit <= 0 || f.Limit > 20 {
		return fmt.Errorf("limit must be between 1 and 20")
	}

	return nil
}
