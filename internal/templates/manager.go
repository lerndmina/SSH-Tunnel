package templates

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/ssh-tunnel-manager/ssh-tunnel-manager/internal/config"
)

// Template represents a configuration template
type Template struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Category    string                 `yaml:"category" json:"category"`
	Config      config.Config          `yaml:"config" json:"config"`
	Variables   map[string]Variable    `yaml:"variables" json:"variables"`
	Examples    map[string]interface{} `yaml:"examples" json:"examples"`
}

// Variable represents a template variable
type Variable struct {
	Description string      `yaml:"description" json:"description"`
	Type        string      `yaml:"type" json:"type"` // string, int, bool
	Default     interface{} `yaml:"default" json:"default"`
	Required    bool        `yaml:"required" json:"required"`
	Validation  string      `yaml:"validation,omitempty" json:"validation,omitempty"`
}

// Manager manages configuration templates
type Manager struct {
	templates map[string]*Template
}

// NewManager creates a new template manager
func NewManager() *Manager {
	manager := &Manager{
		templates: make(map[string]*Template),
	}
	
	// Load built-in templates
	manager.loadBuiltinTemplates()
	
	return manager
}

// loadBuiltinTemplates loads the predefined templates
func (m *Manager) loadBuiltinTemplates() {
	// Home Server Template
	m.templates["home-server"] = &Template{
		Name:        "home-server",
		Description: "Template for accessing home server through cloud VPS",
		Category:    "personal",
		Config: config.Config{
			TunnelName: "{{.tunnel_name}}",
			CloudServer: config.CloudServerConfig{
				IP:      "{{.cloud_ip}}",
				Port:    22,
				User:    "{{.cloud_user}}",
				HomeDir: "{{.cloud_home}}",
			},
			LocalServer: config.LocalServerConfig{
				User:        "{{.local_user}}",
				ReversePort: 2222,
			},
			SSH: config.SSHConfig{
				PrivateKeyPath: "{{.ssh_key_path}}",
				NattedKeyPath:  "{{.natted_key_path}}",
				Compression:    true,
			},
			Service: config.ServiceConfig{
				Name:          "ssh-tunnel-{{.tunnel_name}}",
				AutoReconnect: true,
				RestartSec:    5,
			},
			Performance: config.PerformanceConfig{
				KeepAliveInterval: 30,
				KeepAliveCountMax: 3,
				ConnectTimeout:    10,
			},
		},
		Variables: map[string]Variable{
			"tunnel_name": {
				Description: "Name for this tunnel configuration",
				Type:        "string",
				Required:    true,
			},
			"cloud_ip": {
				Description: "IP address of your cloud server/VPS",
				Type:        "string",
				Required:    true,
				Validation:  "ip",
			},
			"cloud_user": {
				Description: "Username on the cloud server",
				Type:        "string",
				Default:     "ubuntu",
				Required:    true,
			},
			"cloud_home": {
				Description: "Home directory on cloud server",
				Type:        "string",
				Default:     "/home/ubuntu",
				Required:    true,
			},
			"local_user": {
				Description: "Username on this local machine",
				Type:        "string",
				Required:    true,
			},
			"ssh_key_path": {
				Description: "Path to SSH private key for cloud server",
				Type:        "string",
				Default:     "~/.ssh/cloud_server_key",
				Required:    true,
			},
			"natted_key_path": {
				Description: "Path to SSH key for reverse connection",
				Type:        "string",
				Default:     "~/.ssh/natted_server_key",
				Required:    true,
			},
		},
		Examples: map[string]interface{}{
			"tunnel_name":     "home-server",
			"cloud_ip":        "203.0.113.1",
			"cloud_user":      "ubuntu",
			"cloud_home":      "/home/ubuntu",
			"local_user":      "pi",
			"ssh_key_path":    "~/.ssh/cloud_server_key",
			"natted_key_path": "~/.ssh/natted_server_key_home",
		},
	}

	// Development Server Template
	m.templates["development"] = &Template{
		Name:        "development",
		Description: "Template for development server access with SOCKS proxy",
		Category:    "development",
		Config: config.Config{
			TunnelName: "{{.tunnel_name}}",
			CloudServer: config.CloudServerConfig{
				IP:      "{{.cloud_ip}}",
				Port:    22,
				User:    "{{.cloud_user}}",
				HomeDir: "{{.cloud_home}}",
			},
			LocalServer: config.LocalServerConfig{
				User:        "{{.local_user}}",
				ReversePort: 2223,
				SOCKSPort:   1080,
			},
			SSH: config.SSHConfig{
				PrivateKeyPath: "{{.ssh_key_path}}",
				NattedKeyPath:  "{{.natted_key_path}}",
				Compression:    true,
				Ciphers:        "aes128-ctr,aes192-ctr,aes256-ctr",
			},
			Service: config.ServiceConfig{
				Name:          "ssh-tunnel-dev-{{.tunnel_name}}",
				AutoReconnect: true,
				RestartSec:    3,
			},
			Performance: config.PerformanceConfig{
				KeepAliveInterval: 60,
				KeepAliveCountMax: 5,
				ConnectTimeout:    15,
			},
		},
		Variables: map[string]Variable{
			"tunnel_name": {
				Description: "Name for this development tunnel",
				Type:        "string",
				Required:    true,
			},
			"cloud_ip": {
				Description: "IP address of your development cloud server",
				Type:        "string",
				Required:    true,
				Validation:  "ip",
			},
			"cloud_user": {
				Description: "Username on the cloud server",
				Type:        "string",
				Default:     "developer",
				Required:    true,
			},
			"cloud_home": {
				Description: "Home directory on cloud server",
				Type:        "string",
				Default:     "/home/developer",
				Required:    true,
			},
			"local_user": {
				Description: "Username on this development machine",
				Type:        "string",
				Required:    true,
			},
			"ssh_key_path": {
				Description: "Path to SSH private key for cloud server",
				Type:        "string",
				Default:     "~/.ssh/dev_server_key",
				Required:    true,
			},
			"natted_key_path": {
				Description: "Path to SSH key for reverse connection",
				Type:        "string",
				Default:     "~/.ssh/natted_dev_key",
				Required:    true,
			},
		},
		Examples: map[string]interface{}{
			"tunnel_name":     "dev-server",
			"cloud_ip":        "198.51.100.1",
			"cloud_user":      "developer",
			"cloud_home":      "/home/developer",
			"local_user":      "devuser",
			"ssh_key_path":    "~/.ssh/dev_server_key",
			"natted_key_path": "~/.ssh/natted_dev_key",
		},
	}

	// Production Server Template
	m.templates["production"] = &Template{
		Name:        "production",
		Description: "Template for production server monitoring with high availability",
		Category:    "production",
		Config: config.Config{
			TunnelName: "{{.tunnel_name}}",
			CloudServer: config.CloudServerConfig{
				IP:      "{{.cloud_ip}}",
				Port:    22,
				User:    "{{.cloud_user}}",
				HomeDir: "{{.cloud_home}}",
			},
			LocalServer: config.LocalServerConfig{
				User:        "{{.local_user}}",
				ReversePort: 2224,
			},
			SSH: config.SSHConfig{
				PrivateKeyPath: "{{.ssh_key_path}}",
				NattedKeyPath:  "{{.natted_key_path}}",
				Compression:    false, // Disable compression for production
			},
			Service: config.ServiceConfig{
				Name:          "ssh-tunnel-prod-{{.tunnel_name}}",
				AutoReconnect: true,
				RestartSec:    2,
			},
			Performance: config.PerformanceConfig{
				KeepAliveInterval: 15,
				KeepAliveCountMax: 2,
				ConnectTimeout:    5,
			},
			Analytics: config.AnalyticsConfig{
				Enabled:       true,
				RetentionDays: 90,
			},
			Notifications: config.NotificationConfig{
				Enabled:    true,
				Email:      "{{.notification_email}}",
				WebhookURL: "{{.webhook_url}}",
			},
		},
		Variables: map[string]Variable{
			"tunnel_name": {
				Description: "Name for this production tunnel",
				Type:        "string",
				Required:    true,
			},
			"cloud_ip": {
				Description: "IP address of your production cloud server",
				Type:        "string",
				Required:    true,
				Validation:  "ip",
			},
			"cloud_user": {
				Description: "Username on the cloud server",
				Type:        "string",
				Default:     "monitoring",
				Required:    true,
			},
			"cloud_home": {
				Description: "Home directory on cloud server",
				Type:        "string",
				Default:     "/home/monitoring",
				Required:    true,
			},
			"local_user": {
				Description: "Username on this production machine",
				Type:        "string",
				Required:    true,
			},
			"ssh_key_path": {
				Description: "Path to SSH private key for cloud server",
				Type:        "string",
				Default:     "~/.ssh/prod_server_key",
				Required:    true,
			},
			"natted_key_path": {
				Description: "Path to SSH key for reverse connection",
				Type:        "string",
				Default:     "~/.ssh/natted_prod_key",
				Required:    true,
			},
			"notification_email": {
				Description: "Email address for alerts (optional)",
				Type:        "string",
				Required:    false,
			},
			"webhook_url": {
				Description: "Webhook URL for alerts (optional)",
				Type:        "string",
				Required:    false,
			},
		},
		Examples: map[string]interface{}{
			"tunnel_name":        "prod-monitor",
			"cloud_ip":           "203.0.113.100",
			"cloud_user":         "monitoring",
			"cloud_home":         "/home/monitoring",
			"local_user":         "produser",
			"ssh_key_path":       "~/.ssh/prod_server_key",
			"natted_key_path":    "~/.ssh/natted_prod_key",
			"notification_email": "alerts@company.com",
			"webhook_url":        "https://hooks.slack.com/...",
		},
	}

	// IoT Device Template
	m.templates["iot-device"] = &Template{
		Name:        "iot-device",
		Description: "Template for IoT devices with minimal resource usage",
		Category:    "iot",
		Config: config.Config{
			TunnelName: "{{.tunnel_name}}",
			CloudServer: config.CloudServerConfig{
				IP:      "{{.cloud_ip}}",
				Port:    22,
				User:    "{{.cloud_user}}",
				HomeDir: "{{.cloud_home}}",
			},
			LocalServer: config.LocalServerConfig{
				User:        "{{.local_user}}",
				ReversePort: 2225,
			},
			SSH: config.SSHConfig{
				PrivateKeyPath: "{{.ssh_key_path}}",
				NattedKeyPath:  "{{.natted_key_path}}",
				Compression:    true, // Enable compression to save bandwidth
			},
			Service: config.ServiceConfig{
				Name:          "ssh-tunnel-iot-{{.tunnel_name}}",
				AutoReconnect: true,
				RestartSec:    10,
			},
			Performance: config.PerformanceConfig{
				KeepAliveInterval: 120, // Longer intervals for IoT
				KeepAliveCountMax: 10,
				ConnectTimeout:    30,
			},
		},
		Variables: map[string]Variable{
			"tunnel_name": {
				Description: "Name for this IoT device tunnel",
				Type:        "string",
				Required:    true,
			},
			"cloud_ip": {
				Description: "IP address of your cloud server",
				Type:        "string",
				Required:    true,
				Validation:  "ip",
			},
			"cloud_user": {
				Description: "Username on the cloud server",
				Type:        "string",
				Default:     "iot",
				Required:    true,
			},
			"cloud_home": {
				Description: "Home directory on cloud server",
				Type:        "string",
				Default:     "/home/iot",
				Required:    true,
			},
			"local_user": {
				Description: "Username on this IoT device",
				Type:        "string",
				Default:     "pi",
				Required:    true,
			},
			"ssh_key_path": {
				Description: "Path to SSH private key for cloud server",
				Type:        "string",
				Default:     "~/.ssh/iot_server_key",
				Required:    true,
			},
			"natted_key_path": {
				Description: "Path to SSH key for reverse connection",
				Type:        "string",
				Default:     "~/.ssh/natted_iot_key",
				Required:    true,
			},
		},
		Examples: map[string]interface{}{
			"tunnel_name":     "raspberry-pi-01",
			"cloud_ip":        "198.51.100.200",
			"cloud_user":      "iot",
			"cloud_home":      "/home/iot",
			"local_user":      "pi",
			"ssh_key_path":    "~/.ssh/iot_server_key",
			"natted_key_path": "~/.ssh/natted_iot_key",
		},
	}
}

