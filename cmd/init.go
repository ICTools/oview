package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/agents"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/detector"
)

var (
	forceInit       bool
	nonInteractive  bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize oview for the current project",
	Long: `Detects the project stack and creates the .oview/ directory structure:
- Detects Symfony, Docker, Makefile, frontend stack
- Creates .oview/project.yaml with project configuration
- Creates .oview/rag.yaml with chunking rules
- Generates Claude agent instruction files in .oview/agents/`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&forceInit, "force", false, "Overwrite existing .oview configuration")
	initCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Skip interactive prompts (use defaults)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("🔍 Initializing oview for this project...")
	fmt.Println()

	// Check if .oview already exists
	oviewDir := filepath.Join(projectPath, ".oview")
	if _, err := os.Stat(oviewDir); err == nil && !forceInit {
		return fmt.Errorf(".oview directory already exists. Use --force to overwrite")
	}

	// Create .oview directory structure
	fmt.Println("📁 Creating .oview directory structure...")
	dirs := []string{
		".oview",
		".oview/agents",
		".oview/index",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	fmt.Println("   ✓ Directory structure created")

	// Detect stack
	fmt.Println("🔎 Detecting project stack...")
	detect := detector.New(projectPath)
	stack, err := detect.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect stack: %w", err)
	}

	// Print detected stack
	fmt.Println("   ✓ Stack detected:")
	if stack.Symfony {
		fmt.Println("     - Symfony")
	}
	if stack.Docker {
		fmt.Println("     - Docker")
	}
	if stack.Makefile {
		fmt.Println("     - Makefile")
	}
	if stack.Frontend.Detected {
		fmt.Printf("     - Frontend: %v\n", stack.Frontend.Frameworks)
	}
	if len(stack.Languages) > 0 {
		fmt.Printf("     - Languages: %v\n", stack.Languages)
	}
	if stack.Infrastructure.Redis || stack.Infrastructure.RabbitMQ || stack.Infrastructure.Elasticsearch {
		infra := []string{}
		if stack.Infrastructure.Redis {
			infra = append(infra, "Redis")
		}
		if stack.Infrastructure.RabbitMQ {
			infra = append(infra, "RabbitMQ")
		}
		if stack.Infrastructure.Elasticsearch {
			infra = append(infra, "Elasticsearch")
		}
		fmt.Printf("     - Infrastructure: %v\n", infra)
	}

	// Generate project slug and ID
	slug := detector.GenerateProjectSlug(projectPath)
	projectID := generateProjectID()

	// Detect commands
	commands := detect.DetectCommands(stack)

	// Interactive configuration
	var embeddingsConfig config.EmbeddingsConfig
	var llmConfig config.LLMConfig

	if !nonInteractive {
		fmt.Println()
		fmt.Println("🤖 Configuration interactive")
		fmt.Println()

		// Prompt for embeddings
		embeddingsConfig = promptEmbeddingsConfig()

		// Prompt for LLM
		llmConfig = promptLLMConfig()
	} else {
		// Non-interactive defaults
		embeddingsConfig = config.EmbeddingsConfig{
			Provider: "stub",
			Model:    "stub-hash-based",
			Dim:      1536,
		}
		llmConfig = config.LLMConfig{
			Provider: "claude-code",
			Model:    "claude-sonnet-4.5",
		}
	}

	// Create project config
	fmt.Println()
	fmt.Println("📝 Creating project configuration...")
	projectConfig := &config.ProjectConfig{
		ProjectID:   projectID,
		ProjectSlug: slug,
		Stack:       *stack,
		Commands:    commands,
		Trello: config.TrelloConfig{
			BoardID: "",
			ListIDs: map[string]string{
				"backlog":        "",
				"todo":           "",
				"in_progress":    "",
				"review":         "",
				"done":           "",
			},
		},
		Embeddings: embeddingsConfig,
		LLM:        llmConfig,
	}

	if err := projectConfig.Save(projectPath); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}
	fmt.Printf("   ✓ Project config saved (slug: %s)\n", slug)

	// Create RAG config
	fmt.Println("📋 Creating RAG configuration...")
	ragConfig := config.DefaultRAGConfig()
	if err := config.SaveRAGConfig(projectPath, ragConfig); err != nil {
		return fmt.Errorf("failed to save RAG config: %w", err)
	}
	fmt.Println("   ✓ RAG config saved")

	// Create empty manifest and stats
	fmt.Println("📊 Creating index manifests...")
	emptyJSON := []byte("{}\n")
	manifestPath := filepath.Join(projectPath, ".oview", "index", "manifest.json")
	if err := os.WriteFile(manifestPath, emptyJSON, 0644); err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}
	statsPath := filepath.Join(projectPath, ".oview", "index", "stats.json")
	if err := os.WriteFile(statsPath, emptyJSON, 0644); err != nil {
		return fmt.Errorf("failed to create stats: %w", err)
	}
	fmt.Println("   ✓ Index manifests created")

	// Generate agent files
	fmt.Println("🤖 Generating Claude agent instruction files...")
	agentGen := agents.New(projectPath, stack)
	if err := agentGen.GenerateAll(); err != nil {
		return fmt.Errorf("failed to generate agent files: %w", err)
	}
	fmt.Println("   ✓ Agent files generated")

	// Summary
	fmt.Println()
	fmt.Println("✅ Initialization complete!")
	fmt.Println()
	fmt.Println("Created:")
	fmt.Println("  .oview/project.yaml     - Project configuration")
	fmt.Println("  .oview/rag.yaml         - RAG indexing rules")
	fmt.Println("  .oview/agents/          - Claude agent instructions")
	fmt.Println("  .oview/index/           - Index metadata (empty)")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review and customize .oview/project.yaml if needed")
	fmt.Println("  2. Add Trello credentials if using Trello integration")
	fmt.Println("  3. Run: oview up")
	fmt.Println("  4. Run: oview index")

	return nil
}

