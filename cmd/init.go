package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/agents"
	"github.com/yourusername/oview/internal/claude"
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
	var oldConfig *config.ProjectConfig
	configExists := false

	if _, err := os.Stat(oviewDir); err == nil {
		if !forceInit {
			return fmt.Errorf(".oview directory already exists. Use --force to overwrite")
		}

		// Load old config to compare embeddings
		oldConfig, err = config.LoadProjectConfig(projectPath)
		if err == nil {
			configExists = true
		}
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
			Provider: "openai",
			Model:    "text-embedding-3-small",
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
		Embeddings: embeddingsConfig,
		LLM:        llmConfig,
	}

	// Check if embeddings model changed
	if configExists && oldConfig != nil {
		embeddingsChanged := oldConfig.Embeddings.Model != embeddingsConfig.Model ||
			oldConfig.Embeddings.Dim != embeddingsConfig.Dim ||
			oldConfig.Embeddings.Provider != embeddingsConfig.Provider

		if embeddingsChanged {
			fmt.Println()
			fmt.Println("⚠️  ATTENTION: Le modèle d'embeddings a changé!")
			fmt.Println()
			fmt.Printf("   Ancien: %s / %s (%d dimensions)\n",
				oldConfig.Embeddings.Provider,
				oldConfig.Embeddings.Model,
				oldConfig.Embeddings.Dim)
			fmt.Printf("   Nouveau: %s / %s (%d dimensions)\n",
				embeddingsConfig.Provider,
				embeddingsConfig.Model,
				embeddingsConfig.Dim)
			fmt.Println()
			fmt.Println("⚠️  Vous devez recréer la base de données:")
			fmt.Println()
			fmt.Println("   1. Supprimer l'ancienne base:")
			fmt.Printf("      docker exec oview-postgres psql -U oview -c \"DROP DATABASE oview_%s;\"\n", slug)
			fmt.Println()
			fmt.Println("   2. Recréer avec la nouvelle dimension:")
			fmt.Println("      oview up")
			fmt.Println()
			fmt.Println("   3. Réindexer:")
			fmt.Println("      oview index")
			fmt.Println()
		}
	}

	if err := projectConfig.Save(projectPath); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}
	fmt.Printf("   ✓ Project config saved (slug: %s)\n", slug)

	// Create RAG config
	fmt.Println("📋 Creating RAG configuration...")
	ragConfig := config.DefaultRAGConfig(stack)
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

	// Claude Code integration
	var claudeMdStatus claude.ClaudeMDStatus
	var enableClaudeIntegration bool
	var mcpAdded bool

	if !nonInteractive {
		fmt.Println()
		fmt.Println("🎯 Claude Code Integration")
		fmt.Println()
		fmt.Println("Enable Claude Code integration with RAG-first MCP guidance?")
		fmt.Println("This will:")
		fmt.Println("  - Ensure CLAUDE.md exists (via claude /init or fallback)")
		fmt.Println("  - Add oview RAG-first policy to CLAUDE.md")
		fmt.Println("  - Create .oview/claude_mcp.json with MCP configuration")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		choice := promptChoice(reader, "Enable Claude Code integration? [Y/n]", []string{"y", "n", "yes", "no"}, "y")
		enableClaudeIntegration = choice == "y" || choice == "yes"
	} else {
		// Non-interactive default: enabled
		enableClaudeIntegration = true
	}

	if enableClaudeIntegration {
		fmt.Println()
		fmt.Println("📝 Setting up Claude Code integration...")

		// Check if user chose Claude Code as AI assistant
		if llmConfig.Provider == "claude-code" {
			// Verify CLAUDE.md exists - required for Claude Code integration
			claudeMdPath := filepath.Join(projectPath, "CLAUDE.md")
			if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
				fmt.Println()
				fmt.Println("❌ CLAUDE.md not found!")
				fmt.Println()
				fmt.Println("Claude Code requires a CLAUDE.md file for proper integration.")
				fmt.Println()
				fmt.Println("Please run the following command first:")
				fmt.Println()
				fmt.Println("   claude /init")
				fmt.Println()
				fmt.Println("Then run 'oview init --force' again to complete the setup.")
				fmt.Println()
				return fmt.Errorf("CLAUDE.md required for Claude Code integration")
			}
		}

		// 1. Ensure CLAUDE.md exists
		status, err := claude.EnsureClaudeMd(projectPath)
		if err != nil {
			fmt.Printf("   ⚠️  Warning: Failed to ensure CLAUDE.md: %v\n", err)
		} else {
			claudeMdStatus = status
			switch status {
			case claude.StatusAlreadyExists:
				fmt.Println("   ✓ CLAUDE.md already exists")
			case claude.StatusCreatedViaCLI:
				fmt.Println("   ✓ CLAUDE.md created via Claude Code")
			case claude.StatusCreatedFallback:
				fmt.Println("   ✓ CLAUDE.md created with minimal template")
			}
		}

		// 2. Add/update oview RAG-first section
		if err := claude.UpsertOviewRagFirstSection(projectPath); err != nil {
			fmt.Printf("   ⚠️  Warning: Failed to update CLAUDE.md with RAG policy: %v\n", err)
		} else {
			fmt.Println("   ✓ oview RAG-first policy added to CLAUDE.md")
		}

		// 3. Create MCP snippet file
		if err := claude.WriteClaudeMcpSnippet(projectPath); err != nil {
			fmt.Printf("   ⚠️  Warning: Failed to create MCP snippet: %v\n", err)
		} else {
			fmt.Println("   ✓ MCP configuration snippet created")
		}

		// 4. Prompt to add to ~/.claude/mcp_servers.json
		if !nonInteractive {
			fmt.Println()
			reader := bufio.NewReader(os.Stdin)
			choice := promptChoice(reader, "Add oview MCP server to ~/.claude/mcp_servers.json automatically? [Y/n]", []string{"y", "n", "yes", "no"}, "y")
			if choice == "y" || choice == "yes" {
				if err := claude.AddToClaudeMcpConfig(projectPath); err != nil {
					fmt.Printf("   ⚠️  Warning: Failed to add to MCP config: %v\n", err)
					fmt.Println("   You can add it manually from .oview/claude_mcp.json")
				} else {
					fmt.Println("   ✓ MCP configuration added to ~/.claude/mcp_servers.json")
					mcpAdded = true
				}
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("✅ Initialization complete!")
	fmt.Println()
	fmt.Println("Created:")
	fmt.Println("  .oview/project.yaml     - Project configuration")
	fmt.Println("  .oview/rag.yaml         - RAG indexing rules")
	fmt.Println("  .oview/agents/          - Claude agent instructions")
	fmt.Println("  .oview/index/           - Index metadata (empty)")

	if enableClaudeIntegration {
		fmt.Println()
		fmt.Println("Claude Code Integration:")
		switch claudeMdStatus {
		case claude.StatusAlreadyExists:
			fmt.Println("  CLAUDE.md               - Already exists (updated with RAG policy)")
		case claude.StatusCreatedViaCLI:
			fmt.Println("  CLAUDE.md               - Created via Claude Code (enriched)")
		case claude.StatusCreatedFallback:
			fmt.Println("  CLAUDE.md               - Created with minimal template")
		}
		fmt.Println("  .oview/claude_mcp.json  - MCP configuration snippet")

		if mcpAdded {
			fmt.Println("  ~/.claude/mcp_servers.json - MCP server configuration updated")
			fmt.Println()
			fmt.Println("To complete Claude Code integration:")
			fmt.Println("  1. Restart Claude Code to load the MCP server")
		} else {
			fmt.Println()
			fmt.Println("To complete Claude Code integration:")
			fmt.Println("  1. Copy MCP config from .oview/claude_mcp.json")
			fmt.Println("  2. Add to your ~/.claude/mcp_servers.json")
			fmt.Println("  3. Restart Claude Code to load the MCP server")
		}
	}

	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review and customize .oview/project.yaml if needed")
	fmt.Println("  2. Run: oview up")
	fmt.Println("  3. Run: oview index")

	return nil
}

// generateProjectID generates a random project ID
func generateProjectID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// EmbeddingModelOption represents an embedding model choice
type EmbeddingModelOption struct {
	Name        string
	Provider    string
	Dim         int
	Description string
	BaseURL     string
}

// promptEmbeddingsConfig prompts user for embeddings configuration
func promptEmbeddingsConfig() config.EmbeddingsConfig {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("📊 Configuration des embeddings (vecteurs sémantiques)")
	fmt.Println()
	fmt.Println("Les embeddings permettent la recherche sémantique dans votre code.")
	fmt.Println()

	// All embedding models in one flat list - Ollama first, official recommended models
	models := []EmbeddingModelOption{
		// Ollama models first - official recommended embeddings models from ollama.com
		{"nomic-embed-text", "ollama", 768, "Ollama - 768 dim, 8K context, local, gratuit (recommandé)", "http://localhost:11434"},
		{"mxbai-embed-large", "ollama", 1024, "Ollama - 1024 dim, haute qualité, local", "http://localhost:11434"},
		{"snowflake-arctic-embed", "ollama", 1024, "Ollama - 1024 dim, Snowflake, local", "http://localhost:11434"},
		{"embeddinggemma", "ollama", 768, "Ollama - 768 dim, Google Gemma, local", "http://localhost:11434"},
		// OpenAI models
		{"text-embedding-3-small", "openai", 1536, "OpenAI - $0.02/1M tokens, 1536 dim", ""},
		{"text-embedding-3-large", "openai", 3072, "OpenAI - $0.13/1M tokens, 3072 dim, haute qualité", ""},
		{"text-embedding-ada-002", "openai", 1536, "OpenAI - $0.10/1M tokens, 1536 dim (ancien)", ""},
	}

	fmt.Println("Modèles disponibles:")
	for i, m := range models {
		fmt.Printf("  %d. %-28s - %s\n", i+1, m.Name, m.Description)
	}
	fmt.Println()

	// Get user choice
	modelNames := make([]string, len(models))
	for i, m := range models {
		modelNames[i] = m.Name
	}

	choice := promptChoice(reader, "Choisir modèle [1-7]", modelNames, "1")

	// Find selected model
	var selected EmbeddingModelOption
	for _, m := range models {
		if m.Name == choice {
			selected = m
			break
		}
	}

	// Customize base URL for Ollama if needed
	if selected.Provider == "ollama" {
		fmt.Println()
		selected.BaseURL = promptString(reader, "Base URL Ollama", selected.BaseURL)
	}

	// Validate connection
	fmt.Println()
	fmt.Println("🔌 Validation de la connexion...")

	embeddingsConfig := config.EmbeddingsConfig{
		Provider: selected.Provider,
		Model:    selected.Name,
		Dim:      selected.Dim,
		BaseURL:  selected.BaseURL,
	}

	if err := validateEmbeddingsConnection(reader, &embeddingsConfig); err != nil {
		fmt.Printf("⚠️  Validation échouée: %v\n", err)
		fmt.Println("   Vous pourrez reconfigurer plus tard en éditant .oview/project.yaml")
	}

	return embeddingsConfig
}

// AIAssistantOption represents an AI assistant choice
type AIAssistantOption struct {
	Name        string
	Provider    string
	Description string
	BaseURL     string
}

// promptLLMConfig prompts user for AI assistant configuration
func promptLLMConfig() config.LLMConfig {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("🤖 AI Assistant Configuration")
	fmt.Println()
	fmt.Println("Choose the AI assistant for code analysis and generation.")
	fmt.Println()

	// Simple list: Claude Code (recommended) and Claude API
	models := []AIAssistantOption{
		{"claude-sonnet-4.5", "claude-code", "Claude Code CLI - Integrated (recommended)", ""},
		{"claude-sonnet-4.5", "claude-api", "Claude API - Requires API key", ""},
	}

	fmt.Println("Available assistants:")
	for i, m := range models {
		fmt.Printf("  %d. %-30s - %s\n", i+1, m.Provider, m.Description)
	}
	fmt.Println()

	// Get user choice
	modelNames := make([]string, len(models))
	for i, m := range models {
		modelNames[i] = m.Name
	}

	choiceInput := promptChoice(reader, "Choose assistant [1-2]", modelNames, "1")

	// Find selected model - need to handle index-based selection
	var selected AIAssistantOption
	if choiceNum, err := strconv.Atoi(choiceInput); err == nil && choiceNum >= 1 && choiceNum <= len(models) {
		selected = models[choiceNum-1]
	} else {
		// Try to match by name
		for _, m := range models {
			if m.Name == choiceInput {
				selected = m
				break
			}
		}
		if selected.Name == "" {
			selected = models[0] // Default to first option
		}
	}

	// Validate connection
	fmt.Println()
	fmt.Println("🔌 Connection validation...")

	llmConfig := config.LLMConfig{
		Provider: selected.Provider,
		Model:    selected.Name,
		BaseURL:  selected.BaseURL,
	}

	if err := validateLLMConnection(reader, &llmConfig); err != nil {
		fmt.Printf("⚠️  Validation échouée: %v\n", err)
		fmt.Println("   Vous pourrez reconfigurer plus tard en éditant .oview/project.yaml")
	}

	return llmConfig
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

// validateEmbeddingsConnection validates the embeddings configuration
func validateEmbeddingsConnection(reader *bufio.Reader, cfg *config.EmbeddingsConfig) error {
	switch cfg.Provider {
	case "openai":
		return validateOpenAIEmbeddings(reader, cfg)
	case "ollama":
		return validateOllamaEmbeddings(reader, cfg)
	default:
		return fmt.Errorf("provider non supporté: %s", cfg.Provider)
	}
}

// validateLLMConnection validates the LLM configuration
func validateLLMConnection(reader *bufio.Reader, cfg *config.LLMConfig) error {
	switch cfg.Provider {
	case "claude-code":
		fmt.Println("✅ Claude Code: Utilise le CLI actuel (déjà authentifié)")
		return nil
	case "claude-api":
		return validateClaudeAPI(reader, cfg)
	case "openai":
		return validateOpenAILLM(reader, cfg)
	case "ollama":
		return validateOllamaLLM(reader, cfg)
	default:
		return fmt.Errorf("provider non supporté: %s", cfg.Provider)
	}
}

// validateOpenAIEmbeddings validates OpenAI embeddings connection
func validateOpenAIEmbeddings(reader *bufio.Reader, cfg *config.EmbeddingsConfig) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" && cfg.APIKey == "" {
		fmt.Println("❌ OPENAI_API_KEY non configurée")
		return promptForOpenAIKey(reader, cfg)
	}

	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	// Test connection with a minimal embedding request
	if err := testOpenAIConnection(apiKey, cfg.Model); err != nil {
		fmt.Printf("❌ Échec de connexion: %v\n", err)
		return promptForOpenAIKey(reader, cfg)
	}

	fmt.Println("✅ Connexion OpenAI réussie")
	return nil
}