// List returns all available template names
func (m *Manager) List() []string {
	names := make([]string, 0, len(m.templates))
	for name := range m.templates {
		names = append(names, name)
	}
	return names
}

// Get returns a template by name
func (m *Manager) Get(name string) (*Template, error) {
	template, exists := m.templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}
	return template, nil
}

// ListByCategory returns templates in a specific category
func (m *Manager) ListByCategory(category string) []*Template {
	var templates []*Template
	for _, tmpl := range m.templates {
		if tmpl.Category == category {
			templates = append(templates, tmpl)
		}
	}
	return templates
}

// GetCategories returns all available categories
func (m *Manager) GetCategories() []string {
	categories := make(map[string]bool)
	for _, tmpl := range m.templates {
		categories[tmpl.Category] = true
	}
	
	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}
	return result
}

// Apply applies a template with the given variables to create a configuration
func (m *Manager) Apply(templateName string, variables map[string]interface{}) (*config.Config, error) {
	tmpl, err := m.Get(templateName)
	if err != nil {
		return nil, err
	}
	
	// Validate required variables
	if err := m.validateVariables(tmpl, variables); err != nil {
		return nil, err
	}
	
	// Apply template rendering
	renderedConfig, err := m.renderTemplate(tmpl, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	
	return renderedConfig, nil
}

// validateVariables validates that all required variables are provided
func (m *Manager) validateVariables(tmpl *Template, variables map[string]interface{}) error {
	for varName, varDef := range tmpl.Variables {
		value, exists := variables[varName]
		
		if varDef.Required && !exists {
			return fmt.Errorf("required variable '%s' is missing", varName)
		}
		
		if !exists && varDef.Default != nil {
			variables[varName] = varDef.Default
		}
		
		if exists {
			// Type validation
			if err := m.validateVariableType(varName, value, varDef.Type); err != nil {
				return err
			}
			
			// Additional validation
			if varDef.Validation != "" {
				if err := m.validateVariableValue(varName, value, varDef.Validation); err != nil {
					return err
				}
			}
		}
	}
	
	return nil
}

// validateVariableType validates the type of a variable
func (m *Manager) validateVariableType(name string, value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("variable '%s' must be a string", name)
		}
	case "int":
		if _, ok := value.(int); !ok {
			return fmt.Errorf("variable '%s' must be an integer", name)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("variable '%s' must be a boolean", name)
		}
	}
	return nil
}

