package service

import (
	"fmt"
	"runtime"

	"github.com/kardianos/service"
	"github.com/lerndmina/SSH-Tunnel/internal/config"
	"github.com/lerndmina/SSH-Tunnel/pkg/logger"
)

// ServiceManager handles system service operations
type ServiceManager struct {
	platform string
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		platform: runtime.GOOS,
	}
}

// ServiceConfig represents service configuration
type ServiceConfig struct {
	Name         string
	DisplayName  string
	Description  string
	Executable   string
	Arguments    []string
	Dependencies []string
	UserName     string
	WorkingDir   string
}

// ServiceStatus represents service status information
type ServiceStatus struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	PID         int    `json:"pid,omitempty"`
	StartType   string `json:"start_type"`
	Description string `json:"description"`
}

// program implements the service.Interface
type program struct {
	config         *config.Config
	tunnelName     string
	serviceManager *ServiceManager
}

func (p *program) Start(s service.Service) error {
	logger.Infof("Starting SSH tunnel service for tunnel: %s", p.tunnelName)
	go p.run()
	return nil
}

func (p *program) run() {
	// This would be implemented to run the actual SSH tunnel
	// For now, it's a placeholder
	logger.Infof("SSH tunnel service running for tunnel: %s", p.tunnelName)
}

func (p *program) Stop(s service.Service) error {
	logger.Infof("Stopping SSH tunnel service for tunnel: %s", p.tunnelName)
	return nil
}

// Install installs a service for a tunnel
func (sm *ServiceManager) Install(tunnelConfig *config.Config) error {
	serviceName := tunnelConfig.Service.Name

	// Get current executable path
	executable, err := sm.getExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create service configuration
	svcConfig := &service.Config{
		Name:        serviceName,
		DisplayName: fmt.Sprintf("SSH Tunnel: %s", tunnelConfig.TunnelName),
		Description: fmt.Sprintf("SSH Tunnel service for %s to %s@%s",
			tunnelConfig.TunnelName,
			tunnelConfig.CloudServer.User,
			tunnelConfig.CloudServer.IP),
		Executable: executable,
		Arguments:  []string{"daemon", "--tunnel", tunnelConfig.TunnelName},
		Dependencies: []string{
			"Requires=network.target",
			"After=network.target",
		},
	}

	// Set platform-specific options
	switch sm.platform {
	case "linux":
		svcConfig.Option = service.KeyValue{
			"Restart":        "always",
			"RestartSec":     fmt.Sprintf("%d", tunnelConfig.Service.RestartSec),
			"LimitNOFILE":    "65536",
			"StandardOutput": "journal",
			"StandardError":  "journal",
		}
	case "windows":
		svcConfig.Option = service.KeyValue{
			"StartType": "automatic",
		}
	case "darwin":
		svcConfig.Option = service.KeyValue{
			"KeepAlive": true,
		}
	}

	// Create program instance
	prg := &program{
		config:         tunnelConfig,
		tunnelName:     tunnelConfig.TunnelName,
		serviceManager: sm,
	}

	// Create service
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Install service
	if err := s.Install(); err != nil {
		return fmt.Errorf("failed to install service: %w", err)
	}

	logger.Infof("Service '%s' installed successfully", serviceName)
	return nil
}

// Uninstall removes a service
func (sm *ServiceManager) Uninstall(serviceName string) error {
	// Create a minimal service config for uninstall
	svcConfig := &service.Config{
		Name: serviceName,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service for uninstall: %w", err)
	}

	// Stop service first if running
	if err := s.Stop(); err != nil {
		logger.Warnf("Failed to stop service before uninstall: %v", err)
	}

	// Uninstall service
	if err := s.Uninstall(); err != nil {
		return fmt.Errorf("failed to uninstall service: %w", err)
	}

	logger.Infof("Service '%s' uninstalled successfully", serviceName)
	return nil
}