// validateClaudeAPI validates Claude API connection
func validateClaudeAPI(reader *bufio.Reader, cfg *config.LLMConfig) error {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" && cfg.APIKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY non configurée")
		return promptForClaudeKey(reader, cfg)
	}

	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	// Test connection
	if err := testClaudeConnection(apiKey); err != nil {
		fmt.Printf("❌ Échec de connexion: %v\n", err)
		return promptForClaudeKey(reader, cfg)
	}

	fmt.Println("✅ Connexion Claude API réussie")
	return nil
}

// validateOpenAILLM validates OpenAI LLM connection
func validateOpenAILLM(reader *bufio.Reader, cfg *config.LLMConfig) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" && cfg.APIKey == "" {
		fmt.Println("❌ OPENAI_API_KEY non configurée")
		return promptForOpenAIKeyLLM(reader, cfg)
	}

	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	// Test connection
	if err := testOpenAILLMConnection(apiKey, cfg.Model); err != nil {
		fmt.Printf("❌ Échec de connexion: %v\n", err)
		return promptForOpenAIKeyLLM(reader, cfg)
	}

	fmt.Println("✅ Connexion OpenAI réussie")
	return nil
}

// validateOllamaEmbeddings validates Ollama embeddings connection
func validateOllamaEmbeddings(reader *bufio.Reader, cfg *config.EmbeddingsConfig) error {
	// Check if ollama command exists
	if !isOllamaInstalled() {
		fmt.Println("❌ Ollama n'est pas installé")
		if err := promptInstallOllama(reader); err != nil {
			return err
		}
		// Re-check after installation
		if !isOllamaInstalled() {
			return fmt.Errorf("ollama non installé")
		}
	}

	// Test if Ollama is running
	if err := testOllamaConnection(cfg.BaseURL); err != nil {
		fmt.Printf("❌ Ollama n'est pas lancé sur %s\n", cfg.BaseURL)
		fmt.Println()

		// Try to start ollama serve
		if err := promptStartOllama(reader); err != nil {
			return err
		}

		// Wait and re-test
		time.Sleep(2 * time.Second)
		if err := testOllamaConnection(cfg.BaseURL); err != nil {
			return fmt.Errorf("échec de connexion après lancement: %w", err)
		}
	}

	// Test if model is available
	if err := testOllamaModel(cfg.BaseURL, cfg.Model); err != nil {
		fmt.Printf("❌ Modèle %s non disponible\n", cfg.Model)

		// Propose to pull the model
		if err := promptPullOllamaModel(reader, cfg.Model); err != nil {
			return err
		}

		// Re-check after pull
		if err := testOllamaModel(cfg.BaseURL, cfg.Model); err != nil {
			return fmt.Errorf("modèle toujours non disponible: %w", err)
		}
	}

	fmt.Println("✅ Ollama connecté, modèle disponible")
	return nil
}