// generateProjectID generates a random project ID
func generateProjectID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// promptEmbeddingsConfig prompts user for embeddings configuration
func promptEmbeddingsConfig() config.EmbeddingsConfig {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("📊 Configuration des embeddings (vecteurs sémantiques)")
	fmt.Println()
	fmt.Println("Les embeddings permettent la recherche sémantique dans votre code.")
	fmt.Println()
	fmt.Println("Providers disponibles:")
	fmt.Println("  1. stub         - Placeholder (hash, pas de sémantique) - Gratuit")
	fmt.Println("  2. openai       - OpenAI API (haute qualité) - ~$0.02/1M tokens")
	fmt.Println("  3. ollama       - Local (privé, gratuit) - Nécessite installation")
	fmt.Println()

	// Choose provider
	provider := promptChoice(reader, "Choisir provider [1-3]", []string{"stub", "openai", "ollama"}, "1")

	var model string
	var dim int
	var baseURL string

	switch provider {
	case "stub":
		model = "stub-hash-based"
		dim = 1536
		fmt.Println()
		fmt.Println("ℹ️  Stub: Pas de sémantique, uniquement pour tester l'infrastructure")

	case "openai":
		fmt.Println()
		fmt.Println("Modèles OpenAI disponibles:")
		fmt.Println("  1. text-embedding-3-small  - $0.02/1M tokens, 1536 dim (recommandé)")
		fmt.Println("  2. text-embedding-3-large  - $0.13/1M tokens, 3072 dim (meilleure qualité)")
		fmt.Println("  3. text-embedding-ada-002  - $0.10/1M tokens, 1536 dim (ancien)")
		fmt.Println()

		models := []string{
			"text-embedding-3-small",
			"text-embedding-3-large",
			"text-embedding-ada-002",
		}
		dims := []int{1536, 3072, 1536}

		choice := promptChoice(reader, "Choisir modèle [1-3]", models, "1")
		model = choice
		for i, m := range models {
			if m == choice {
				dim = dims[i]
				break
			}
		}

		fmt.Println()
		fmt.Println("💡 N'oubliez pas de configurer OPENAI_API_KEY dans votre environnement")

	case "ollama":
		fmt.Println()
		fmt.Println("Modèles Ollama populaires:")
		fmt.Println("  1. nomic-embed-text   - 768 dim, 274 MB (recommandé)")
		fmt.Println("  2. mxbai-embed-large  - 1024 dim, 669 MB")
		fmt.Println("  3. bge-code           - 768 dim, optimisé code")
		fmt.Println("  4. all-minilm         - 384 dim, 45 MB (rapide)")
		fmt.Println()

		models := []string{
			"nomic-embed-text",
			"mxbai-embed-large",
			"bge-code",
			"all-minilm",
		}
		dims := []int{768, 1024, 768, 384}

		choice := promptChoice(reader, "Choisir modèle [1-4]", models, "1")
		model = choice
		for i, m := range models {
			if m == choice {
				dim = dims[i]
				break
			}
		}

		baseURL = promptString(reader, "Base URL Ollama", "http://localhost:11434")

		fmt.Println()
		fmt.Println("💡 Avant d'indexer, lancez: ollama serve && ollama pull " + model)
	}

	return config.EmbeddingsConfig{
		Provider: provider,
		Model:    model,
		Dim:      dim,
		BaseURL:  baseURL,
	}
}

