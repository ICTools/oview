package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// GlobalConfig stores the global oview configuration
type GlobalConfig struct {
	// Postgres settings
	PostgresHost          string `yaml:"postgres_host"`
	PostgresPort          int    `yaml:"postgres_port"`
	PostgresUser          string `yaml:"postgres_user"`
	PostgresPassword      string `yaml:"postgres_password"`
	PostgresContainerName string `yaml:"postgres_container_name"`
	PostgresVolume        string `yaml:"postgres_volume"`

	// n8n settings
	N8nURL           string `yaml:"n8n_url"`
	N8nPort          int    `yaml:"n8n_port"`
	N8nContainerName string `yaml:"n8n_container_name"`
	N8nVolume        string `yaml:"n8n_volume"`

	// Docker network
	DockerNetworkName string `yaml:"docker_network_name"`

	mu sync.RWMutex `yaml:"-"`
}

// DefaultGlobalConfig returns a config with sensible defaults
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		PostgresHost:          "localhost",
		PostgresPort:          5432,
		PostgresUser:          "oview",
		PostgresPassword:      generatePassword(),
		PostgresContainerName: "oview-postgres",
		PostgresVolume:        "oview-postgres-data",
		N8nURL:                "http://localhost:5678",
		N8nPort:               5678,
		N8nContainerName:      "oview-n8n",
		N8nVolume:             "oview-n8n-data",
		DockerNetworkName:     "oview-net",
	}
}

// GetConfigPath returns the path to the global config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".oview", "config.yaml"), nil
}

// GetConfigDir returns the path to the .oview directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".oview"), nil
}

// LoadGlobalConfig loads the global config from ~/.oview/config.yaml
func LoadGlobalConfig() (*GlobalConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultGlobalConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves the global config to ~/.oview/config.yaml
func (c *GlobalConfig) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write atomically by writing to temp file and renaming
	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if err := os.Rename(tempPath, configPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
}

// Validate validates the global config
func (c *GlobalConfig) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.PostgresHost == "" {
		return fmt.Errorf("postgres_host is required")
	}
	if c.PostgresPort == 0 {
		return fmt.Errorf("postgres_port is required")
	}
	if c.PostgresUser == "" {
		return fmt.Errorf("postgres_user is required")
	}
	if c.PostgresPassword == "" {
		return fmt.Errorf("postgres_password is required")
	}
	if c.PostgresContainerName == "" {
		return fmt.Errorf("postgres_container_name is required")
	}
	if c.N8nPort == 0 {
		return fmt.Errorf("n8n_port is required")
	}
	if c.N8nContainerName == "" {
		return fmt.Errorf("n8n_container_name is required")
	}
	if c.DockerNetworkName == "" {
		return fmt.Errorf("docker_network_name is required")
	}

	return nil
}

// GetDSN returns the Postgres DSN for a given database
func (c *GlobalConfig) GetDSN(database string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		database,
	)
}

// generatePassword generates a random password for initial setup
func generatePassword() string {
	// For MVP, use a fixed password. In production, should use crypto/rand
	return "oview_password_change_me"
}