// validateOllamaLLM validates Ollama LLM connection
func validateOllamaLLM(reader *bufio.Reader, cfg *config.LLMConfig) error {
	// Check if ollama command exists
	if !isOllamaInstalled() {
		fmt.Println("❌ Ollama n'est pas installé")
		if err := promptInstallOllama(reader); err != nil {
			return err
		}
		// Re-check after installation
		if !isOllamaInstalled() {
			return fmt.Errorf("ollama non installé")
		}
	}

	// Test if Ollama is running
	if err := testOllamaConnection(cfg.BaseURL); err != nil {
		fmt.Printf("❌ Ollama n'est pas lancé sur %s\n", cfg.BaseURL)
		fmt.Println()

		// Try to start ollama serve
		if err := promptStartOllama(reader); err != nil {
			return err
		}

		// Wait and re-test
		time.Sleep(2 * time.Second)
		if err := testOllamaConnection(cfg.BaseURL); err != nil {
			return fmt.Errorf("échec de connexion après lancement: %w", err)
		}
	}

	// Test if model is available
	if err := testOllamaModel(cfg.BaseURL, cfg.Model); err != nil {
		fmt.Printf("❌ Modèle %s non disponible\n", cfg.Model)

		// Propose to pull the model
		if err := promptPullOllamaModel(reader, cfg.Model); err != nil {
			return err
		}

		// Re-check after pull
		if err := testOllamaModel(cfg.BaseURL, cfg.Model); err != nil {
			return fmt.Errorf("modèle toujours non disponible: %w", err)
		}
	}

	fmt.Println("✅ Ollama connecté, modèle disponible")
	return nil
}