// promptLLMConfig prompts user for LLM configuration
func promptLLMConfig() config.LLMConfig {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("🤖 Configuration du LLM (agent AI)")
	fmt.Println()
	fmt.Println("Le LLM sera utilisé par les agents pour analyser et modifier le code.")
	fmt.Println()
	fmt.Println("Providers disponibles:")
	fmt.Println("  1. claude-code   - Claude Code CLI (Sonnet 4.5) - Intégré")
	fmt.Println("  2. claude-api    - Claude API (Anthropic) - Nécessite clé API")
	fmt.Println("  3. openai        - OpenAI API (GPT-4, etc.) - Nécessite clé API")
	fmt.Println("  4. ollama        - Local (Llama 3, etc.) - Gratuit")
	fmt.Println()

	// Choose provider
	provider := promptChoice(reader, "Choisir provider [1-4]", []string{"claude-code", "claude-api", "openai", "ollama"}, "1")

	var model string
	var baseURL string

	switch provider {
	case "claude-code":
		model = "claude-sonnet-4.5"
		fmt.Println()
		fmt.Println("✅ Claude Code: Utilise le CLI actuel (recommandé)")

	case "claude-api":
		fmt.Println()
		fmt.Println("Modèles Claude API:")
		fmt.Println("  1. claude-sonnet-4.5    - Dernier, équilibré (recommandé)")
		fmt.Println("  2. claude-opus-4.5      - Maximum qualité")
		fmt.Println("  3. claude-haiku-4       - Rapide et économique")
		fmt.Println()

		models := []string{
			"claude-sonnet-4.5",
			"claude-opus-4.5",
			"claude-haiku-4",
		}
		model = promptChoice(reader, "Choisir modèle [1-3]", models, "1")

		fmt.Println()
		fmt.Println("💡 Configurez ANTHROPIC_API_KEY dans votre environnement")

	case "openai":
		fmt.Println()
		fmt.Println("Modèles OpenAI:")
		fmt.Println("  1. gpt-4o           - Dernier, multimodal")
		fmt.Println("  2. gpt-4-turbo      - Rapide, fenêtre 128K")
		fmt.Println("  3. gpt-4            - Stable")
		fmt.Println()

		models := []string{
			"gpt-4o",
			"gpt-4-turbo",
			"gpt-4",
		}
		model = promptChoice(reader, "Choisir modèle [1-3]", models, "1")

		fmt.Println()
		fmt.Println("💡 Configurez OPENAI_API_KEY dans votre environnement")

	case "ollama":
		fmt.Println()
		fmt.Println("Modèles Ollama populaires:")
		fmt.Println("  1. llama3.1:70b      - Haute qualité")
		fmt.Println("  2. llama3.1:8b       - Rapide")
		fmt.Println("  3. codellama:34b     - Optimisé code")
		fmt.Println("  4. deepseek-coder    - Spécialisé code")
		fmt.Println()

		models := []string{
			"llama3.1:70b",
			"llama3.1:8b",
			"codellama:34b",
			"deepseek-coder",
		}
		model = promptChoice(reader, "Choisir modèle [1-4]", models, "2")

		baseURL = promptString(reader, "Base URL Ollama", "http://localhost:11434")

		fmt.Println()
		fmt.Println("💡 Avant d'utiliser, lancez: ollama serve && ollama pull " + model)
	}

	return config.LLMConfig{
		Provider: provider,
		Model:    model,
		BaseURL:  baseURL,
	}
}

// promptChoice prompts user to choose from a list
func promptChoice(reader *bufio.Reader, prompt string, choices []string, defaultChoice string) string {
	fmt.Printf("%s (défaut: %s): ", prompt, defaultChoice)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		input = defaultChoice
	}

	// Try to parse as number
	if choice, err := strconv.Atoi(input); err == nil && choice >= 1 && choice <= len(choices) {
		return choices[choice-1]
	}

	// Try to match as string
	for _, c := range choices {
		if strings.EqualFold(input, c) {
			return c
		}
	}

	// Default
	choiceNum, _ := strconv.Atoi(defaultChoice)
	if choiceNum >= 1 && choiceNum <= len(choices) {
		return choices[choiceNum-1]
	}

	return choices[0]
}

// promptString prompts user for a string value
func promptString(reader *bufio.Reader, prompt string, defaultValue string) string {
	fmt.Printf("%s (défaut: %s): ", prompt, defaultValue)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	return input
}
