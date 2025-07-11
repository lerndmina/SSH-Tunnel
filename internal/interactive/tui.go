package interactive

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lerndmina/SSH-Tunnel/internal/config"
	"github.com/lerndmina/SSH-Tunnel/internal/ssh"
	"github.com/lerndmina/SSH-Tunnel/internal/tunnel"
)

// MenuItem represents a menu item
type MenuItem struct {
	title       string
	description string
	action      string
}

func (m MenuItem) Title() string       { return m.title }
func (m MenuItem) Description() string { return m.description }
func (m MenuItem) FilterValue() string { return m.title }

// State represents the current UI state
type State int

const (
	StateMainMenu State = iota
	StateNewTunnel
	StateManageTunnels
	StateSSHKeys
	StateSettings
	StateTemplates
	StateInput
	StateConfirm
)

// Model represents the TUI model
type Model struct {
	state           State
	list            list.Model
	textInput       textinput.Model
	tunnelMgr       *tunnel.Manager
	sshMgr          *ssh.KeyManager
	configMgr       *config.Manager
	currentForm     map[string]string
	formFields      []string
	formIndex       int
	message         string
	selectedTunnel  string
	confirmAction   string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render

	errorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF5555"}).
				Render
)

// NewModel creates a new TUI model
func NewModel() (*Model, error) {
	configMgr, err := config.NewManager("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	items := []list.Item{
		MenuItem{
			title:       "Create New Tunnel",
			description: "Set up a new SSH tunnel configuration",
			action:      "new_tunnel",
		},
		MenuItem{
			title:       "Manage Tunnels",
			description: "Start, stop, or view existing tunnels",
			action:      "manage_tunnels",
		},
		MenuItem{
			title:       "SSH Key Management",
			description: "Generate, install, or test SSH keys",
			action:      "ssh_keys",
		},
		MenuItem{
			title:       "Configuration Templates",
			description: "Use or create tunnel templates",
			action:      "templates",
		},
		MenuItem{
			title:       "Settings",
			description: "Configure application settings",
			action:      "settings",
		},
		MenuItem{
			title:       "Exit",
			description: "Exit the application",
			action:      "exit",
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "SSH Tunnel Manager"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()

	return &Model{
		state:       StateMainMenu,
		list:        l,
		textInput:   ti,
		tunnelMgr:   tunnel.NewManager(),
		sshMgr:      ssh.NewKeyManager(),
		configMgr:   configMgr,
		currentForm: make(map[string]string),
		formFields:  []string{"name", "remote_host", "remote_port", "user"},
	}, nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case StateMainMenu:
			return m.updateMainMenu(msg)
		case StateNewTunnel:
			return m.updateNewTunnel(msg)
		case StateManageTunnels:
			return m.updateManageTunnels(msg)
		case StateSSHKeys:
			return m.updateSSHKeys(msg)
		case StateSettings:
			return m.updateSettings(msg)
		case StateTemplates:
			return m.updateTemplates(msg)
		case StateInput:
			return m.updateInput(msg)
		case StateConfirm:
			return m.updateConfirm(msg)
		}

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 3)
		return m, nil
	}

	return m, cmd
}

// View renders the model
func (m Model) View() string {
	switch m.state {
	case StateMainMenu:
		return m.viewMainMenu()
	case StateNewTunnel:
		return m.viewNewTunnel()
	case StateManageTunnels:
		return m.viewManageTunnels()
	case StateSSHKeys:
		return m.viewSSHKeys()
	case StateSettings:
		return m.viewSettings()
	case StateTemplates:
		return m.viewTemplates()
	case StateInput:
		return m.viewInput()
	case StateConfirm:
		return m.viewConfirm()
	default:
		return m.viewMainMenu()
	}
}