// promptForOpenAIKey prompts user to configure OpenAI API key
func promptForOpenAIKey(reader *bufio.Reader, cfg *config.EmbeddingsConfig) error {
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  1. Entrer la clé API maintenant (sera stockée dans .oview/project.yaml)")
	fmt.Println("  2. Obtenir une clé API (ouvre le navigateur)")
	fmt.Println("  3. Configurer plus tard")
	fmt.Println()

	choice := promptChoice(reader, "Choisir [1-3]", []string{"1", "2", "3"}, "3")

	switch choice {
	case "1":
		fmt.Print("Entrez votre clé API OpenAI: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey != "" {
			cfg.APIKey = apiKey
			fmt.Println("⚠️  Clé stockée dans .oview/project.yaml - Ne pas commiter ce fichier!")
			// Test again
			if err := testOpenAIConnection(apiKey, cfg.Model); err != nil {
				return fmt.Errorf("clé invalide: %w", err)
			}
			fmt.Println("✅ Clé validée")
		}
	case "2":
		openBrowser("https://platform.openai.com/api-keys")
		fmt.Println("📖 Page ouverte dans le navigateur")
		fmt.Println("   Après avoir obtenu votre clé, configurez: export OPENAI_API_KEY='...'")
	case "3":
		fmt.Println("💡 Configurez plus tard: export OPENAI_API_KEY='...'")
	}

	return nil
}

