package docker

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ContainerConfig defines configuration for creating a container
type ContainerConfig struct {
	Name        string
	Image       string
	Network     string
	Ports       map[string]string // host:container
	Env         map[string]string
	Volumes     map[string]string // volume:mountpoint
	Command     []string
	RestartPolicy string
}

// CreatePostgresContainer creates a Postgres container with pgvector
func (c *Client) CreatePostgresContainer(name, network, volume, password string, port int) error {
	// Check if container already exists
	exists, err := c.ContainerExists(name)
	if err != nil {
		return err
	}

	if exists {
		// Start it if not running
		return c.StartContainer(name)
	}

	// Create volume if it doesn't exist
	if err := c.CreateVolume(volume); err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	// Build docker run command
	args := []string{
		"run",
		"-d",
		"--name", name,
		"--network", network,
		"-p", fmt.Sprintf("%d:5432", port),
		"-e", "POSTGRES_USER=oview",
		"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
		"-e", "POSTGRES_DB=postgres",
		"-v", fmt.Sprintf("%s:/var/lib/postgresql/data", volume),
		"--restart", "unless-stopped",
		"pgvector/pgvector:pg16", // Using pgvector image which includes the extension
	}

	cmd := exec.Command("docker", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create postgres container: %w\nOutput: %s", err, output)
	}

	// Wait for Postgres to be ready
	if err := c.WaitForPostgres(name); err != nil {
		return fmt.Errorf("postgres container started but not ready: %w", err)
	}

	return nil
}

// CreateN8nContainer creates an n8n container
func (c *Client) CreateN8nContainer(name, network, volume string, port int) error {
	// Check if container already exists
	exists, err := c.ContainerExists(name)
	if err != nil {
		return err
	}

	if exists {
		// Start it if not running
		return c.StartContainer(name)
	}

	// Create volume if it doesn't exist
	if err := c.CreateVolume(volume); err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	// Build docker run command
	args := []string{
		"run",
		"-d",
		"--name", name,
		"--network", network,
		"-p", fmt.Sprintf("%d:5678", port),
		"-e", "N8N_HOST=localhost",
		"-e", fmt.Sprintf("N8N_PORT=%d", port),
		"-e", "N8N_PROTOCOL=http",
		"-e", "WEBHOOK_URL=http://localhost:" + strconv.Itoa(port) + "/",
		"-v", fmt.Sprintf("%s:/home/node/.n8n", volume),
		"--restart", "unless-stopped",
		"n8nio/n8n:latest",
	}

	cmd := exec.Command("docker", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create n8n container: %w\nOutput: %s", err, output)
	}

	return nil
}

// WaitForPostgres waits for Postgres to be ready
func (c *Client) WaitForPostgres(containerName string) error {
	// Try to connect using pg_isready
	for i := 0; i < 30; i++ {
		cmd := exec.Command("docker", "exec", containerName, "pg_isready", "-U", "oview")
		if err := cmd.Run(); err == nil {
			return nil
		}
		// Wait a bit and retry
		exec.Command("sleep", "1").Run()
	}

	return fmt.Errorf("postgres did not become ready in time")
}

// ExecInContainer executes a command inside a container
func (c *Client) ExecInContainer(containerName string, command []string) (string, error) {
	args := append([]string{"exec", containerName}, command...)
	cmd := exec.Command("docker", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("exec failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

// GetContainerIP gets the IP address of a container in a network
func (c *Client) GetContainerIP(containerName, networkName string) (string, error) {
	template := fmt.Sprintf("{{.NetworkSettings.Networks.%s.IPAddress}}", networkName)
	cmd := exec.Command("docker", "inspect", "-f", template, containerName)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container IP: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