// updateMainMenu handles main menu updates
func (m Model) updateMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "enter":
		if item, ok := m.list.SelectedItem().(MenuItem); ok {
			switch item.action {
			case "new_tunnel":
				m.state = StateNewTunnel
				m.currentForm = make(map[string]string)
				m.formIndex = 0
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Enter tunnel name..."
			case "manage_tunnels":
				m.state = StateManageTunnels
			case "ssh_keys":
				m.state = StateSSHKeys
			case "templates":
				m.state = StateTemplates
			case "settings":
				m.state = StateSettings
			case "exit":
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// viewMainMenu renders the main menu
func (m Model) viewMainMenu() string {
	var messageSection string
	if m.message != "" {
		if strings.Contains(m.message, "successfully") {
			messageSection = "\n" + statusMessageStyle(m.message) + "\n"
		} else {
			messageSection = "\n" + errorMessageStyle(m.message) + "\n"
		}
	}

	return fmt.Sprintf("\n%s%s\n%s\n\n%s",
		titleStyle.Render("SSH Tunnel Manager - Interactive Mode"),
		messageSection,
		m.list.View(),
		"Press 'q' to quit, '↑/↓' to navigate, 'enter' to select",
	)
}

// updateNewTunnel handles new tunnel creation
func (m Model) updateNewTunnel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateMainMenu
		m.currentForm = make(map[string]string)
		m.formIndex = 0
		m.message = "" // Clear any previous messages
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		if m.formIndex < len(m.formFields) {
			// Save current field value
			m.currentForm[m.formFields[m.formIndex]] = strings.TrimSpace(m.textInput.Value())
			m.formIndex++

			if m.formIndex < len(m.formFields) {
				// Move to next field
				m.textInput.SetValue("")
				m.textInput.Placeholder = fmt.Sprintf("Enter %s...", m.formFields[m.formIndex])
			} else {
				// All fields completed, create the tunnel
				return m.createTunnel()
			}
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// createTunnel creates a new tunnel configuration
func (m Model) createTunnel() (tea.Model, tea.Cmd) {
	// Validate required fields
	name := m.currentForm["name"]
	remoteHost := m.currentForm["remote_host"]
	remotePortStr := m.currentForm["remote_port"]
	user := m.currentForm["user"]

	if name == "" || remoteHost == "" || remotePortStr == "" || user == "" {
		m.message = "All fields are required"
		m.state = StateMainMenu
		return m, nil
	}

	remotePort, err := strconv.Atoi(remotePortStr)
	if err != nil || remotePort < 1 || remotePort > 65535 {
		m.message = "Invalid remote port number"
		m.state = StateMainMenu
		return m, nil
	}

	// Show progress message
	m.message = "Creating tunnel configuration and setting up SSH keys..."

	// Create the configuration and setup SSH keys
	return m.setupTunnelWithKeys(name, remoteHost, remotePort, user)
}

// setupTunnelWithKeys creates a tunnel configuration and sets up SSH keys
func (m Model) setupTunnelWithKeys(name, remoteHost string, remotePort int, user string) (tea.Model, tea.Cmd) {
	// Create tunnel configuration
	tunnelConfig := &config.Config{
		TunnelName: name,
		CloudServer: config.CloudServerConfig{
			IP:   remoteHost,
			Port: remotePort,
			User: user,
		},
		LocalServer: config.LocalServerConfig{
			User:        user,
			ReversePort: 2222, // Default reverse port
		},
		SSH: config.SSHConfig{
			PrivateKeyPath: fmt.Sprintf("~/.ssh/%s_key", name),
			NattedKeyPath:  fmt.Sprintf("~/.ssh/%s_key_natted", name),
		},
		Service: config.ServiceConfig{
			Name:          fmt.Sprintf("ssh-tunnel-%s", name),
			AutoReconnect: true,
			RestartSec:    30,
		},
	}

	// Save configuration first
	if err := m.configMgr.SaveConfig(tunnelConfig); err != nil {
		m.message = fmt.Sprintf("Failed to save tunnel configuration: %v", err)
		m.state = StateMainMenu
		return m, nil
	}

	// Generate SSH keys for this tunnel
	if err := m.generateTunnelKeys(tunnelConfig); err != nil {
		m.message = fmt.Sprintf("Failed to generate SSH keys: %v", err)
		m.state = StateMainMenu
		return m, nil
	}

	// Test SSH connection to the cloud server
	if err := m.testSSHConnection(tunnelConfig); err != nil {
		m.message = fmt.Sprintf("Warning: SSH connection test failed: %v", err)
	} else {
		// Try to deploy the public key to the remote server
		if err := m.deployKeyToRemote(tunnelConfig); err != nil {
			m.message = fmt.Sprintf("Tunnel created but key deployment failed: %v", err)
		} else {
			m.message = fmt.Sprintf("Tunnel '%s' created successfully with SSH keys deployed!", name)
		}
	}

	m.state = StateMainMenu
	m.currentForm = make(map[string]string)
	m.formIndex = 0
	return m, nil
}

// viewNewTunnel renders the new tunnel form
func (m Model) viewNewTunnel() string {
	progress := fmt.Sprintf("Step %d of %d", m.formIndex+1, len(m.formFields))

	var currentField string
	if m.formIndex < len(m.formFields) {
		currentField = m.formFields[m.formIndex]
	}

	var instructions string
	switch currentField {
	case "name":
		instructions = "Enter a unique name for your tunnel (e.g., 'webserver', 'database')"
	case "remote_host":
		instructions = "Enter the IP address or hostname of your remote server"
	case "remote_port":
		instructions = "Enter the SSH port of your remote server (usually 22)"
	case "user":
		instructions = "Enter the username for SSH connection"
	default:
		instructions = "Configuration complete!"
	}

	// Show previous values
	var previousValues string
	if len(m.currentForm) > 0 {
		previousValues = "\nPrevious entries:\n"
		for i := 0; i < m.formIndex && i < len(m.formFields); i++ {
			field := m.formFields[i]
			value := m.currentForm[field]
			previousValues += fmt.Sprintf("  %s: %s\n", field, value)
		}
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n%s\n\n%s\n\n%s",
		titleStyle.Render("Create New Tunnel - "+progress),
		instructions,
		previousValues,
		m.textInput.View(),
		"Press 'enter' to continue, 'esc' to go back",
		statusMessageStyle("Fill in the tunnel configuration details step by step"),
	)
}

// updateManageTunnels handles tunnel management
func (m Model) updateManageTunnels(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateMainMenu
		m.message = ""
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "s":
		// Start selected tunnel
		if m.selectedTunnel != "" {
			if err := m.tunnelMgr.Start(m.selectedTunnel); err != nil {
				m.message = fmt.Sprintf("Failed to start tunnel '%s': %v", m.selectedTunnel, err)
			} else {
				m.message = fmt.Sprintf("Tunnel '%s' started successfully", m.selectedTunnel)
			}
		}
		return m, nil
	case "t":
		// Stop selected tunnel
		if m.selectedTunnel != "" {
			if err := m.tunnelMgr.Stop(m.selectedTunnel); err != nil {
				m.message = fmt.Sprintf("Failed to stop tunnel '%s': %v", m.selectedTunnel, err)
			} else {
				m.message = fmt.Sprintf("Tunnel '%s' stopped successfully", m.selectedTunnel)
			}
		}
		return m, nil
	case "r":
		// Restart selected tunnel
		if m.selectedTunnel != "" {
			if err := m.tunnelMgr.Restart(m.selectedTunnel); err != nil {
				m.message = fmt.Sprintf("Failed to restart tunnel '%s': %v", m.selectedTunnel, err)
			} else {
				m.message = fmt.Sprintf("Tunnel '%s' restarted successfully", m.selectedTunnel)
			}
		}
		return m, nil
	case "d":
		// Delete selected tunnel
		if m.selectedTunnel != "" {
			m.state = StateConfirm
			m.confirmAction = "delete_tunnel"
			m.message = fmt.Sprintf("Are you sure you want to delete tunnel '%s'?", m.selectedTunnel)
		}
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Select tunnel by number
		configNames := m.configMgr.ListConfigs()
		index := int(msg.String()[0] - '1')
		if index >= 0 && index < len(configNames) {
			m.selectedTunnel = configNames[index]
			m.message = fmt.Sprintf("Selected tunnel: %s", m.selectedTunnel)
		}
		return m, nil
	}
	return m, nil
}

// viewManageTunnels renders tunnel management
func (m Model) viewManageTunnels() string {
	configNames := m.configMgr.ListConfigs()

	content := "Available Tunnels:\n\n"
	if len(configNames) == 0 {
		content += "No tunnels configured. Create a new tunnel first.\n"
	} else {
		for i, name := range configNames {
			status := "Stopped"
			statusColor := "31" // Red
			if tunnelStatus, err := m.tunnelMgr.GetStatus(name); err == nil && tunnelStatus != nil {
				status = tunnelStatus.Status.String()
				if status == "running" {
					statusColor = "32" // Green
				} else if status == "starting" || status == "stopping" {
					statusColor = "33" // Yellow
				}
			}
			
			selection := " "
			if name == m.selectedTunnel {
				selection = "►"
			}
			
			content += fmt.Sprintf("%s %d. %-20s [\033[%sm%s\033[0m]\n", 
				selection, i+1, name, statusColor, status)
		}
		
		content += "\n"
		if m.selectedTunnel != "" {
			content += fmt.Sprintf("Selected: %s\n\n", m.selectedTunnel)
			content += "Actions:\n"
			content += "  [s] Start tunnel\n"
			content += "  [t] Stop tunnel\n"
			content += "  [r] Restart tunnel\n"
			content += "  [d] Delete tunnel\n\n"
		}
		content += "Select tunnel by number (1-9)\n"
	}

	var messageSection string
	if m.message != "" {
		if strings.Contains(m.message, "successfully") {
			messageSection = "\n" + statusMessageStyle(m.message) + "\n"
		} else {
			messageSection = "\n" + errorMessageStyle(m.message) + "\n"
		}
	}

	return fmt.Sprintf("\n%s%s\n\n%s\n\n%s",
		titleStyle.Render("Manage Tunnels"),
		messageSection,
		content,
		"Press 'esc' to go back",
	)
}

// updateSSHKeys handles SSH key management
func (m Model) updateSSHKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateMainMenu
		m.message = ""
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "g":
		// Generate new key pair
		m.message = "Generating new SSH key pair..."
		keyPath := "~/.ssh/id_ed25519_tunnel"
		if err := m.sshMgr.GenerateKeyPair("ed25519", keyPath); err != nil {
			m.message = fmt.Sprintf("Failed to generate key pair: %v", err)
		} else {
			m.message = "SSH key pair generated successfully at " + keyPath
		}
		return m, nil
	case "t":
		// Test SSH connection (will need host/user input)
		m.message = "SSH connection testing feature coming soon..."
		return m, nil
	case "v":
		// View existing keys
		m.message = "Key listing feature coming soon..."
		return m, nil
	}
	return m, nil
}

