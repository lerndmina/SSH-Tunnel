package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

// Manager handles configuration management
type Manager struct {
	configPath   string
	configs      map[string]*Config
	activeConfig string
	mu           sync.RWMutex
}

var (
	globalManager *Manager
	once          sync.Once
)

// Config represents a tunnel configuration
type Config struct {
	TunnelName    string            `yaml:"tunnel_name" json:"tunnel_name" validate:"required"`
	CloudServer   CloudServerConfig `yaml:"cloud_server" json:"cloud_server"`
	LocalServer   LocalServerConfig `yaml:"local_server" json:"local_server"`
	SSH           SSHConfig         `yaml:"ssh" json:"ssh"`
	Service       ServiceConfig     `yaml:"service" json:"service"`
	Analytics     AnalyticsConfig   `yaml:"analytics" json:"analytics"`
	Notifications NotificationConfig `yaml:"notifications" json:"notifications"`
	Performance   PerformanceConfig `yaml:"performance" json:"performance"`
	CreatedAt     time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt     time.Time         `yaml:"updated_at" json:"updated_at"`
}

// CloudServerConfig contains cloud server connection details
type CloudServerConfig struct {
	IP       string `yaml:"ip" json:"ip" validate:"required,ip"`
	Port     int    `yaml:"port" json:"port" validate:"required,min=1,max=65535"`
	User     string `yaml:"user" json:"user" validate:"required"`
	HomeDir  string `yaml:"home_dir" json:"home_dir"`
}

// LocalServerConfig contains local server details
type LocalServerConfig struct {
	User        string `yaml:"user" json:"user" validate:"required"`
	ReversePort int    `yaml:"reverse_port" json:"reverse_port" validate:"required,min=1,max=65535"`
	SOCKSPort   int    `yaml:"socks_port,omitempty" json:"socks_port,omitempty"`
}

// SSHConfig contains SSH-related configuration
type SSHConfig struct {
	PrivateKeyPath string `yaml:"private_key_path" json:"private_key_path" validate:"required"`
	NattedKeyPath  string `yaml:"natted_key_path" json:"natted_key_path" validate:"required"`
	KnownHostsFile string `yaml:"known_hosts_file" json:"known_hosts_file"`
	Compression    bool   `yaml:"compression" json:"compression"`
	Ciphers        string `yaml:"ciphers,omitempty" json:"ciphers,omitempty"`
}

// ServiceConfig contains system service configuration
type ServiceConfig struct {
	Name          string `yaml:"name" json:"name" validate:"required"`
	AutoReconnect bool   `yaml:"auto_reconnect" json:"auto_reconnect"`
	RestartSec    int    `yaml:"restart_sec" json:"restart_sec"`
}

// AnalyticsConfig contains analytics and monitoring settings
type AnalyticsConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	DataFile     string `yaml:"data_file" json:"data_file"`
	MetricsPort  int    `yaml:"metrics_port,omitempty" json:"metrics_port,omitempty"`
	RetentionDays int   `yaml:"retention_days" json:"retention_days"`
}

// NotificationConfig contains notification settings
type NotificationConfig struct {
	Email      string `yaml:"email,omitempty" json:"email,omitempty"`
	WebhookURL string `yaml:"webhook_url,omitempty" json:"webhook_url,omitempty"`
	Enabled    bool   `yaml:"enabled" json:"enabled"`
}

// PerformanceConfig contains performance tuning settings
type PerformanceConfig struct {
	KeepAliveInterval int `yaml:"keep_alive_interval" json:"keep_alive_interval"`
	KeepAliveCountMax int `yaml:"keep_alive_count_max" json:"keep_alive_count_max"`
	ConnectTimeout    int `yaml:"connect_timeout" json:"connect_timeout"`
}

// Initialize initializes the global configuration manager
func Initialize(configPath string) error {
	var err error
	once.Do(func() {
		globalManager, err = NewManager(configPath)
	})
	return err
}

// NewManager creates a new configuration manager
func NewManager(configPath string) (*Manager, error) {
	if configPath == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, ".ssh-tunnel-manager")
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	manager := &Manager{
		configPath: configPath,
		configs:    make(map[string]*Config),
	}

	// Load existing configurations
	if err := manager.loadConfigs(); err != nil {
		return nil, fmt.Errorf("failed to load configurations: %w", err)
	}

	return manager, nil
}

// GetManager returns the global configuration manager
func GetManager() *Manager {
	return globalManager
}

// loadConfigs loads all configuration files from the config directory
func (m *Manager) loadConfigs() error {
	configsDir := filepath.Join(m.configPath, "tunnels")
	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		return nil // No configs directory yet
	}

	entries, err := os.ReadDir(configsDir)
	if err != nil {
		return fmt.Errorf("failed to read configs directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		configFile := filepath.Join(configsDir, entry.Name())
		config, err := m.loadConfig(configFile)
		if err != nil {
			// Log error but continue loading other configs
			fmt.Printf("Warning: failed to load config %s: %v\n", entry.Name(), err)
			continue
		}

		m.configs[config.TunnelName] = config
	}

	return nil
}

// loadConfig loads a single configuration file
func (m *Manager) loadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves a configuration to disk
func (m *Manager) SaveConfig(config *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config.UpdatedAt = time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = config.UpdatedAt
	}

	// Ensure tunnels directory exists
	tunnelsDir := filepath.Join(m.configPath, "tunnels")
	if err := os.MkdirAll(tunnelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create tunnels directory: %w", err)
	}

	// Write config file
	configFile := filepath.Join(tunnelsDir, config.TunnelName+".yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.configs[config.TunnelName] = config
	return nil
}

// GetConfig retrieves a configuration by name
func (m *Manager) GetConfig(name string) (*Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[name]
	if !exists {
		return nil, fmt.Errorf("configuration '%s' not found", name)
	}

	return config, nil
}

// ListConfigs returns all configuration names
func (m *Manager) ListConfigs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.configs))
	for name := range m.configs {
		names = append(names, name)
	}

	return names
}

// DeleteConfig removes a configuration
func (m *Manager) DeleteConfig(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; !exists {
		return fmt.Errorf("configuration '%s' not found", name)
	}

	// Remove config file
	configFile := filepath.Join(m.configPath, "tunnels", name+".yaml")
	if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	delete(m.configs, name)
	return nil
}

// SetActiveConfig sets the active configuration
func (m *Manager) SetActiveConfig(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; !exists {
		return fmt.Errorf("configuration '%s' not found", name)
	}

	m.activeConfig = name

	// Save active config to file
	activeConfigFile := filepath.Join(m.configPath, "active")
	return os.WriteFile(activeConfigFile, []byte(name), 0644)
}

// GetActiveConfig returns the active configuration
func (m *Manager) GetActiveConfig() (*Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeConfig == "" {
		// Try to load from file
		activeConfigFile := filepath.Join(m.configPath, "active")
		data, err := os.ReadFile(activeConfigFile)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("no active configuration set")
			}
			return nil, fmt.Errorf("failed to read active config: %w", err)
		}
		m.activeConfig = string(data)
	}

	config, exists := m.configs[m.activeConfig]
	if !exists {
		return nil, fmt.Errorf("active configuration '%s' not found", m.activeConfig)
	}

	return config, nil
}

// GetConfigPath returns the configuration directory path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}
