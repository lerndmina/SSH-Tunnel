package tunnel

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/ssh-tunnel-manager/ssh-tunnel-manager/internal/config"
	"github.com/ssh-tunnel-manager/ssh-tunnel-manager/pkg/logger"
)

// Status represents the status of a tunnel
type Status int

const (
	StatusStopped Status = iota
	StatusStarting
	StatusRunning
	StatusStopping
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusStopped:
		return "stopped"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusStopping:
		return "stopping"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// Tunnel represents an active SSH tunnel
type Tunnel struct {
	ID              string
	Config          *config.Config
	Process         *exec.Cmd
	Status          Status
	StartTime       time.Time
	LastHealthCheck time.Time
	Error           error
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
}

// Manager manages multiple SSH tunnels
type Manager struct {
	tunnels map[string]*Tunnel
	mu      sync.RWMutex
}

// NewManager creates a new tunnel manager
func NewManager() *Manager {
	return &Manager{
		tunnels: make(map[string]*Tunnel),
	}
}

// Start starts a tunnel with the given configuration
func (m *Manager) Start(tunnelName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if tunnel is already running
	if tunnel, exists := m.tunnels[tunnelName]; exists {
		tunnel.mu.RLock()
		status := tunnel.Status
		tunnel.mu.RUnlock()

		if status == StatusRunning || status == StatusStarting {
			return fmt.Errorf("tunnel '%s' is already %s", tunnelName, status)
		}
	}

	// Get configuration
	configManager := config.GetManager()
	cfg, err := configManager.GetConfig(tunnelName)
	if err != nil {
		return fmt.Errorf("failed to get configuration for tunnel '%s': %w", tunnelName, err)
	}

	// Create tunnel context
	ctx, cancel := context.WithCancel(context.Background())

	tunnel := &Tunnel{
		ID:     tunnelName,
		Config: cfg,
		Status: StatusStarting,
		ctx:    ctx,
		cancel: cancel,
	}

	// Start the tunnel process
	if err := tunnel.start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start tunnel '%s': %w", tunnelName, err)
	}

	m.tunnels[tunnelName] = tunnel
	logger.Infof("Started tunnel '%s'", tunnelName)

	return nil
}

// Stop stops a tunnel
func (m *Manager) Stop(tunnelName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnel, exists := m.tunnels[tunnelName]
	if !exists {
		return fmt.Errorf("tunnel '%s' not found", tunnelName)
	}

	tunnel.mu.Lock()
	defer tunnel.mu.Unlock()

	if tunnel.Status == StatusStopped || tunnel.Status == StatusStopping {
		return fmt.Errorf("tunnel '%s' is already %s", tunnelName, tunnel.Status)
	}

	tunnel.Status = StatusStopping

	// Cancel context to signal shutdown
	if tunnel.cancel != nil {
		tunnel.cancel()
	}

	// Kill the process if it exists
	if tunnel.Process != nil && tunnel.Process.Process != nil {
		if err := tunnel.Process.Process.Kill(); err != nil {
			logger.Warnf("Failed to kill tunnel process: %v", err)
		}
	}

	tunnel.Status = StatusStopped
	delete(m.tunnels, tunnelName)

	logger.Infof("Stopped tunnel '%s'", tunnelName)
	return nil
}

// Restart restarts a tunnel
func (m *Manager) Restart(tunnelName string) error {
	logger.Infof("Restarting tunnel '%s'", tunnelName)

	// Stop the tunnel if it's running
	if err := m.Stop(tunnelName); err != nil {
		// Log the error but continue with start
		logger.Warnf("Error stopping tunnel during restart: %v", err)
	}

	// Wait a moment for cleanup
	time.Sleep(1 * time.Second)

	// Start the tunnel
	return m.Start(tunnelName)
}

// GetStatus returns the status of a tunnel
func (m *Manager) GetStatus(tunnelName string) (*TunnelStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tunnel, exists := m.tunnels[tunnelName]
	if !exists {
		return &TunnelStatus{
			Name:   tunnelName,
			Status: StatusStopped,
		}, nil
	}

	tunnel.mu.RLock()
	defer tunnel.mu.RUnlock()

	status := &TunnelStatus{
		Name:            tunnelName,
		Status:          tunnel.Status,
		StartTime:       tunnel.StartTime,
		LastHealthCheck: tunnel.LastHealthCheck,
		Error:           tunnel.Error,
		Uptime:          time.Since(tunnel.StartTime),
	}

	if tunnel.Process != nil && tunnel.Process.Process != nil {
		status.PID = tunnel.Process.Process.Pid
	}

	return status, nil
}

