package chaos

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DockerCompose manages docker compose operations
type DockerCompose struct {
	projectName string
	composeFile string
}

// NewDockerCompose creates a new docker compose client
func NewDockerCompose() *DockerCompose {
	return &DockerCompose{
		projectName: "businessos",
		composeFile: "docker-compose.yml",
	}
}

// StopService stops a specific service
func StopService(service string) error {
	dc := NewDockerCompose()
	return dc.StopService(service)
}

// StartService starts a specific service
func StartService(service string) error {
	dc := NewDockerCompose()
	return dc.StartService(service)
}

// IsolateService isolates a service from the network (chaos injection)
func IsolateService(service string) error {
	dc := NewDockerCompose()
	return dc.IsolateService(service)
}

// RestoreService restores a service's network connectivity
func RestoreService(service string) error {
	dc := NewDockerCompose()
	return dc.RestoreService(service)
}

// StopService stops a docker compose service
func (dc *DockerCompose) StopService(service string) error {
	cmd := exec.Command("docker", "compose", "stop", service)
	cmd.Dir = "/Users/sac/chatmangpt/BusinessOS"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w\nOutput: %s", service, err, string(output))
	}
	return nil
}

// StartService starts a docker compose service
func (dc *DockerCompose) StartService(service string) error {
	cmd := exec.Command("docker", "compose", "start", service)
	cmd.Dir = "/Users/sac/chatmangpt/BusinessOS"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w\nOutput: %s", service, err, string(output))
	}

	// Wait a moment for the service to start responding
	time.Sleep(2 * time.Second)
	return nil
}

// RestartService restarts a docker compose service
func (dc *DockerCompose) RestartService(service string) error {
	cmd := exec.Command("docker", "compose", "restart", service)
	cmd.Dir = "/Users/sac/chatmangpt/BusinessOS"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %w\nOutput: %s", service, err, string(output))
	}
	return nil
}

// IsolateService disconnects a service from all networks
func (dc *DockerCompose) IsolateService(service string) error {
	// Get the container ID for the service
	containerID, err := dc.getContainerID(service)
	if err != nil {
		return fmt.Errorf("failed to get container ID: %w", err)
	}

	// Disconnect from all networks
	cmd := exec.Command("docker", "network", "disconnect", "-f", "businessos_default", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to isolate service %s: %w\nOutput: %s", service, err, string(output))
	}
	return nil
}

// RestoreService reconnects a service to the network
func (dc *DockerCompose) RestoreService(service string) error {
	// Get the container ID for the service
	containerID, err := dc.getContainerID(service)
	if err != nil {
		return fmt.Errorf("failed to get container ID: %w", err)
	}

	// Reconnect to the network
	cmd := exec.Command("docker", "network", "connect", "businessos_default", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restore service %s: %w\nOutput: %s", service, err, string(output))
	}
	return nil
}

// ServiceStatus returns the status of a service
func (dc *DockerCompose) ServiceStatus(service string) (string, error) {
	cmd := exec.Command("docker", "compose", "ps", "--services", "--filter", fmt.Sprintf("name=^%s$", service))
	cmd.Dir = "/Users/sac/chatmangpt/BusinessOS"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get service status: %w\nOutput: %s", err, string(output))
	}

	status := strings.TrimSpace(string(output))
	return status, nil
}

// getContainerID retrieves the container ID for a service
func (dc *DockerCompose) getContainerID(service string) (string, error) {
	cmd := exec.Command("docker", "compose", "ps", "-q", service)
	cmd.Dir = "/Users/sac/chatmangpt/BusinessOS"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get container ID: %w\nOutput: %s", err, string(output))
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return "", fmt.Errorf("no container found for service %s", service)
	}

	return containerID, nil
}

// ListRunningServices returns all currently running services
func (dc *DockerCompose) ListRunningServices() ([]string, error) {
	cmd := exec.Command("docker", "compose", "ps", "--services", "--filter", "status=running")
	cmd.Dir = "/Users/sac/chatmangpt/BusinessOS"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w\nOutput: %s", err, string(output))
	}

	services := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Filter out empty strings
	var result []string
	for _, svc := range services {
		if svc != "" {
			result = append(result, svc)
		}
	}

	return result, nil
}

// IsServiceRunning checks if a service is currently running
func (dc *DockerCompose) IsServiceRunning(service string) bool {
	services, err := dc.ListRunningServices()
	if err != nil {
		return false
	}

	for _, svc := range services {
		if svc == service {
			return true
		}
	}
	return false
}

// GetServiceLogs retrieves logs from a service
func (dc *DockerCompose) GetServiceLogs(service string, tail int) (string, error) {
	cmd := exec.Command("docker", "compose", "logs", "--tail", fmt.Sprintf("%d", tail), service)
	cmd.Dir = "/Users/sac/chatmangpt/BusinessOS"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get logs for %s: %w\nOutput: %s", service, err, string(output))
	}

	return string(output), nil
}

// ExecuteCommand runs a command inside a service container
func (dc *DockerCompose) ExecuteCommand(service string, command []string) (string, error) {
	args := []string{"compose", "exec", "-T", service}
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	cmd.Dir = "/Users/sac/chatmangpt/BusinessOS"

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nStderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
