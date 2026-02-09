package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/oview/internal/config"
)

// Detector scans a project directory to detect the technology stack
type Detector struct {
	projectPath string
}

// New creates a new Detector
func New(projectPath string) *Detector {
	return &Detector{projectPath: projectPath}
}

// Detect analyzes the project and returns stack information
func (d *Detector) Detect() (*config.StackInfo, error) {
	stack := &config.StackInfo{
		Languages:  []string{},
		Frameworks: []string{},
		Frontend: config.FrontendInfo{
			Frameworks: []string{},
			BuildTools: []string{},
		},
	}

	// Detect all programming languages
	languages := d.detectLanguages()
	stack.Languages = languages

	// Detect Symfony (PHP framework)
	if d.hasSymfony() {
		stack.Symfony = true
		stack.Frameworks = append(stack.Frameworks, "Symfony")
	}

	// Detect Docker
	if d.hasDocker() {
		stack.Docker = true
	}

	// Detect Makefile
	if d.hasMakefile() {
		stack.Makefile = true
	}

	// Detect Frontend
	frontend := d.detectFrontend()
	stack.Frontend = frontend

	// Detect Infrastructure
	stack.Infrastructure = d.detectInfrastructure()

	// Deduplicate languages and frameworks
	stack.Languages = unique(stack.Languages)
	stack.Frameworks = unique(stack.Frameworks)

	return stack, nil
}

// detectLanguages detects all programming languages in the project
func (d *Detector) detectLanguages() []string {
	languages := []string{}

	// Go
	if d.hasGo() {
		languages = append(languages, "Go")
	}

	// Python
	if d.hasPython() {
		languages = append(languages, "Python")
	}

	// Ruby
	if d.hasRuby() {
		languages = append(languages, "Ruby")
	}

	// Rust
	if d.hasRust() {
		languages = append(languages, "Rust")
	}

	// Java
	if d.hasJava() {
		languages = append(languages, "Java")
	}

	// C/C++
	if d.hasCpp() {
		languages = append(languages, "C/C++")
	}

	// C#
	if d.hasCSharp() {
		languages = append(languages, "C#")
	}

	// PHP
	if d.hasPHP() {
		languages = append(languages, "PHP")
	}

	// JavaScript/TypeScript (check file extensions, not just package.json)
	if d.hasJavaScriptFiles() {
		languages = append(languages, "JavaScript")
	}

	// TypeScript
	if d.hasTypeScriptFiles() {
		languages = append(languages, "TypeScript")
	}

	// Shell scripts
	if d.hasShell() {
		languages = append(languages, "Shell")
	}

	return languages
}

// hasGo checks if the project uses Go
func (d *Detector) hasGo() bool {
	indicators := []string{"go.mod", "go.sum"}
	for _, indicator := range indicators {
		if d.fileExists(indicator) {
			return true
		}
	}
	return d.hasFilesWithExtension(".go")
}

// hasPython checks if the project uses Python
func (d *Detector) hasPython() bool {
	indicators := []string{
		"requirements.txt",
		"setup.py",
		"pyproject.toml",
		"Pipfile",
		"poetry.lock",
		"setup.cfg",
	}
	for _, indicator := range indicators {
		if d.fileExists(indicator) {
			return true
		}
	}
	return d.hasFilesWithExtension(".py")
}

// hasRuby checks if the project uses Ruby
func (d *Detector) hasRuby() bool {
	indicators := []string{"Gemfile", "Gemfile.lock", ".ruby-version"}
	for _, indicator := range indicators {
		if d.fileExists(indicator) {
			return true
		}
	}
	return d.hasFilesWithExtension(".rb")
}

// hasRust checks if the project uses Rust
func (d *Detector) hasRust() bool {
	indicators := []string{"Cargo.toml", "Cargo.lock"}
	for _, indicator := range indicators {
		if d.fileExists(indicator) {
			return true
		}
	}
	return d.hasFilesWithExtension(".rs")
}

// hasJava checks if the project uses Java
func (d *Detector) hasJava() bool {
	indicators := []string{
		"pom.xml",
		"build.gradle",
		"build.gradle.kts",
		"settings.gradle",
		"gradlew",
	}
	for _, indicator := range indicators {
		if d.fileExists(indicator) {
			return true
		}
	}
	return d.hasFilesWithExtension(".java")
}

// hasCpp checks if the project uses C/C++
func (d *Detector) hasCpp() bool {
	indicators := []string{
		"CMakeLists.txt",
		"configure.ac",
		"meson.build",
	}
	for _, indicator := range indicators {
		if d.fileExists(indicator) {
			return true
		}
	}
	// Check for C/C++ extensions
	exts := []string{".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx"}
	for _, ext := range exts {
		if d.hasFilesWithExtension(ext) {
			return true
		}
	}
	return false
}

// hasCSharp checks if the project uses C#
func (d *Detector) hasCSharp() bool {
	indicators := []string{
		".csproj",
		".sln",
	}
	for _, indicator := range indicators {
		if d.hasFilesWithPattern(indicator) {
			return true
		}
	}
	return d.hasFilesWithExtension(".cs")
}

