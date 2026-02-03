package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/embeddings"
)

var (
	benchmarkOutputFile string
	benchmarkQueries    int
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run performance benchmarks on the RAG system",
	Long: `Run comprehensive benchmarks to test:
- Embedding generation speed
- Search query performance
- Result relevance
- Database performance
- End-to-end latency

Results are saved to a JSON file for analysis.`,
	RunE: runBenchmark,
}

func init() {
	benchmarkCmd.Flags().StringVarP(&benchmarkOutputFile, "output", "o", "benchmark_results.json", "Output file for results")
	benchmarkCmd.Flags().IntVarP(&benchmarkQueries, "queries", "n", 10, "Number of test queries to run")
	rootCmd.AddCommand(benchmarkCmd)
}

type BenchmarkResults struct {
	Timestamp          time.Time              `json:"timestamp"`
	ProjectID          string                 `json:"project_id"`
	ProjectSlug        string                 `json:"project_slug"`
	TotalChunks        int                    `json:"total_chunks"`
	EmbeddingProvider  string                 `json:"embedding_provider"`
	EmbeddingModel     string                 `json:"embedding_model"`
	EmbeddingDimension int                    `json:"embedding_dimension"`
	Tests              []BenchmarkTest        `json:"tests"`
	Summary            BenchmarkSummary       `json:"summary"`
	SystemInfo         map[string]interface{} `json:"system_info"`
}

type BenchmarkTest struct {
	Name             string        `json:"name"`
	Query            string        `json:"query,omitempty"`
	Duration         time.Duration `json:"duration_ms"`
	Success          bool          `json:"success"`
	Error            string        `json:"error,omitempty"`
	ResultCount      int           `json:"result_count,omitempty"`
	TopSimilarity    float64       `json:"top_similarity,omitempty"`
	AverageSimilarity float64      `json:"average_similarity,omitempty"`
}

type BenchmarkSummary struct {
	TotalTests              int           `json:"total_tests"`
	SuccessfulTests         int           `json:"successful_tests"`
	FailedTests             int           `json:"failed_tests"`
	AverageEmbeddingTime    time.Duration `json:"avg_embedding_time_ms"`
	AverageSearchTime       time.Duration `json:"avg_search_time_ms"`
	AverageEndToEndTime     time.Duration `json:"avg_end_to_end_time_ms"`
	MinSearchTime           time.Duration `json:"min_search_time_ms"`
	MaxSearchTime           time.Duration `json:"max_search_time_ms"`
	AverageResultRelevance  float64       `json:"avg_result_relevance"`
	ThroughputQueriesPerSec float64       `json:"throughput_queries_per_sec"`
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("🏁 Starting oview RAG Benchmark")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	// Load configs
	projectConfig, err := config.LoadProjectConfig(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Initialize results
	results := &BenchmarkResults{
		Timestamp:          time.Now(),
		ProjectID:          projectConfig.ProjectID,
		ProjectSlug:        projectConfig.ProjectSlug,
		EmbeddingProvider:  projectConfig.Embeddings.Provider,
		EmbeddingModel:     projectConfig.Embeddings.Model,
		EmbeddingDimension: projectConfig.Embeddings.Dim,
		Tests:              []BenchmarkTest{},
		SystemInfo:         make(map[string]interface{}),
	}

	// Get system info
	results.SystemInfo["go_version"] = "1.23+"
	results.SystemInfo["embedding_base_url"] = projectConfig.Embeddings.BaseURL

	// Connect to database
	dbName := fmt.Sprintf("oview_%s", projectConfig.ProjectSlug)
	dbUser := fmt.Sprintf("oview_%s", projectConfig.ProjectSlug)
	dbPassword := projectConfig.Database.Password

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, globalConfig.PostgresHost, globalConfig.PostgresPort, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Get total chunks
	err = db.QueryRow("SELECT COUNT(*) FROM chunks WHERE project_id = $1", projectConfig.ProjectID).Scan(&results.TotalChunks)
	if err != nil {
		return fmt.Errorf("failed to count chunks: %w", err)
	}

	fmt.Printf("📊 Project: %s\n", projectConfig.ProjectSlug)
	fmt.Printf("📦 Total chunks: %d\n", results.TotalChunks)
	fmt.Printf("🤖 Embeddings: %s / %s (%d dim)\n",
		projectConfig.Embeddings.Provider,
		projectConfig.Embeddings.Model,
		projectConfig.Embeddings.Dim)
	fmt.Println()

	// Initialize embeddings generator
	var generator embeddings.Generator
	switch projectConfig.Embeddings.Provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			apiKey = projectConfig.Embeddings.APIKey
		}
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY not set")
		}
		generator = embeddings.NewOpenAIGenerator(apiKey, projectConfig.Embeddings.Model)

	case "ollama":
		baseURL := projectConfig.Embeddings.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		generator = embeddings.NewOllamaGenerator(baseURL, projectConfig.Embeddings.Model)

	default:
		return fmt.Errorf("unsupported embeddings provider: %s", projectConfig.Embeddings.Provider)
	}

	// Test queries (diverse to test different scenarios)
	testQueries := []string{
		"authentication implementation",
		"database connection",
		"error handling",
		"configuration files",
		"user management",
		"API endpoints",
		"cache system",
		"security measures",
		"data validation",
		"logging mechanism",
	}

	// Limit to requested number of queries
	if benchmarkQueries < len(testQueries) {
		testQueries = testQueries[:benchmarkQueries]
	}

	fmt.Println("🧪 Running Benchmark Tests...")
	fmt.Println()

	// Test 1: Database connection latency
	fmt.Println("1️⃣  Testing database connection...")
	results.Tests = append(results.Tests, benchmarkDatabaseConnection(db))

	// Test 2: Embedding generation speed
	fmt.Println("2️⃣  Testing embedding generation...")
	embeddingTests := benchmarkEmbeddingGeneration(generator, testQueries)
	results.Tests = append(results.Tests, embeddingTests...)

	// Test 3: Search performance
	fmt.Println("3️⃣  Testing search performance...")
	searchTests := benchmarkSearchPerformance(db, generator, projectConfig.ProjectID, testQueries)
	results.Tests = append(results.Tests, searchTests...)

	// Test 4: Concurrent searches
	fmt.Println("4️⃣  Testing concurrent searches...")
	results.Tests = append(results.Tests, benchmarkConcurrentSearches(db, generator, projectConfig.ProjectID, testQueries[0]))

	// Calculate summary
	fmt.Println()
	fmt.Println("📈 Calculating summary statistics...")
	results.Summary = calculateSummary(results.Tests)

	// Save results
	fmt.Println()
	fmt.Println("💾 Saving results...")
	if err := saveResults(results, benchmarkOutputFile); err != nil {
		return fmt.Errorf("failed to save results: %w", err)
	}

	// Print summary
	printSummary(results)

	return nil
}

