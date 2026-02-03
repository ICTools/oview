package docker

import (
	"fmt"
	"net"
	"time"
)

// IsPortAvailable checks if a port is available on localhost
func IsPortAvailable(port int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		// Port is available (connection failed)
		return true
	}
	conn.Close()
	// Port is in use
	return false
}

// FindFreePort finds the next available port starting from startPort
func FindFreePort(startPort int) (int, error) {
	// Try up to 100 ports
	for i := 0; i < 100; i++ {
		port := startPort + i
		if IsPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no free port found in range %d-%d", startPort, startPort+100)
}

// EnsurePortAvailable returns the given port if available, otherwise finds a free one
func EnsurePortAvailable(preferredPort int) (int, bool, error) {
	if IsPortAvailable(preferredPort) {
		return preferredPort, false, nil
	}

	// Find alternative port
	freePort, err := FindFreePort(preferredPort + 1)
	if err != nil {
		return 0, false, err
	}

	return freePort, true, nil
}
