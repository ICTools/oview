package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStrategy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Strategy
		wantErr  bool
	}{
		{"default", "default", StrategyDefault, false},
		{"empty string", "", StrategyDefault, false},
		{"analysis", "analysis", StrategyAnalysis, false},
		{"debug", "debug", StrategyDebug, false},
		{"exploration", "exploration", StrategyExploration, false},
		{"documentation", "documentation", StrategyDocumentation, false},
		{"unknown", "unknown", "", true},
		{"invalid", "invalid_strategy", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStrategy(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unknown strategy")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetStrategyConfig_Default(t *testing.T) {
	config := GetStrategyConfig(StrategyDefault)

	assert.NotNil(t, config)
	assert.Equal(t, 5, config.Limit)
	assert.Equal(t, 0.0, config.MinSimilarity)
	assert.Equal(t, []string{"code", "doc", "config"}, config.PreferredTypes)
	assert.Empty(t, config.ExcludedTypes)
	assert.False(t, config.IncludeRelated)
	assert.Equal(t, 0.0, config.DiversityFactor)
	assert.NotEmpty(t, config.Description)
}

func TestGetStrategyConfig_Analysis(t *testing.T) {
	config := GetStrategyConfig(StrategyAnalysis)

	assert.NotNil(t, config)
	assert.Equal(t, 10, config.Limit)
	assert.Equal(t, 0.3, config.MinSimilarity)
	assert.Equal(t, []string{"code", "doc", "config"}, config.PreferredTypes)
	assert.Empty(t, config.ExcludedTypes)
	assert.True(t, config.IncludeRelated)
	assert.Equal(t, 0.5, config.DiversityFactor)
	assert.Contains(t, config.Description, "architecture")
}

func TestGetStrategyConfig_Debug(t *testing.T) {
	config := GetStrategyConfig(StrategyDebug)

	assert.NotNil(t, config)
	assert.Equal(t, 5, config.Limit)
	assert.Equal(t, 0.4, config.MinSimilarity)
	assert.Equal(t, []string{"code"}, config.PreferredTypes)
	assert.Equal(t, []string{"doc"}, config.ExcludedTypes)
	assert.False(t, config.IncludeRelated)
	assert.Equal(t, 0.0, config.DiversityFactor)
	assert.Contains(t, config.Description, "debug")
}

func TestGetStrategyConfig_Exploration(t *testing.T) {
	config := GetStrategyConfig(StrategyExploration)

	assert.NotNil(t, config)
	assert.Equal(t, 15, config.Limit)
	assert.Equal(t, 0.2, config.MinSimilarity)
	assert.Equal(t, []string{"code", "doc", "config", "test"}, config.PreferredTypes)
	assert.Empty(t, config.ExcludedTypes)
	assert.False(t, config.IncludeRelated)
	assert.Equal(t, 0.8, config.DiversityFactor)
	assert.Contains(t, config.Description, "exploration")
}

func TestGetStrategyConfig_Documentation(t *testing.T) {
	config := GetStrategyConfig(StrategyDocumentation)

	assert.NotNil(t, config)
	assert.Equal(t, 8, config.Limit)
	assert.Equal(t, 0.3, config.MinSimilarity)
	assert.Equal(t, []string{"doc", "code"}, config.PreferredTypes)
	assert.Empty(t, config.ExcludedTypes)
	assert.False(t, config.IncludeRelated)
	assert.Equal(t, 0.0, config.DiversityFactor)
	assert.Contains(t, config.Description, "Documentation")
}

func TestApplyStrategy_Default(t *testing.T) {
	filters := NewSearchFilters()
	ApplyStrategy(filters, StrategyDefault)

	assert.Equal(t, 5, filters.Limit)
	assert.Equal(t, 0.0, filters.MinSimilarity)
}

func TestApplyStrategy_Analysis(t *testing.T) {
	filters := NewSearchFilters()
	ApplyStrategy(filters, StrategyAnalysis)

	assert.Equal(t, 10, filters.Limit)
	assert.Equal(t, 0.3, filters.MinSimilarity)
}

func TestApplyStrategy_Debug(t *testing.T) {
	filters := NewSearchFilters()
	ApplyStrategy(filters, StrategyDebug)

	assert.Equal(t, 5, filters.Limit)
	assert.Equal(t, 0.4, filters.MinSimilarity)
}

func TestApplyStrategy_Exploration(t *testing.T) {
	filters := NewSearchFilters()
	ApplyStrategy(filters, StrategyExploration)

	assert.Equal(t, 15, filters.Limit)
	assert.Equal(t, 0.2, filters.MinSimilarity)
}

func TestApplyStrategy_Documentation(t *testing.T) {
	filters := NewSearchFilters()
	ApplyStrategy(filters, StrategyDocumentation)

	assert.Equal(t, 8, filters.Limit)
	assert.Equal(t, 0.3, filters.MinSimilarity)
}

func TestApplyStrategy_PreserveExplicitLimit(t *testing.T) {
	filters := &SearchFilters{
		Limit:         10,
		MinSimilarity: 0.0,
	}

	ApplyStrategy(filters, StrategyAnalysis)

	// Should preserve explicit limit
	assert.Equal(t, 10, filters.Limit)
	// But apply strategy's min similarity
	assert.Equal(t, 0.3, filters.MinSimilarity)
}

func TestApplyStrategy_PreserveExplicitMinSimilarity(t *testing.T) {
	filters := &SearchFilters{
		Limit:         5,
		MinSimilarity: 0.8,
	}

	ApplyStrategy(filters, StrategyAnalysis)

	// Should apply strategy's limit
	assert.Equal(t, 10, filters.Limit)
	// But preserve explicit min similarity
	assert.Equal(t, 0.8, filters.MinSimilarity)
}

func TestApplyStrategy_PreserveTypes(t *testing.T) {
	filters := &SearchFilters{
		Limit: 5,
		Types: []string{"code"},
	}

	ApplyStrategy(filters, StrategyDocumentation)

	// Should preserve explicit types
	assert.Equal(t, []string{"code"}, filters.Types)
	// But apply strategy's limit
	assert.Equal(t, 8, filters.Limit)
}

func TestGetAvailableStrategies(t *testing.T) {
	strategies := GetAvailableStrategies()

	assert.NotNil(t, strategies)
	assert.Equal(t, 5, len(strategies))

	// Verify all strategies are present
	assert.Contains(t, strategies, StrategyDefault)
	assert.Contains(t, strategies, StrategyAnalysis)
	assert.Contains(t, strategies, StrategyDebug)
	assert.Contains(t, strategies, StrategyExploration)
	assert.Contains(t, strategies, StrategyDocumentation)

	// Verify descriptions are not empty
	for strategy, description := range strategies {
		assert.NotEmpty(t, description, "Strategy %s should have description", strategy)
	}
}

func TestStrategyConstants(t *testing.T) {
	// Verify strategy constants have expected values
	assert.Equal(t, Strategy("default"), StrategyDefault)
	assert.Equal(t, Strategy("analysis"), StrategyAnalysis)
	assert.Equal(t, Strategy("debug"), StrategyDebug)
	assert.Equal(t, Strategy("exploration"), StrategyExploration)
	assert.Equal(t, Strategy("documentation"), StrategyDocumentation)
}

func TestStrategyConfig_AllStrategiesHaveValidLimits(t *testing.T) {
	strategies := []Strategy{
		StrategyDefault,
		StrategyAnalysis,
		StrategyDebug,
		StrategyExploration,
		StrategyDocumentation,
	}

	for _, strategy := range strategies {
		config := GetStrategyConfig(strategy)
		assert.Greater(t, config.Limit, 0, "Strategy %s should have positive limit", strategy)
		assert.LessOrEqual(t, config.Limit, 20, "Strategy %s limit should not exceed 20", strategy)
	}
}

func TestStrategyConfig_AllStrategiesHaveValidSimilarity(t *testing.T) {
	strategies := []Strategy{
		StrategyDefault,
		StrategyAnalysis,
		StrategyDebug,
		StrategyExploration,
		StrategyDocumentation,
	}

	for _, strategy := range strategies {
		config := GetStrategyConfig(strategy)
		assert.GreaterOrEqual(t, config.MinSimilarity, 0.0, "Strategy %s should have non-negative similarity", strategy)
		assert.LessOrEqual(t, config.MinSimilarity, 1.0, "Strategy %s similarity should not exceed 1.0", strategy)
	}
}

func TestStrategyConfig_AllStrategiesHaveDescription(t *testing.T) {
	strategies := []Strategy{
		StrategyDefault,
		StrategyAnalysis,
		StrategyDebug,
		StrategyExploration,
		StrategyDocumentation,
	}

	for _, strategy := range strategies {
		config := GetStrategyConfig(strategy)
		assert.NotEmpty(t, config.Description, "Strategy %s should have description", strategy)
		assert.Greater(t, len(config.Description), 10, "Strategy %s description should be meaningful", strategy)
	}
}
