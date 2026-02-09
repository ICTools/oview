package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSearchFilters(t *testing.T) {
	filters := NewSearchFilters()

	assert.NotNil(t, filters)
	assert.Equal(t, 0.0, filters.MinSimilarity)
	assert.Equal(t, 5, filters.Limit)
	assert.Empty(t, filters.Languages)
	assert.Empty(t, filters.Types)
	assert.Empty(t, filters.PathPattern)
	assert.Empty(t, filters.Components)
	assert.Empty(t, filters.SymbolPattern)
}

func TestSearchFilters_BuildWhereClause_NoFilters(t *testing.T) {
	filters := NewSearchFilters()
	baseConditions := []string{"project_id = $1"}

	whereClause, args := filters.BuildWhereClause("test-project", baseConditions)

	assert.Equal(t, "WHERE project_id = $1", whereClause)
	assert.Equal(t, []interface{}{"test-project"}, args)
}

func TestSearchFilters_BuildWhereClause_WithLanguages(t *testing.T) {
	filters := &SearchFilters{
		Languages: []string{"PHP", "JavaScript"},
		Limit:     5,
	}
	baseConditions := []string{"project_id = $1"}

	whereClause, args := filters.BuildWhereClause("test-project", baseConditions)

	assert.Contains(t, whereClause, "project_id = $1")
	assert.Contains(t, whereClause, "language IN")
	assert.Contains(t, whereClause, "$2")
	assert.Contains(t, whereClause, "$3")
	assert.Equal(t, 3, len(args))
	assert.Equal(t, "test-project", args[0])
	assert.Equal(t, "PHP", args[1])
	assert.Equal(t, "JavaScript", args[2])
}

func TestSearchFilters_BuildWhereClause_WithTypes(t *testing.T) {
	filters := &SearchFilters{
		Types: []string{"code", "doc"},
		Limit: 5,
	}
	baseConditions := []string{"project_id = $1"}

	whereClause, args := filters.BuildWhereClause("test-project", baseConditions)

	assert.Contains(t, whereClause, "type IN")
	assert.Contains(t, whereClause, "$2")
	assert.Contains(t, whereClause, "$3")
	assert.Equal(t, 3, len(args))
	assert.Equal(t, "code", args[1])
	assert.Equal(t, "doc", args[2])
}

func TestSearchFilters_BuildWhereClause_WithPathPattern(t *testing.T) {
	filters := &SearchFilters{
		PathPattern: "src/Controller/*",
		Limit:       5,
	}
	baseConditions := []string{"project_id = $1"}

	whereClause, args := filters.BuildWhereClause("test-project", baseConditions)

	assert.Contains(t, whereClause, "path LIKE $2")
	assert.Equal(t, 2, len(args))
	assert.Equal(t, "test-project", args[0])
	assert.Equal(t, "src/Controller/%", args[1]) // * converted to %
}

func TestSearchFilters_BuildWhereClause_WithComponents(t *testing.T) {
	filters := &SearchFilters{
		Components: []string{"Controller", "Service"},
		Limit:      5,
	}
	baseConditions := []string{"project_id = $1"}

	whereClause, args := filters.BuildWhereClause("test-project", baseConditions)

	assert.Contains(t, whereClause, "component IN")
	assert.Contains(t, whereClause, "$2")
	assert.Contains(t, whereClause, "$3")
	assert.Equal(t, 3, len(args))
	assert.Equal(t, "Controller", args[1])
	assert.Equal(t, "Service", args[2])
}

func TestSearchFilters_BuildWhereClause_WithSymbolPattern(t *testing.T) {
	filters := &SearchFilters{
		SymbolPattern: "User*",
		Limit:         5,
	}
	baseConditions := []string{"project_id = $1"}

	whereClause, args := filters.BuildWhereClause("test-project", baseConditions)

	assert.Contains(t, whereClause, "symbol LIKE $2")
	assert.Equal(t, 2, len(args))
	assert.Equal(t, "test-project", args[0])
	assert.Equal(t, "User%", args[1]) // * converted to %
}

func TestSearchFilters_BuildWhereClause_MultipleFilters(t *testing.T) {
	filters := &SearchFilters{
		Languages:     []string{"PHP"},
		Types:         []string{"code"},
		PathPattern:   "src/*",
		Components:    []string{"Controller"},
		SymbolPattern: "User*",
		Limit:         5,
	}
	baseConditions := []string{"project_id = $1"}

	whereClause, args := filters.BuildWhereClause("test-project", baseConditions)

	assert.Contains(t, whereClause, "project_id = $1")
	assert.Contains(t, whereClause, "language IN")
	assert.Contains(t, whereClause, "type IN")
	assert.Contains(t, whereClause, "path LIKE")
	assert.Contains(t, whereClause, "component IN")
	assert.Contains(t, whereClause, "symbol LIKE")
	assert.Equal(t, 6, len(args))
	assert.Equal(t, "test-project", args[0])
	assert.Equal(t, "PHP", args[1])
	assert.Equal(t, "code", args[2])
	assert.Equal(t, "src/%", args[3])
	assert.Equal(t, "Controller", args[4])
	assert.Equal(t, "User%", args[5])
}

