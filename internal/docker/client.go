package docker

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Client provides Docker operations
type Client struct {
	// Using CLI for simplicity in MVP
}

// NewClient creates a new Docker client
func NewClient() (*Client, error) {
	// Check if docker is available
	if err := exec.Command("docker", "version").Run(); err != nil {
		return nil, fmt.Errorf("docker is not available: %w (make sure Docker is installed and running)", err)
	}
	return &Client{}, nil
}

// NetworkExists checks if a Docker network exists
func (c *Client) NetworkExists(name string) (bool, error) {
	cmd := exec.Command("docker", "network", "inspect", name)
	if err := cmd.Run(); err != nil {
		// Network doesn't exist
		return false, nil
	}
	return true, nil
}

// CreateNetwork creates a Docker network
func (c *Client) CreateNetwork(name string) error {
	exists, err := c.NetworkExists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already exists, idempotent
	}

	cmd := exec.Command("docker", "network", "create", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create network: %w\nOutput: %s", err, output)
	}

	return nil
}

// ContainerExists checks if a container exists (running or stopped)
func (c *Client) ContainerExists(name string) (bool, error) {
	cmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=^%s$", name), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check container: %w", err)
	}

	return strings.TrimSpace(string(output)) == name, nil
}

// ContainerIsRunning checks if a container is currently running
func (c *Client) ContainerIsRunning(name string) (bool, error) {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=^%s$", name), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check container status: %w", err)
	}

	return strings.TrimSpace(string(output)) == name, nil
}

// StartContainer starts a container (creates if doesn't exist)
func (c *Client) StartContainer(name string) error {
	running, err := c.ContainerIsRunning(name)
	if err != nil {
		return err
	}
	if running {
		return nil // Already running
	}

	exists, err := c.ContainerExists(name)
	if err != nil {
		return err
	}

	if exists {
		// Container exists but not running, start it
		cmd := exec.Command("docker", "start", name)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to start container: %w\nOutput: %s", err, output)
		}
		return nil
	}

	return fmt.Errorf("container %s does not exist, use CreateContainer first", name)
}

// StopContainer stops a running container
func (c *Client) StopContainer(name string) error {
	running, err := c.ContainerIsRunning(name)
	if err != nil {
		return err
	}
	if !running {
		return nil // Already stopped
	}

	cmd := exec.Command("docker", "stop", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop container: %w\nOutput: %s", err, output)
	}

	return nil
}

// RemoveContainer removes a container (stops it first if running)
func (c *Client) RemoveContainer(name string) error {
	exists, err := c.ContainerExists(name)
	if err != nil {
		return err
	}
	if !exists {
		return nil // Already removed
	}

	// Stop first if running
	if err := c.StopContainer(name); err != nil {
		return err
	}

	cmd := exec.Command("docker", "rm", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove container: %w\nOutput: %s", err, output)
	}

	return nil
}

// VolumeExists checks if a Docker volume exists
func (c *Client) VolumeExists(name string) (bool, error) {
	cmd := exec.Command("docker", "volume", "inspect", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

// CreateVolume creates a Docker volume
func (c *Client) CreateVolume(name string) error {
	exists, err := c.VolumeExists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already exists
	}

	cmd := exec.Command("docker", "volume", "create", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create volume: %w\nOutput: %s", err, output)
	}

	return nil
}

// RunCommand runs a docker command and returns the output
func (c *Client) RunCommand(args ...string) (string, error) {
	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker command failed: %w\nStderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