func benchmarkDatabaseConnection(db *sql.DB) BenchmarkTest {
	start := time.Now()
	err := db.Ping()
	duration := time.Since(start)

	return BenchmarkTest{
		Name:     "Database Connection",
		Duration: duration,
		Success:  err == nil,
		Error:    errorString(err),
	}
}

func benchmarkEmbeddingGeneration(generator embeddings.Generator, queries []string) []BenchmarkTest {
	tests := []BenchmarkTest{}

	for i, query := range queries {
		fmt.Printf("   Embedding %d/%d: %s\n", i+1, len(queries), truncate(query, 40))
		start := time.Now()
		_, err := generator.Embed(query)
		duration := time.Since(start)

		tests = append(tests, BenchmarkTest{
			Name:     fmt.Sprintf("Embedding Generation #%d", i+1),
			Query:    query,
			Duration: duration,
			Success:  err == nil,
			Error:    errorString(err),
		})
	}

	return tests
}

func benchmarkSearchPerformance(db *sql.DB, generator embeddings.Generator, projectID string, queries []string) []BenchmarkTest {
	tests := []BenchmarkTest{}

	for i, query := range queries {
		fmt.Printf("   Search %d/%d: %s\n", i+1, len(queries), truncate(query, 40))

		// End-to-end test (embedding + search)
		start := time.Now()

		// Generate embedding
		embedding, err := generator.Embed(query)
		if err != nil {
			tests = append(tests, BenchmarkTest{
				Name:     fmt.Sprintf("Search #%d (E2E)", i+1),
				Query:    query,
				Duration: time.Since(start),
				Success:  false,
				Error:    fmt.Sprintf("embedding failed: %v", err),
			})
			continue
		}

		// Search
		results, err := searchSimilarChunks(db, projectID, embedding, 5)
		duration := time.Since(start)

		var topSim, avgSim float64
		if len(results) > 0 {
			topSim = results[0].Similarity
			sum := 0.0
			for _, r := range results {
				sum += r.Similarity
			}
			avgSim = sum / float64(len(results))
		}

		tests = append(tests, BenchmarkTest{
			Name:              fmt.Sprintf("Search #%d (E2E)", i+1),
			Query:             query,
			Duration:          duration,
			Success:           err == nil,
			Error:             errorString(err),
			ResultCount:       len(results),
			TopSimilarity:     topSim,
			AverageSimilarity: avgSim,
		})
	}

	return tests
}