func TestParseFiltersFromArgs_Empty(t *testing.T) {
	args := map[string]interface{}{}

	filters := ParseFiltersFromArgs(args)

	assert.NotNil(t, filters)
	assert.Equal(t, 5, filters.Limit) // Default
	assert.Empty(t, filters.Languages)
	assert.Empty(t, filters.Types)
}

func TestParseFiltersFromArgs_Limit(t *testing.T) {
	args := map[string]interface{}{
		"limit": float64(10),
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, 10, filters.Limit)
}

func TestParseFiltersFromArgs_LimitCapped(t *testing.T) {
	args := map[string]interface{}{
		"limit": float64(100),
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, 20, filters.Limit) // Max 20
}

func TestParseFiltersFromArgs_LanguagesSingle(t *testing.T) {
	args := map[string]interface{}{
		"language": "PHP",
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, []string{"PHP"}, filters.Languages)
}

func TestParseFiltersFromArgs_LanguagesMultiple(t *testing.T) {
	args := map[string]interface{}{
		"languages": []interface{}{"PHP", "JavaScript"},
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, 2, len(filters.Languages))
	assert.Contains(t, filters.Languages, "PHP")
	assert.Contains(t, filters.Languages, "JavaScript")
}

func TestParseFiltersFromArgs_TypesSingle(t *testing.T) {
	args := map[string]interface{}{
		"type": "code",
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, []string{"code"}, filters.Types)
}

func TestParseFiltersFromArgs_TypesMultiple(t *testing.T) {
	args := map[string]interface{}{
		"types": []interface{}{"code", "doc"},
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, 2, len(filters.Types))
	assert.Contains(t, filters.Types, "code")
	assert.Contains(t, filters.Types, "doc")
}

func TestParseFiltersFromArgs_PathPattern(t *testing.T) {
	args := map[string]interface{}{
		"path_pattern": "src/Controller/*",
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, "src/Controller/*", filters.PathPattern)
}

func TestParseFiltersFromArgs_ComponentsSingle(t *testing.T) {
	args := map[string]interface{}{
		"component": "Controller",
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, []string{"Controller"}, filters.Components)
}

func TestParseFiltersFromArgs_ComponentsMultiple(t *testing.T) {
	args := map[string]interface{}{
		"components": []interface{}{"Controller", "Service"},
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, 2, len(filters.Components))
	assert.Contains(t, filters.Components, "Controller")
	assert.Contains(t, filters.Components, "Service")
}

func TestParseFiltersFromArgs_SymbolPattern(t *testing.T) {
	args := map[string]interface{}{
		"symbol_pattern": "User*",
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, "User*", filters.SymbolPattern)
}

func TestParseFiltersFromArgs_MinSimilarity(t *testing.T) {
	args := map[string]interface{}{
		"min_similarity": float64(0.8),
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, 0.8, filters.MinSimilarity)
}

func TestParseFiltersFromArgs_All(t *testing.T) {
	args := map[string]interface{}{
		"limit":          float64(10),
		"languages":      []interface{}{"PHP", "JavaScript"},
		"types":          []interface{}{"code", "doc"},
		"path_pattern":   "src/*",
		"components":     []interface{}{"Controller", "Service"},
		"symbol_pattern": "User*",
		"min_similarity": float64(0.7),
	}

	filters := ParseFiltersFromArgs(args)

	assert.Equal(t, 10, filters.Limit)
	assert.Equal(t, 2, len(filters.Languages))
	assert.Equal(t, 2, len(filters.Types))
	assert.Equal(t, "src/*", filters.PathPattern)
	assert.Equal(t, 2, len(filters.Components))
	assert.Equal(t, "User*", filters.SymbolPattern)
	assert.Equal(t, 0.7, filters.MinSimilarity)
}

func TestSearchFilters_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filters *SearchFilters
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid filters",
			filters: &SearchFilters{
				MinSimilarity: 0.5,
				Limit:         10,
			},
			wantErr: false,
		},
		{
			name: "min similarity too low",
			filters: &SearchFilters{
				MinSimilarity: -0.1,
				Limit:         10,
			},
			wantErr: true,
			errMsg:  "min_similarity must be between 0 and 1",
		},
		{
			name: "min similarity too high",
			filters: &SearchFilters{
				MinSimilarity: 1.1,
				Limit:         10,
			},
			wantErr: true,
			errMsg:  "min_similarity must be between 0 and 1",
		},
		{
			name: "limit too low",
			filters: &SearchFilters{
				MinSimilarity: 0.5,
				Limit:         0,
			},
			wantErr: true,
			errMsg:  "limit must be between 1 and 20",
		},
		{
			name: "limit too high",
			filters: &SearchFilters{
				MinSimilarity: 0.5,
				Limit:         21,
			},
			wantErr: true,
			errMsg:  "limit must be between 1 and 20",
		},
		{
			name: "edge case: min similarity 0",
			filters: &SearchFilters{
				MinSimilarity: 0.0,
				Limit:         5,
			},
			wantErr: false,
		},
		{
			name: "edge case: min similarity 1",
			filters: &SearchFilters{
				MinSimilarity: 1.0,
				Limit:         5,
			},
			wantErr: false,
		},
		{
			name: "edge case: limit 1",
			filters: &SearchFilters{
				MinSimilarity: 0.5,
				Limit:         1,
			},
			wantErr: false,
		},
		{
			name: "edge case: limit 20",
			filters: &SearchFilters{
				MinSimilarity: 0.5,
				Limit:         20,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filters.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