// hasPHP checks if the project uses PHP
func (d *Detector) hasPHP() bool {
	indicators := []string{"composer.json", "composer.lock"}
	for _, indicator := range indicators {
		if d.fileExists(indicator) {
			return true
		}
	}
	return d.hasFilesWithExtension(".php")
}

// hasJavaScriptFiles checks if the project has JavaScript files
func (d *Detector) hasJavaScriptFiles() bool {
	return d.hasFilesWithExtension(".js") || d.hasFilesWithExtension(".jsx") || d.fileExists("package.json")
}

// hasTypeScriptFiles checks if the project has TypeScript files
func (d *Detector) hasTypeScriptFiles() bool {
	return d.hasFilesWithExtension(".ts") || d.hasFilesWithExtension(".tsx") || d.fileExists("tsconfig.json")
}

// hasShell checks if the project has shell scripts
func (d *Detector) hasShell() bool {
	return d.hasFilesWithExtension(".sh") || d.hasFilesWithExtension(".bash")
}

// hasFilesWithExtension checks if any files with the given extension exist in the project
func (d *Detector) hasFilesWithExtension(ext string) bool {
	found := false
	filepath.Walk(d.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return nil
		}
		// Skip hidden directories and common ignore directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" || name == "var" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(info.Name(), ext) {
			found = true
		}
		return nil
	})
	return found
}

// hasFilesWithPattern checks if any files matching the pattern exist in the project
func (d *Detector) hasFilesWithPattern(pattern string) bool {
	found := false
	filepath.Walk(d.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return nil
		}
		// Skip hidden directories and common ignore directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" || name == "var" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.Contains(info.Name(), pattern) {
			found = true
		}
		return nil
	})
	return found
}

// hasSymfony checks if this is a Symfony project
func (d *Detector) hasSymfony() bool {
	indicators := []string{
		"symfony.lock",
		"bin/console",
		"config/bundles.php",
		"src/Kernel.php",
	}

	for _, indicator := range indicators {
		if d.fileExists(indicator) {
			return true
		}
	}

	// Check composer.json for symfony packages
	if d.fileExists("composer.json") {
		data, err := os.ReadFile(filepath.Join(d.projectPath, "composer.json"))
		if err == nil {
			if strings.Contains(string(data), "symfony/") {
				return true
			}
		}
	}

	return false
}

// hasDocker checks if the project uses Docker
func (d *Detector) hasDocker() bool {
	dockerFiles := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
		"Dockerfile",
		".dockerignore",
	}

	for _, file := range dockerFiles {
		if d.fileExists(file) {
			return true
		}
	}

	return false
}

// hasMakefile checks if the project has a Makefile
func (d *Detector) hasMakefile() bool {
	return d.fileExists("Makefile")
}

// detectFrontend detects frontend stack
func (d *Detector) detectFrontend() config.FrontendInfo {
	info := config.FrontendInfo{
		Frameworks: []string{},
		BuildTools: []string{},
	}

	// Check for package.json
	if !d.fileExists("package.json") {
		return info
	}

	info.Detected = true

	// Read package.json
	data, err := os.ReadFile(filepath.Join(d.projectPath, "package.json"))
	if err != nil {
		return info
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return info
	}

	// Detect package manager
	if d.fileExists("package-lock.json") {
		info.PackageManager = "npm"
	} else if d.fileExists("yarn.lock") {
		info.PackageManager = "yarn"
	} else if d.fileExists("pnpm-lock.yaml") {
		info.PackageManager = "pnpm"
	}

	// Check dependencies for frameworks
	deps := make(map[string]bool)
	if devDeps, ok := pkg["devDependencies"].(map[string]interface{}); ok {
		for name := range devDeps {
			deps[name] = true
		}
	}
	if dependencies, ok := pkg["dependencies"].(map[string]interface{}); ok {
		for name := range dependencies {
			deps[name] = true
		}
	}

	// Detect frameworks
	if deps["react"] || deps["react-dom"] {
		info.Frameworks = append(info.Frameworks, "React")
	}
	if deps["vue"] {
		info.Frameworks = append(info.Frameworks, "Vue")
	}
	if deps["@angular/core"] {
		info.Frameworks = append(info.Frameworks, "Angular")
	}
	if deps["svelte"] {
		info.Frameworks = append(info.Frameworks, "Svelte")
	}

	// Detect build tools
	if deps["webpack"] || deps["@symfony/webpack-encore"] {
		info.BuildTools = append(info.BuildTools, "Webpack")
	}
	if deps["vite"] {
		info.BuildTools = append(info.BuildTools, "Vite")
	}
	if deps["@symfony/webpack-encore"] {
		info.BuildTools = append(info.BuildTools, "Webpack Encore")
	}
	if deps["@hotwired/stimulus"] || deps["@symfony/stimulus-bridge"] {
		info.Frameworks = append(info.Frameworks, "Stimulus")
	}

	// Check for assets directory
	if d.dirExists("assets") {
		// Likely Symfony UX or similar
		info.Frameworks = append(info.Frameworks, "Symfony UX")
	}

	return info
}