// viewSSHKeys renders SSH key management
func (m Model) viewSSHKeys() string {
	content := "SSH Key Management:\n\n"
	content += "Available Actions:\n"
	content += "  [g] Generate new SSH key pair\n"
	content += "  [t] Test SSH connection\n"
	content += "  [v] View existing SSH keys\n\n"

	var messageSection string
	if m.message != "" {
		if strings.Contains(m.message, "successfully") {
			messageSection = statusMessageStyle(m.message) + "\n\n"
		} else {
			messageSection = errorMessageStyle(m.message) + "\n\n"
		}
	}

	return fmt.Sprintf("\n%s\n\n%s%s\n%s",
		titleStyle.Render("SSH Key Management"),
		messageSection,
		content,
		"Press 'esc' to go back",
	)
}

// updateSettings handles settings
func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateMainMenu
		m.message = ""
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// viewSettings renders settings
func (m Model) viewSettings() string {
	return fmt.Sprintf("\n%s\n\n%s\n\n%s",
		titleStyle.Render("Settings"),
		"Application settings will be implemented here...",
		"Press 'esc' to go back",
	)
}

// updateTemplates handles templates
func (m Model) updateTemplates(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateMainMenu
		m.message = ""
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// viewTemplates renders templates
func (m Model) viewTemplates() string {
	return fmt.Sprintf("\n%s\n\n%s\n\n%s",
		titleStyle.Render("Configuration Templates"),
		"Template management will be implemented here...",
		"Press 'esc' to go back",
	)
}

// updateInput handles input state
func (m Model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateMainMenu
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		// Process input
		m.state = StateMainMenu
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// viewInput renders input form
func (m Model) viewInput() string {
	return fmt.Sprintf("\n%s\n\n%s\n\n%s",
		titleStyle.Render("Input"),
		m.textInput.View(),
		"Press 'enter' to submit, 'esc' to cancel",
	)
}

// updateConfirm handles confirmation
func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.state = StateManageTunnels
		m.confirmAction = ""
		m.message = ""
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "y", "enter":
		// Process confirmation
		if m.confirmAction == "delete_tunnel" && m.selectedTunnel != "" {
			// Stop tunnel if running
			if err := m.tunnelMgr.Stop(m.selectedTunnel); err != nil {
				// Log error but continue with deletion
			}
			
			// Delete configuration
			if err := m.configMgr.DeleteConfig(m.selectedTunnel); err != nil {
				m.message = fmt.Sprintf("Failed to delete tunnel: %v", err)
			} else {
				m.message = fmt.Sprintf("Tunnel '%s' deleted successfully", m.selectedTunnel)
				m.selectedTunnel = ""
			}
		}
		m.state = StateManageTunnels
		m.confirmAction = ""
		return m, nil
	}
	return m, nil
}