// List returns all tunnel statuses
func (m *Manager) List() ([]*TunnelStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]*TunnelStatus, 0, len(m.tunnels))

	for name := range m.tunnels {
		status, err := m.GetStatus(name)
		if err != nil {
			logger.Warnf("Failed to get status for tunnel '%s': %v", name, err)
			continue
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// HealthCheck performs a health check on a tunnel
func (m *Manager) HealthCheck(tunnelName string) error {
	m.mu.RLock()
	tunnel, exists := m.tunnels[tunnelName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("tunnel '%s' not found", tunnelName)
	}

	tunnel.mu.Lock()
	defer tunnel.mu.Unlock()

	if tunnel.Status != StatusRunning {
		return fmt.Errorf("tunnel '%s' is not running", tunnelName)
	}

	// Check if process is still alive
	if tunnel.Process == nil || tunnel.Process.Process == nil {
		tunnel.Status = StatusError
		tunnel.Error = fmt.Errorf("tunnel process not found")
		return tunnel.Error
	}

	// Check process state
	if tunnel.Process.ProcessState != nil && tunnel.Process.ProcessState.Exited() {
		tunnel.Status = StatusError
		tunnel.Error = fmt.Errorf("tunnel process has exited")
		return tunnel.Error
	}

	tunnel.LastHealthCheck = time.Now()
	return nil
}

// TunnelStatus represents the status information of a tunnel
type TunnelStatus struct {
	Name            string        `json:"name"`
	Status          Status        `json:"status"`
	StartTime       time.Time     `json:"start_time"`
	LastHealthCheck time.Time     `json:"last_health_check"`
	Uptime          time.Duration `json:"uptime"`
	PID             int           `json:"pid"`
	Error           error         `json:"error,omitempty"`
}

// start starts the SSH tunnel process
func (t *Tunnel) start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Build SSH command
	args := t.buildSSHArgs()

	logger.Debugf("Starting SSH tunnel with command: ssh %v", args)

	// Create the command
	cmd := exec.CommandContext(t.ctx, "ssh", args...)

	// Set up process attributes
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "AUTOSSH_GATETIME=0")

	// Start the process
	if err := cmd.Start(); err != nil {
		t.Status = StatusError
		t.Error = fmt.Errorf("failed to start SSH process: %w", err)
		return t.Error
	}

	t.Process = cmd
	t.Status = StatusRunning
	t.StartTime = time.Now()
	t.Error = nil

	// Monitor the process in a goroutine
	go t.monitor()

	return nil
}

// buildSSHArgs builds the SSH command arguments
func (t *Tunnel) buildSSHArgs() []string {
	cfg := t.Config
	args := []string{
		"-N", // Don't execute remote command
		"-T", // Disable pseudo-terminal allocation
	}

	// Add SSH options
	args = append(args,
		"-o", "ServerAliveInterval="+fmt.Sprintf("%d", cfg.Performance.KeepAliveInterval),
		"-o", "ServerAliveCountMax="+fmt.Sprintf("%d", cfg.Performance.KeepAliveCountMax),
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ConnectTimeout="+fmt.Sprintf("%d", cfg.Performance.ConnectTimeout),
	)

	// Add compression if enabled
	if cfg.SSH.Compression {
		args = append(args, "-o", "Compression=yes")
	}

	// Add custom ciphers if specified
	if cfg.SSH.Ciphers != "" {
		args = append(args, "-o", "Ciphers="+cfg.SSH.Ciphers)
	}

	// Add private key
	args = append(args, "-i", cfg.SSH.PrivateKeyPath)

	// Add port
	args = append(args, "-p", fmt.Sprintf("%d", cfg.CloudServer.Port))

	// Add reverse port forwarding
	reverseForward := fmt.Sprintf("%d:localhost:22", cfg.LocalServer.ReversePort)
	args = append(args, "-R", reverseForward)

	// Add SOCKS proxy if configured
	if cfg.LocalServer.SOCKSPort > 0 {
		args = append(args, "-D", fmt.Sprintf("%d", cfg.LocalServer.SOCKSPort))
	}

	// Add destination
	destination := fmt.Sprintf("%s@%s", cfg.CloudServer.User, cfg.CloudServer.IP)
	args = append(args, destination)

	return args
}

// monitor monitors the tunnel process
func (t *Tunnel) monitor() {
	defer func() {
		t.mu.Lock()
		if t.Status == StatusRunning {
			t.Status = StatusStopped
		}
		t.mu.Unlock()
	}()

	// Wait for process to complete
	err := t.Process.Wait()

	t.mu.Lock()
	if err != nil && t.ctx.Err() == nil {
		// Process exited unexpectedly
		t.Status = StatusError
		t.Error = fmt.Errorf("SSH process exited unexpectedly: %w", err)
		logger.Errorf("Tunnel '%s' process exited unexpectedly: %v", t.ID, err)
	} else if t.ctx.Err() != nil {
		// Process was cancelled
		t.Status = StatusStopped
		logger.Debugf("Tunnel '%s' process was cancelled", t.ID)
	}
	t.mu.Unlock()
}