// detectInfrastructure detects infrastructure components
func (d *Detector) detectInfrastructure() config.InfraInfo {
	info := config.InfraInfo{}

	// Check docker-compose files
	composeFiles := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
	}

	for _, file := range composeFiles {
		if data, err := os.ReadFile(filepath.Join(d.projectPath, file)); err == nil {
			content := string(data)
			if strings.Contains(content, "redis:") || strings.Contains(content, "redis/redis") {
				info.Redis = true
			}
			if strings.Contains(content, "rabbitmq:") || strings.Contains(content, "rabbitmq/rabbitmq") {
				info.RabbitMQ = true
			}
			if strings.Contains(content, "elasticsearch:") || strings.Contains(content, "elastic/elasticsearch") {
				info.Elasticsearch = true
			}
		}
	}

	// Check .env files
	envFiles := []string{".env", ".env.local", ".env.dist"}
	for _, file := range envFiles {
		if data, err := os.ReadFile(filepath.Join(d.projectPath, file)); err == nil {
			content := strings.ToUpper(string(data))
			if strings.Contains(content, "REDIS") {
				info.Redis = true
			}
			if strings.Contains(content, "RABBITMQ") || strings.Contains(content, "AMQP") {
				info.RabbitMQ = true
			}
			if strings.Contains(content, "ELASTICSEARCH") {
				info.Elasticsearch = true
			}
		}
	}

	return info
}

// fileExists checks if a file exists in the project
func (d *Detector) fileExists(path string) bool {
	fullPath := filepath.Join(d.projectPath, path)
	info, err := os.Stat(fullPath)
	return err == nil && !info.IsDir()
}

// dirExists checks if a directory exists in the project
func (d *Detector) dirExists(path string) bool {
	fullPath := filepath.Join(d.projectPath, path)
	info, err := os.Stat(fullPath)
	return err == nil && info.IsDir()
}

// unique removes duplicates from a string slice
func unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// GenerateProjectSlug generates a project slug from the directory name
func GenerateProjectSlug(projectPath string) string {
	base := filepath.Base(projectPath)
	// Convert to lowercase and replace spaces/special chars with hyphens
	slug := strings.ToLower(base)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove any non-alphanumeric chars except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// DetectCommands detects available commands in the project
func (d *Detector) DetectCommands(stack *config.StackInfo) config.CommandConfig {
	commands := config.CommandConfig{
		Test:           []string{},
		Lint:           []string{},
		StaticAnalysis: []string{},
		Build:          []string{},
		Start:          []string{},
	}

	// Symfony commands
	if stack.Symfony {
		commands.Test = append(commands.Test, "bin/phpunit")
		commands.Lint = append(commands.Lint, "bin/console lint:yaml config")
		commands.Lint = append(commands.Lint, "bin/console lint:twig templates")

		// Check for common QA tools
		if d.fileExists("vendor/bin/phpstan") {
			commands.StaticAnalysis = append(commands.StaticAnalysis, "vendor/bin/phpstan analyse")
		}
		if d.fileExists("vendor/bin/php-cs-fixer") {
			commands.Lint = append(commands.Lint, "vendor/bin/php-cs-fixer fix --dry-run")
		}
	}

	// Makefile commands
	if stack.Makefile {
		if targets := d.parseMakefileTargets(); len(targets) > 0 {
			for _, target := range targets {
				lower := strings.ToLower(target)
				if strings.Contains(lower, "test") {
					commands.Test = append(commands.Test, fmt.Sprintf("make %s", target))
				}
				if strings.Contains(lower, "lint") {
					commands.Lint = append(commands.Lint, fmt.Sprintf("make %s", target))
				}
				if strings.Contains(lower, "build") {
					commands.Build = append(commands.Build, fmt.Sprintf("make %s", target))
				}
				if strings.Contains(lower, "start") || strings.Contains(lower, "up") {
					commands.Start = append(commands.Start, fmt.Sprintf("make %s", target))
				}
			}
		}
	}

	// Frontend commands
	if stack.Frontend.Detected {
		commands.Build = append(commands.Build, "npm run build")
		commands.Test = append(commands.Test, "npm test")
		commands.Lint = append(commands.Lint, "npm run lint")
	}

	// Docker commands
	if stack.Docker {
		commands.Start = append(commands.Start, "docker-compose up -d")
	}

	return commands
}

// parseMakefileTargets extracts target names from Makefile
func (d *Detector) parseMakefileTargets() []string {
	data, err := os.ReadFile(filepath.Join(d.projectPath, "Makefile"))
	if err != nil {
		return nil
	}

	targets := []string{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Target lines end with ':'
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "\t") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 0 {
				target := strings.TrimSpace(parts[0])
				// Skip special targets and variables
				if target != "" && !strings.Contains(target, "=") && !strings.HasPrefix(target, ".") {
					targets = append(targets, target)
				}
			}
		}
	}

	return targets
}