// viewConfirm renders confirmation dialog
func (m Model) viewConfirm() string {
	var confirmText string
	if m.confirmAction == "delete_tunnel" && m.selectedTunnel != "" {
		confirmText = fmt.Sprintf("Are you sure you want to delete tunnel '%s'?", m.selectedTunnel)
	} else {
		confirmText = "Are you sure?"
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s",
		titleStyle.Render("Confirm Action"),
		confirmText,
		"Press 'y' to confirm, 'n' or 'esc' to cancel",
	)
}

// generateTunnelKeys generates SSH keys for a tunnel
func (m *Model) generateTunnelKeys(cfg *config.Config) error {
	// Generate primary SSH key for connecting to cloud server
	if err := m.sshMgr.GenerateKeyPair("ed25519", cfg.SSH.PrivateKeyPath); err != nil {
		return fmt.Errorf("failed to generate primary SSH key: %w", err)
	}

	// Generate natted key for reverse connections
	if err := m.sshMgr.GenerateKeyPair("ed25519", cfg.SSH.NattedKeyPath); err != nil {
		return fmt.Errorf("failed to generate natted SSH key: %w", err)
	}

	return nil
}

// testSSHConnection tests SSH connectivity to the cloud server
func (m *Model) testSSHConnection(cfg *config.Config) error {
	return m.sshMgr.TestConnection(cfg.CloudServer.IP, cfg.CloudServer.User, cfg.SSH.PrivateKeyPath, cfg.CloudServer.Port)
}

// deployKeyToRemote deploys the public key to the remote server
func (m *Model) deployKeyToRemote(cfg *config.Config) error {
	return m.sshMgr.DeployPublicKey(cfg.CloudServer.IP, cfg.CloudServer.Port, cfg.CloudServer.User, cfg.SSH.PrivateKeyPath)
}

// StartInteractiveMode starts the simple command-line interface
func StartInteractiveMode() error {
	tui, err := NewSimpleTUI()
	if err != nil {
		return fmt.Errorf("failed to create TUI: %v", err)
	}
	
	return tui.Run()
}
