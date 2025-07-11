package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	manager, err := NewManager(tempDir)
	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, tempDir, manager.configPath)

	// Check that config directory was created
	assert.DirExists(t, tempDir)
}

func TestSaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	// Create test configuration
	config := &Config{
		TunnelName: "test-tunnel",
		CloudServer: CloudServerConfig{
			IP:      "192.168.1.100",
			Port:    22,
			User:    "testuser",
			HomeDir: "/home/testuser",
		},
		LocalServer: LocalServerConfig{
			User:        "localuser",
			ReversePort: 2222,
		},
		SSH: SSHConfig{
			PrivateKeyPath: "/path/to/private/key",
			NattedKeyPath:  "/path/to/natted/key",
		},
		Service: ServiceConfig{
			Name:          "test-service",
			AutoReconnect: true,
			RestartSec:    5,
		},
		Performance: PerformanceConfig{
			KeepAliveInterval: 30,
			KeepAliveCountMax: 3,
			ConnectTimeout:    10,
		},
		CreatedAt: time.Now(),
	}

	// Save configuration
	err = manager.SaveConfig(config)
	require.NoError(t, err)

	// Load configuration
	loadedConfig, err := manager.GetConfig("test-tunnel")
	require.NoError(t, err)

	// Verify configuration
	assert.Equal(t, config.TunnelName, loadedConfig.TunnelName)
	assert.Equal(t, config.CloudServer.IP, loadedConfig.CloudServer.IP)
	assert.Equal(t, config.CloudServer.Port, loadedConfig.CloudServer.Port)
	assert.Equal(t, config.LocalServer.ReversePort, loadedConfig.LocalServer.ReversePort)
	assert.Equal(t, config.SSH.PrivateKeyPath, loadedConfig.SSH.PrivateKeyPath)

	// Check that config file exists
	configFile := filepath.Join(tempDir, "tunnels", "test-tunnel.yaml")
	assert.FileExists(t, configFile)
}

func TestListConfigs(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	// Initially should be empty
	configs := manager.ListConfigs()
	assert.Empty(t, configs)

	// Add some configurations
	config1 := &Config{TunnelName: "tunnel1", CreatedAt: time.Now()}
	config2 := &Config{TunnelName: "tunnel2", CreatedAt: time.Now()}

	err = manager.SaveConfig(config1)
	require.NoError(t, err)

	err = manager.SaveConfig(config2)
	require.NoError(t, err)

	// List should contain both
	configs = manager.ListConfigs()
	assert.Len(t, configs, 2)
	assert.Contains(t, configs, "tunnel1")
	assert.Contains(t, configs, "tunnel2")
}

func TestDeleteConfig(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	// Create and save a configuration
	config := &Config{TunnelName: "test-tunnel", CreatedAt: time.Now()}
	err = manager.SaveConfig(config)
	require.NoError(t, err)

	// Verify it exists
	_, err = manager.GetConfig("test-tunnel")
	require.NoError(t, err)

	// Delete it
	err = manager.DeleteConfig("test-tunnel")
	require.NoError(t, err)

	// Verify it's gone
	_, err = manager.GetConfig("test-tunnel")
	assert.Error(t, err)

	// Config file should be deleted
	configFile := filepath.Join(tempDir, "tunnels", "test-tunnel.yaml")
	assert.NoFileExists(t, configFile)
}

func TestActiveConfig(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	// Initially no active config
	_, err = manager.GetActiveConfig()
	assert.Error(t, err)

	// Create and save a configuration
	config := &Config{TunnelName: "test-tunnel", CreatedAt: time.Now()}
	err = manager.SaveConfig(config)
	require.NoError(t, err)

	// Set as active
	err = manager.SetActiveConfig("test-tunnel")
	require.NoError(t, err)

	// Get active config
	activeConfig, err := manager.GetActiveConfig()
	require.NoError(t, err)
	assert.Equal(t, "test-tunnel", activeConfig.TunnelName)

	// Active config file should exist
	activeFile := filepath.Join(tempDir, "active")
	assert.FileExists(t, activeFile)

	// Content should be correct
	content, err := os.ReadFile(activeFile)
	require.NoError(t, err)
	assert.Equal(t, "test-tunnel", string(content))
}

func TestGetConfigNotFound(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	_, err = manager.GetConfig("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteConfigNotFound(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	err = manager.DeleteConfig("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