// promptForOpenAIKeyLLM prompts user to configure OpenAI API key for LLM
func promptForOpenAIKeyLLM(reader *bufio.Reader, cfg *config.LLMConfig) error {
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  1. Entrer la clé API maintenant (sera stockée dans .oview/project.yaml)")
	fmt.Println("  2. Obtenir une clé API (ouvre le navigateur)")
	fmt.Println("  3. Configurer plus tard")
	fmt.Println()

	choice := promptChoice(reader, "Choisir [1-3]", []string{"1", "2", "3"}, "3")

	switch choice {
	case "1":
		fmt.Print("Entrez votre clé API OpenAI: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey != "" {
			cfg.APIKey = apiKey
			fmt.Println("⚠️  Clé stockée dans .oview/project.yaml - Ne pas commiter ce fichier!")
			// Test again
			if err := testOpenAILLMConnection(apiKey, cfg.Model); err != nil {
				return fmt.Errorf("clé invalide: %w", err)
			}
			fmt.Println("✅ Clé validée")
		}
	case "2":
		openBrowser("https://platform.openai.com/api-keys")
		fmt.Println("📖 Page ouverte dans le navigateur")
		fmt.Println("   Après avoir obtenu votre clé, configurez: export OPENAI_API_KEY='...'")
	case "3":
		fmt.Println("💡 Configurez plus tard: export OPENAI_API_KEY='...'")
	}

	return nil
}