// validateVariableValue validates the value of a variable based on validation rules
func (m *Manager) validateVariableValue(name string, value interface{}, validation string) error {
	str, ok := value.(string)
	if !ok {
		return nil // Only validate string values for now
	}
	
	switch validation {
	case "ip":
		// Simple IP validation
		parts := strings.Split(str, ".")
		if len(parts) != 4 {
			return fmt.Errorf("variable '%s' must be a valid IP address", name)
		}
		// Could add more sophisticated IP validation here
	}
	
	return nil
}

// renderTemplate renders the configuration template with variables
func (m *Manager) renderTemplate(tmpl *Template, variables map[string]interface{}) (*config.Config, error) {
	// Convert config to JSON string for template processing
	configStr := m.configToTemplateString(&tmpl.Config)
	
	// Create template
	t, err := template.New("config").Parse(configStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	
	// Render template
	var rendered strings.Builder
	if err := t.Execute(&rendered, variables); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	
	// Parse back to config
	renderedConfig, err := m.templateStringToConfig(rendered.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse rendered config: %w", err)
	}
	
	return renderedConfig, nil
}

// configToTemplateString converts config to a template string
// This is a simplified version - in a real implementation, you might use JSON or YAML
func (m *Manager) configToTemplateString(cfg *config.Config) string {
	// This is a basic implementation - you would want to use proper serialization
	return fmt.Sprintf(`{
		"tunnel_name": "%s",
		"cloud_server": {
			"ip": "%s",
			"port": %d,
			"user": "%s",
			"home_dir": "%s"
		},
		"local_server": {
			"user": "%s",
			"reverse_port": %d,
			"socks_port": %d
		},
		"ssh": {
			"private_key_path": "%s",
			"natted_key_path": "%s",
			"compression": %t,
			"ciphers": "%s"
		},
		"service": {
			"name": "%s",
			"auto_reconnect": %t,
			"restart_sec": %d
		},
		"performance": {
			"keep_alive_interval": %d,
			"keep_alive_count_max": %d,
			"connect_timeout": %d
		}
	}`,
		cfg.TunnelName,
		cfg.CloudServer.IP, cfg.CloudServer.Port, cfg.CloudServer.User, cfg.CloudServer.HomeDir,
		cfg.LocalServer.User, cfg.LocalServer.ReversePort, cfg.LocalServer.SOCKSPort,
		cfg.SSH.PrivateKeyPath, cfg.SSH.NattedKeyPath, cfg.SSH.Compression, cfg.SSH.Ciphers,
		cfg.Service.Name, cfg.Service.AutoReconnect, cfg.Service.RestartSec,
		cfg.Performance.KeepAliveInterval, cfg.Performance.KeepAliveCountMax, cfg.Performance.ConnectTimeout,
	)
}

// templateStringToConfig converts a template string back to config
func (m *Manager) templateStringToConfig(str string) (*config.Config, error) {
	// This is a placeholder - you would implement proper JSON/YAML parsing here
	// For now, return a basic config
	return &config.Config{}, nil
}