func benchmarkConcurrentSearches(db *sql.DB, generator embeddings.Generator, projectID string, query string) BenchmarkTest {
	concurrent := 5
	start := time.Now()

	// Generate embedding once
	embedding, err := generator.Embed(query)
	if err != nil {
		return BenchmarkTest{
			Name:     "Concurrent Searches",
			Query:    query,
			Duration: time.Since(start),
			Success:  false,
			Error:    fmt.Sprintf("embedding failed: %v", err),
		}
	}

	// Run concurrent searches
	done := make(chan error, concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			_, err := searchSimilarChunks(db, projectID, embedding, 5)
			done <- err
		}()
	}

	// Wait for all
	errors := 0
	for i := 0; i < concurrent; i++ {
		if err := <-done; err != nil {
			errors++
		}
	}

	duration := time.Since(start)

	return BenchmarkTest{
		Name:     fmt.Sprintf("Concurrent Searches (x%d)", concurrent),
		Query:    query,
		Duration: duration,
		Success:  errors == 0,
		Error:    fmt.Sprintf("%d/%d failed", errors, concurrent),
	}
}

func calculateSummary(tests []BenchmarkTest) BenchmarkSummary {
	summary := BenchmarkSummary{}

	var embeddingDurations []time.Duration
	var searchDurations []time.Duration
	var relevanceScores []float64

	for _, test := range tests {
		summary.TotalTests++
		if test.Success {
			summary.SuccessfulTests++
		} else {
			summary.FailedTests++
		}

		if test.Name[:9] == "Embedding" {
			embeddingDurations = append(embeddingDurations, test.Duration)
		}

		if test.Name[:6] == "Search" {
			searchDurations = append(searchDurations, test.Duration)
			if test.TopSimilarity > 0 {
				relevanceScores = append(relevanceScores, test.TopSimilarity)
			}
		}
	}

	// Calculate averages
	if len(embeddingDurations) > 0 {
		var sum time.Duration
		for _, d := range embeddingDurations {
			sum += d
		}
		summary.AverageEmbeddingTime = sum / time.Duration(len(embeddingDurations))
	}

	if len(searchDurations) > 0 {
		var sum time.Duration
		min := searchDurations[0]
		max := searchDurations[0]

		for _, d := range searchDurations {
			sum += d
			if d < min {
				min = d
			}
			if d > max {
				max = d
			}
		}

		summary.AverageSearchTime = sum / time.Duration(len(searchDurations))
		summary.AverageEndToEndTime = summary.AverageSearchTime // Same for now
		summary.MinSearchTime = min
		summary.MaxSearchTime = max

		// Calculate throughput
		totalTime := sum.Seconds()
		if totalTime > 0 {
			summary.ThroughputQueriesPerSec = float64(len(searchDurations)) / totalTime
		}
	}

	if len(relevanceScores) > 0 {
		sum := 0.0
		for _, score := range relevanceScores {
			sum += score
		}
		summary.AverageResultRelevance = sum / float64(len(relevanceScores))
	}

	return summary
}

func saveResults(results *BenchmarkResults, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func printSummary(results *BenchmarkResults) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("📊 BENCHMARK RESULTS")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	s := results.Summary

	fmt.Printf("✅ Success Rate: %d/%d tests (%.1f%%)\n",
		s.SuccessfulTests, s.TotalTests,
		float64(s.SuccessfulTests)/float64(s.TotalTests)*100)
	fmt.Println()

	fmt.Println("⚡ Performance:")
	fmt.Printf("   Avg Embedding Time:  %v\n", s.AverageEmbeddingTime)
	fmt.Printf("   Avg Search Time:     %v\n", s.AverageSearchTime)
	fmt.Printf("   Min Search Time:     %v\n", s.MinSearchTime)
	fmt.Printf("   Max Search Time:     %v\n", s.MaxSearchTime)
	fmt.Printf("   Throughput:          %.2f queries/sec\n", s.ThroughputQueriesPerSec)
	fmt.Println()

	fmt.Println("🎯 Relevance:")
	fmt.Printf("   Avg Top Result:      %.1f%%\n", s.AverageResultRelevance*100)
	fmt.Println()

	fmt.Printf("💾 Full results saved to: %s\n", benchmarkOutputFile)
	fmt.Println()

	// Performance rating
	fmt.Println("📈 Performance Rating:")
	if s.AverageSearchTime < 100*time.Millisecond {
		fmt.Println("   🚀 EXCELLENT - Blazing fast!")
	} else if s.AverageSearchTime < 500*time.Millisecond {
		fmt.Println("   ✅ GOOD - Fast enough for real-time")
	} else if s.AverageSearchTime < 1*time.Second {
		fmt.Println("   ⚠️  ACCEPTABLE - Usable but could be faster")
	} else {
		fmt.Println("   ❌ SLOW - Consider optimization")
	}

	if s.AverageResultRelevance > 0.8 {
		fmt.Println("   🎯 HIGH RELEVANCE - Results are very pertinent")
	} else if s.AverageResultRelevance > 0.6 {
		fmt.Println("   ✅ GOOD RELEVANCE - Results are useful")
	} else {
		fmt.Println("   ⚠️  LOW RELEVANCE - Consider re-indexing or tuning")
	}
	fmt.Println()
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