// Start starts a service
func (sm *ServiceManager) Start(serviceName string) error {
	svcConfig := &service.Config{
		Name: serviceName,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	logger.Infof("Service '%s' started successfully", serviceName)
	return nil
}

// Stop stops a service
func (sm *ServiceManager) Stop(serviceName string) error {
	svcConfig := &service.Config{
		Name: serviceName,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := s.Stop(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	logger.Infof("Service '%s' stopped successfully", serviceName)
	return nil
}

// Status gets the status of a service
func (sm *ServiceManager) Status(serviceName string) (*ServiceStatus, error) {
	svcConfig := &service.Config{
		Name: serviceName,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	status, err := s.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get service status: %w", err)
	}

	statusStr := "unknown"
	switch status {
	case service.StatusRunning:
		statusStr = "running"
	case service.StatusStopped:
		statusStr = "stopped"
	case service.StatusUnknown:
		statusStr = "unknown"
	}

	return &ServiceStatus{
		Name:        serviceName,
		Status:      statusStr,
		Description: svcConfig.Description,
	}, nil
}

// IsServiceInstalled checks if a service is installed
func (sm *ServiceManager) IsServiceInstalled(serviceName string) bool {
	status, err := sm.Status(serviceName)
	if err != nil {
		return false
	}
	return status != nil
}

// getExecutablePath returns the path to the current executable
func (sm *ServiceManager) getExecutablePath() (string, error) {
	switch sm.platform {
	case "windows":
		// On Windows, we might need to handle .exe extension
		return "ssh-tunnel.exe", nil
	default:
		return "ssh-tunnel", nil
	}
}

// GetServiceNames returns all SSH tunnel service names
func (sm *ServiceManager) GetServiceNames() ([]string, error) {
	// This would need to be implemented to query the system for
	// installed SSH tunnel services. For now, return empty list.
	return []string{}, nil
}

// EnableAutoStart enables automatic startup for a service
func (sm *ServiceManager) EnableAutoStart(serviceName string) error {
	switch sm.platform {
	case "linux":
		return sm.enableLinuxAutoStart(serviceName)
	case "windows":
		return sm.enableWindowsAutoStart(serviceName)
	case "darwin":
		return sm.enableDarwinAutoStart(serviceName)
	default:
		return fmt.Errorf("auto-start not supported on platform: %s", sm.platform)
	}
}

// DisableAutoStart disables automatic startup for a service
func (sm *ServiceManager) DisableAutoStart(serviceName string) error {
	switch sm.platform {
	case "linux":
		return sm.disableLinuxAutoStart(serviceName)
	case "windows":
		return sm.disableWindowsAutoStart(serviceName)
	case "darwin":
		return sm.disableDarwinAutoStart(serviceName)
	default:
		return fmt.Errorf("auto-start control not supported on platform: %s", sm.platform)
	}
}

// Platform-specific auto-start methods (stubs for now)
func (sm *ServiceManager) enableLinuxAutoStart(serviceName string) error {
	logger.Infof("Enabling auto-start for Linux service: %s", serviceName)
	return nil
}

func (sm *ServiceManager) disableLinuxAutoStart(serviceName string) error {
	logger.Infof("Disabling auto-start for Linux service: %s", serviceName)
	return nil
}

func (sm *ServiceManager) enableWindowsAutoStart(serviceName string) error {
	logger.Infof("Enabling auto-start for Windows service: %s", serviceName)
	return nil
}

func (sm *ServiceManager) disableWindowsAutoStart(serviceName string) error {
	logger.Infof("Disabling auto-start for Windows service: %s", serviceName)
	return nil
}

func (sm *ServiceManager) enableDarwinAutoStart(serviceName string) error {
	logger.Infof("Enabling auto-start for macOS service: %s", serviceName)
	return nil
}

func (sm *ServiceManager) disableDarwinAutoStart(serviceName string) error {
	logger.Infof("Disabling auto-start for macOS service: %s", serviceName)
	return nil
}