// promptForClaudeKey prompts user to configure Claude API key
func promptForClaudeKey(reader *bufio.Reader, cfg *config.LLMConfig) error {
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  1. Entrer la clé API maintenant (sera stockée dans .oview/project.yaml)")
	fmt.Println("  2. Obtenir une clé API (ouvre le navigateur)")
	fmt.Println("  3. Configurer plus tard")
	fmt.Println()

	choice := promptChoice(reader, "Choisir [1-3]", []string{"1", "2", "3"}, "3")

	switch choice {
	case "1":
		fmt.Print("Entrez votre clé API Anthropic: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey != "" {
			cfg.APIKey = apiKey
			fmt.Println("⚠️  Clé stockée dans .oview/project.yaml - Ne pas commiter ce fichier!")
			// Test again
			if err := testClaudeConnection(apiKey); err != nil {
				return fmt.Errorf("clé invalide: %w", err)
			}
			fmt.Println("✅ Clé validée")
		}
	case "2":
		openBrowser("https://console.anthropic.com/settings/keys")
		fmt.Println("📖 Page ouverte dans le navigateur")
		fmt.Println("   Après avoir obtenu votre clé, configurez: export ANTHROPIC_API_KEY='...'")
	case "3":
		fmt.Println("💡 Configurez plus tard: export ANTHROPIC_API_KEY='...'")
	}

	return nil
}

// testOpenAIConnection tests OpenAI API connection
func testOpenAIConnection(apiKey, model string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := strings.NewReader(`{"input":"test","model":"` + model + `"}`)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connexion échouée: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// testOpenAILLMConnection tests OpenAI LLM API connection
func testOpenAILLMConnection(apiKey, model string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := strings.NewReader(`{"model":"` + model + `","messages":[{"role":"user","content":"test"}],"max_tokens":5}`)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connexion échouée: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// testClaudeConnection tests Claude API connection
func testClaudeConnection(apiKey string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := strings.NewReader(`{"model":"claude-3-5-sonnet-20241022","max_tokens":5,"messages":[{"role":"user","content":"test"}]}`)
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connexion échouée: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// testOllamaConnection tests if Ollama is running
func testOllamaConnection(baseURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return fmt.Errorf("connexion échouée: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("code %d", resp.StatusCode)
	}

	return nil
}

// testOllamaModel tests if a model is available in Ollama
func testOllamaModel(baseURL, modelName string) error {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	for _, m := range result.Models {
		if strings.HasPrefix(m.Name, modelName) {
			return nil
		}
	}

	return fmt.Errorf("modèle non trouvé")
}

// openBrowser opens a URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("plateforme non supportée")
	}

	return cmd.Start()
}

// isOllamaInstalled checks if ollama command is available
func isOllamaInstalled() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

// promptInstallOllama prompts user to install Ollama
func promptInstallOllama(reader *bufio.Reader) error {
	fmt.Println()
	fmt.Println("Ollama est requis pour utiliser des modèles locaux.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  1. Installer Ollama maintenant (recommandé)")
	fmt.Println("  2. Installer manuellement plus tard")
	fmt.Println()

	choice := promptChoice(reader, "Choisir [1-2]", []string{"1", "2"}, "1")

	if choice == "1" {
		fmt.Println()
		fmt.Println("🔧 Installation d'Ollama...")

		switch runtime.GOOS {
		case "linux":
			// Use the official install script
			cmd := exec.Command("bash", "-c", "curl -fsSL https://ollama.com/install.sh | sh")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("échec d'installation: %w", err)
			}
			fmt.Println("✅ Ollama installé")

		case "darwin":
			fmt.Println("Pour macOS, installez Ollama avec:")
			fmt.Println()
			fmt.Println("Option 1 - Homebrew:")
			fmt.Println("  brew install ollama")
			fmt.Println()
			fmt.Println("Option 2 - Application:")
			fmt.Println("  Télécharger depuis https://ollama.com/download")
			fmt.Println()
			openBrowser("https://ollama.com/download")
			fmt.Println("📖 Page ouverte dans le navigateur")
			fmt.Println()
			fmt.Println("Appuyez sur Entrée après l'installation...")
			reader.ReadString('\n')

		default:
			openBrowser("https://ollama.com/download")
			fmt.Println("📖 Page de téléchargement ouverte")
			fmt.Println()
			fmt.Println("Appuyez sur Entrée après l'installation...")
			reader.ReadString('\n')
		}

		return nil
	}

	fmt.Println()
	fmt.Println("💡 Installez Ollama plus tard: https://ollama.com/download")
	return fmt.Errorf("ollama non installé")
}

// promptStartOllama prompts user to start Ollama
func promptStartOllama(reader *bufio.Reader) error {
	fmt.Println("Options:")
	fmt.Println("  1. Lancer Ollama maintenant (en arrière-plan)")
	fmt.Println("  2. Lancer manuellement plus tard")
	fmt.Println()

	choice := promptChoice(reader, "Choisir [1-2]", []string{"1", "2"}, "1")

	if choice == "1" {
		fmt.Println()
		fmt.Println("🚀 Lancement d'Ollama...")

		// Start ollama serve in background
		cmd := exec.Command("ollama", "serve")
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("échec de lancement: %w", err)
		}

		fmt.Println("✅ Ollama lancé en arrière-plan")
		return nil
	}

	fmt.Println()
	fmt.Println("💡 Lancez Ollama plus tard avec: ollama serve")
	return fmt.Errorf("ollama non lancé")
}

// promptPullOllamaModel prompts user to pull an Ollama model
func promptPullOllamaModel(reader *bufio.Reader, model string) error {
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  1. Télécharger le modèle maintenant (recommandé)")
	fmt.Println("  2. Télécharger manuellement plus tard")
	fmt.Println()

	choice := promptChoice(reader, "Choisir [1-2]", []string{"1", "2"}, "1")

	if choice == "1" {
		fmt.Println()
		fmt.Printf("📥 Téléchargement du modèle %s...\n", model)
		fmt.Println("   (Cela peut prendre quelques minutes)")
		fmt.Println()

		// Pull the model
		cmd := exec.Command("ollama", "pull", model)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("échec du téléchargement: %w", err)
		}

		fmt.Println()
		fmt.Println("✅ Modèle téléchargé")
		return nil
	}

	fmt.Println()
	fmt.Printf("💡 Téléchargez le modèle plus tard avec: ollama pull %s\n", model)
	return fmt.Errorf("modèle non téléchargé")
}
